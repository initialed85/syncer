package syncer

import (
	"fmt"
	"github.com/initialed85/syncer/internal/utils"
	"github.com/rjeczalik/notify"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

type Watcher struct {
	mu                     sync.Mutex
	fsEvents               chan notify.EventInfo
	bufferedFsEvents       []notify.EventInfo
	lastFsEvent            time.Time
	errors                 chan error
	started, stop, stopped chan bool
	handler                *Handler
	ticker                 *time.Ticker
	watching               map[string]*File
	path                   string
	rate, debounce         time.Duration
}

func GetWatcher(path string, rate time.Duration, debounce time.Duration, handler *Handler) (*Watcher, error) {
	w := Watcher{
		errors:   make(chan error),
		started:  make(chan bool),
		stop:     make(chan bool),
		stopped:  make(chan bool),
		watching: make(map[string]*File, 0),
		path:     path,
		rate:     rate,
		debounce: debounce,
		handler:  handler,
	}

	handler.setWatcher(&w)

	err := w.start()
	if err != nil {
		return nil, err
	}

	return &w, nil
}

func (w *Watcher) bufferFsEvent(fsEvent notify.EventInfo) {
	utils.DebugLog("raw_fs_event", fsEvent.Event().String(), fsEvent.Path())

	w.mu.Lock()
	defer w.mu.Unlock()

	w.bufferedFsEvents = append(w.bufferedFsEvents, fsEvent)
	w.lastFsEvent = time.Now()
}

func (w *Watcher) handleFsEvent(fsEvent notify.EventInfo) {
	utils.DebugLog("fs_event", fsEvent.Event().String(), fsEvent.Path())

	event := Event{
		Operation: Unknown,
		Name:      fsEvent.Path(),
	}

	if fsEvent.Event()&notify.Create == notify.Create {
		event.Operation = Created
	} else if fsEvent.Event()&notify.Remove == notify.Remove {
		event.Operation = Deleted
	} else if fsEvent.Event()&notify.Write == notify.Write {
		event.Operation = Modified
	} else if fsEvent.Event()&notify.Rename == notify.Rename {
		event.Operation = Moved
	}

	w.mu.Lock()
	h := w.handler
	w.mu.Unlock()

	if h == nil {
		log.Printf("warning: handler is nil, cannot handle %#+v", event)
		return
	}

	err := h.handleEvent(&event)
	if err != nil {
		log.Printf("warning: w.handler.handleEvent with %v caused %v", event, err)
		return
	}
}

func (w *Watcher) handleBufferedFsEvents() {
	w.mu.Lock()
	lastFsEvent := w.lastFsEvent
	w.mu.Unlock()

	if lastFsEvent.Equal(time.Time{}) {
		return
	}

	if time.Now().Before(lastFsEvent.Add(w.debounce)) {
		return
	}

	w.mu.Lock()
	bufferedFsEvents := w.bufferedFsEvents
	w.bufferedFsEvents = nil
	w.mu.Unlock()

	if len(bufferedFsEvents) == 0 {
		return
	}

	// TODO: only works for macOS
	// sort the events by ID so the order is as sane as possible
	sort.SliceStable(
		bufferedFsEvents,
		func(i, j int) bool {
			fsEventA := bufferedFsEvents[i]
			fsEventB := bufferedFsEvents[j]

			rawSysA := fsEventA.Sys()
			rawSysB := fsEventB.Sys()

			if rawSysA == nil || rawSysB == nil { // will be nil for OSes other than macOS
				return false
			}

			sysA, ok := rawSysA.(*notify.FSEvent)
			if !ok {
				return false
			}

			sysB, ok := rawSysB.(*notify.FSEvent)
			if !ok {
				return false
			}

			return sysA.ID < sysB.ID
		},
	)

	for _, fsEvent := range bufferedFsEvents {
		w.handleFsEvent(fsEvent)
	}

	w.mu.Lock()
	h := w.handler
	w.mu.Unlock()

	h.updateDiffer()
}

func (w *Watcher) run() {
	var err error

	w.ticker = time.NewTicker(w.rate)

	defer func() {
		_ = w.ticker.Stop
	}()

	w.fsEvents = make(chan notify.EventInfo, 65536) // should be more than enough to ensure we don't block the OS
	w.bufferedFsEvents = make([]notify.EventInfo, 0)

	path := fmt.Sprintf(
		"%v/...",
		strings.TrimRight(w.path, "/"),
	)

	err = notify.Watch(path, w.fsEvents, notify.Create, notify.Remove, notify.Write, notify.Rename)

	w.errors <- err

	if err != nil {
		return
	}

	defer func() {
		notify.Stop(w.fsEvents)
		w.stopped <- true
	}()

	w.started <- true

	log.Printf("looping until stopped...")

	for {
		select {

		case <-w.stop:
			return

		case fsEvent := <-w.fsEvents:
			w.bufferFsEvent(fsEvent)

		case <-w.ticker.C:
			w.handleBufferedFsEvents()
		}
	}
}

func (w *Watcher) start() error {
	log.Printf("starting...")
	go w.run()

	err := <-w.errors
	if err != nil {
		return err
	}

	<-w.started
	log.Printf("started.")

	return nil
}

func (w *Watcher) Close() {
	log.Printf("stopping...")
	w.stop <- true
	<-w.stopped
	log.Printf("stopped.")
}
