package syncer

import (
	"github.com/initialed85/syncer/internal/utils"
	"log"
	"reflect"
	"sync"
)

type Differ struct {
	mu                         sync.Mutex
	fileByPath, lastFileByPath map[string]*File
}

func GetDiffer() (*Differ, error) {
	s := Differ{
		fileByPath:     make(map[string]*File),
		lastFileByPath: make(map[string]*File),
	}

	return &s, nil
}

func (s *Differ) update(fileByPath map[string]*File) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if reflect.DeepEqual(fileByPath, s.fileByPath) {
		if utils.Debug {
			log.Printf("differ update ignored; fileByPath=%v, lastFileByPath=%v- no chnages", len(fileByPath), len(s.fileByPath))
		}
		return
	}

	s.lastFileByPath = CopyFileByPath(s.fileByPath)
	s.fileByPath = CopyFileByPath(fileByPath)

	if utils.Debug {
		log.Printf("differ update honoured; fileByPath=%v, lastFileByPath=%v", len(s.fileByPath), len(s.lastFileByPath))

		files, err := GetFilesFromFileByPath(s.fileByPath)
		if err != nil {
			log.Printf("warning: GetFilesFromFileByPath for %#+v caused %v", s.fileByPath, err)
			return
		}

		SortFilesInPlace(files)

		for _, file := range files {
			utils.DebugLog("differ", "state", file.Path)
		}
	}
}

func (s *Differ) diff() (map[string]*File, map[string]*File, map[string]*File) {
	s.mu.Lock()
	defer s.mu.Unlock()

	added := make(map[string]*File)
	removed := make(map[string]*File)
	modified := make(map[string]*File)

	for path, file := range s.fileByPath {
		lastFile, ok := s.lastFileByPath[path]
		if ok {
			if file.Modified == lastFile.Modified && file.Size == lastFile.Size {
				continue
			}

			modified[path] = file
			continue
		}

		added[path] = file
	}

	for lastPath, lastFile := range s.lastFileByPath {
		_, ok := s.fileByPath[lastPath]
		if ok {
			continue
		}

		removed[lastPath] = lastFile
	}

	addedFiles := 0
	removedFiles := 0
	addedFolders := 0
	modifiedFiles := 0
	removedFolders := 0
	modifiedFolders := 0

	for _, file := range added {
		utils.DebugLog("differ", "added", file.Path)

		if !file.IsDir {
			addedFiles++
			continue
		}
		addedFolders++
	}

	for _, file := range removed {
		utils.DebugLog("differ", "removed", file.Path)

		if !file.IsDir {
			removedFiles++
			continue
		}
		removedFolders++
	}

	for _, file := range modified {
		utils.DebugLog("differ", "modified", file.Path)

		if !file.IsDir {
			modifiedFiles++
			continue
		}
		modifiedFolders++
	}

	log.Printf(
		"files: %v added, %v removed, %v modified; folders: %v added, %v removed, %v modified",
		addedFiles,
		removedFiles,
		modifiedFiles,
		addedFolders,
		removedFolders,
		modifiedFolders,
	)

	return added, removed, modified
}
