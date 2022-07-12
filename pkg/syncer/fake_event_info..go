package syncer

import "github.com/rjeczalik/notify"

type FakeEventInfo struct {
	event notify.Event
	path  string
}

func (f FakeEventInfo) Event() notify.Event {
	return f.event
}

func (f FakeEventInfo) Path() string {
	return f.path
}

func (f FakeEventInfo) Sys() interface{} {
	return nil
}
