package syncer

import (
	"fmt"
	"github.com/initialed85/syncer/internal/utils"
	ignore "github.com/sabhiram/go-gitignore"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Handler struct {
	mu              sync.Mutex
	fileByPath      map[string]*File
	gitIgnoreByPath map[string]*ignore.GitIgnore
	watcher         *Watcher
	path            string
	differ          *Differ
}

func GetHandler(path string, differ *Differ) (*Handler, error) {
	h := Handler{
		fileByPath:      make(map[string]*File),
		gitIgnoreByPath: make(map[string]*ignore.GitIgnore),
		path:            path,
		differ:          differ,
	}

	return &h, nil
}

func (h *Handler) handleFileByPath(operation Operation, fileByPath map[string]*File) error {
	commonPath := ""
	for path := range fileByPath {
		if commonPath != "" && len(path) > len(commonPath) {
			continue
		}

		commonPath = path
	}

	if operation == Created {
		h.mu.Lock()

		for path, file := range fileByPath {
			h.fileByPath[path] = file
		}

		h.mu.Unlock()
	}

	if operation == Modified {
		h.mu.Lock()

		for path, file := range fileByPath {
			h.fileByPath[path] = file
		}

		pathsToDelete := make([]string, 0)

		if commonPath != "" {
			for existingPath := range h.fileByPath {
				if !strings.HasPrefix(existingPath, commonPath) {
					continue
				}

				_, ok := fileByPath[existingPath]
				if ok {
					continue
				}

				pathsToDelete = append(pathsToDelete, existingPath)
			}
		}

		for _, path := range pathsToDelete {
			delete(h.fileByPath, path)
		}
		h.mu.Unlock()
	}

	if operation == Deleted {
		h.mu.Lock()

		pathsToDelete := make([]string, 0)

		for path := range fileByPath {
			for existingPath := range h.fileByPath {
				if !strings.HasPrefix(existingPath, path) {
					continue
				}

				pathsToDelete = append(pathsToDelete, existingPath)
			}
		}

		for _, path := range pathsToDelete {
			delete(h.fileByPath, path)
		}

		h.mu.Unlock()
	}

	return nil
}

func (h *Handler) handleGitIgnoreByPath(operation Operation, gitIgnoreByPath map[string]*ignore.GitIgnore) error {
	if operation == Created || operation == Modified {
		h.mu.Lock()
		for path, gitIgnore := range gitIgnoreByPath {
			h.gitIgnoreByPath[path] = gitIgnore
		}
		h.mu.Unlock()
	}

	if operation == Deleted {
		h.mu.Lock()
		pathsToDelete := make([]string, 0)

		for path := range gitIgnoreByPath {
			for existingPath := range h.gitIgnoreByPath {
				if !strings.HasPrefix(existingPath, path) {
					continue
				}

				pathsToDelete = append(pathsToDelete, existingPath)
			}
		}

		for _, path := range pathsToDelete {
			delete(h.fileByPath, path)
		}
		h.mu.Unlock()
	}

	return nil
}

func (h *Handler) add(path string) {
	before := time.Now()

	fileByPath, folderByPath, gitIgnoreByPath, err := GetFileByPathAndFolderByPathAndGitIgnoreByPathForPath(path)
	if err != nil { // this can occur if things are quickly added then deleted- not much we can do about it
		return
	}

	err = h.handleGitIgnoreByPath(Created, gitIgnoreByPath)
	if err != nil {
		log.Printf(
			"warning: handleGitIgnoreByPath for %#+v caused %v",
			gitIgnoreByPath,
			err,
		)
		return
	}

	err = h.handleFileByPath(Created, fileByPath)
	if err != nil {
		log.Printf(
			"warning: handleFileByPath for %#+v caused %v",
			fileByPath,
			err,
		)
		return
	}

	after := time.Now()

	if len(fileByPath) <= 1 && len(folderByPath) == 0 {
		return
	}

	if utils.Debug {
		log.Printf(
			"walked %v to add %v files, %v folders and %v .gitignores in %v",
			path, len(fileByPath), len(folderByPath), len(gitIgnoreByPath), after.Sub(before),
		)
	}
}

func (h *Handler) remove(path string) {
	before := time.Now()

	var err error
	var fileByPath map[string]*File
	var folderByPath map[string]*File
	var gitIgnoreByPath map[string]*ignore.GitIgnore

	madeAssumptions := false

	fileByPath, folderByPath, gitIgnoreByPath, err = GetFileByPathAndFolderByPathAndGitIgnoreByPathForPath(path)
	if err != nil { // path doesn't exist (possible); so assume it's a folder
		madeAssumptions = true

		fileByPath = make(map[string]*File)
		fileByPath[path] = GetFolderWithFakeInfo(path)

		files, err := GetFilesFromFileByPath(fileByPath)
		if err != nil {
			log.Printf("warning: GetFilesFromFileByPath for %#+v caused %v", fileByPath, err)
			return
		}

		h.mu.Lock()
		g := h.gitIgnoreByPath
		h.mu.Unlock()

		files, err = FilterFiles(files, g)
		if err != nil {
			log.Printf("warning: FilterFiles for %#+v and %#+v caused %v", fileByPath, g, err)
			return
		}

		fileByPath, err = GetFileByPathFromFiles(files)
		if err != nil {
			log.Printf("warning: GetFileByPathFromFiles for %#+v caused %v", files, err)
			return
		}

		folderByPath = fileByPath
		gitIgnoreByPath = make(map[string]*ignore.GitIgnore)
	}

	err = h.handleGitIgnoreByPath(Deleted, gitIgnoreByPath)
	if err != nil {
		log.Printf(
			"warning: handleGitIgnoreByPath for %#+v caused %v",
			gitIgnoreByPath,
			err,
		)
		return
	}

	err = h.handleFileByPath(Deleted, fileByPath)
	if err != nil {
		log.Printf(
			"warning: handleFileByPath for %#+v caused %v",
			fileByPath,
			err,
		)
		return
	}

	after := time.Now()

	if len(fileByPath) <= 1 && len(folderByPath) == 0 || len(fileByPath) == 1 && len(folderByPath) == 1 && madeAssumptions {
		return
	}

	if utils.Debug {
		log.Printf(
			"walked %v to remove %v files, %v folders and %v .gitignores in %v",
			path, len(fileByPath), len(folderByPath), len(gitIgnoreByPath), after.Sub(before),
		)
	}
}

func (h *Handler) update(path string) {
	before := time.Now()

	fileByPath, folderByPath, gitIgnoreByPath, err := GetFileByPathAndFolderByPathAndGitIgnoreByPathForPath(path)
	if err != nil {
		return
	}

	err = h.handleGitIgnoreByPath(Modified, gitIgnoreByPath)
	if err != nil {
		log.Printf(
			"warning: handleGitIgnoreByPath for %#+v caused %v",
			gitIgnoreByPath,
			err,
		)
		return
	}

	err = h.handleFileByPath(Modified, fileByPath)
	if err != nil {
		log.Printf(
			"warning: handleFileByPath for %#+v caused %v",
			fileByPath,
			err,
		)
		return
	}

	after := time.Now()

	if len(fileByPath) <= 1 && len(folderByPath) == 0 {
		return
	}

	if utils.Debug {
		log.Printf(
			"walked %v to update %v files, %v folders and %v .gitignores in %v",
			path, len(fileByPath), len(folderByPath), len(gitIgnoreByPath), after.Sub(before),
		)
	}
}

func (h *Handler) handleEvent(event *Event) error {
	path, err := filepath.Abs(event.Name)
	if err != nil {
		return err
	}

	parentPath, _ := filepath.Split(path)

	event.Path = path
	event.ParentPath = strings.TrimRight(parentPath, "/")

	h.mu.Lock()
	w := h.watcher
	h.mu.Unlock()

	if w == nil {
		return fmt.Errorf("watcher is nil, cannot handle %#+v", event)
	}

	utils.DebugLog("event", string(event.Operation), event.Path)

	if event.Operation == Created {
		h.add(event.Path)
	}

	if event.Operation == Deleted {
		h.remove(event.Path)
	}

	if event.Operation == Modified || event.Operation == Moved && strings.HasPrefix(event.ParentPath, h.path) {
		h.remove(event.ParentPath)
		h.add(event.ParentPath)
	}

	return nil
}

func (h *Handler) updateDiffer() {
	h.mu.Lock()
	fileByPath := h.fileByPath
	h.mu.Unlock()

	h.differ.update(fileByPath)
	_, _, _ = h.differ.diff()
}

func (h *Handler) setWatcher(watcher *Watcher) {
	h.mu.Lock()
	h.watcher = watcher
	h.mu.Unlock()

	log.Printf("walking %v to build base state", h.path)
	h.add(h.path)

	h.updateDiffer()
}
