package main

import (
	"github.com/initialed85/syncer/internal/args"
	"github.com/initialed85/syncer/pkg/syncer"
	"log"
	"time"
)

func main() {
	runArgs := args.ValidateArgs(args.ParseArgs())

	var before, after time.Time
	_ = before
	_ = after

	before = time.Now()
	allFiles, gitIgnoreByPath, err := syncer.GetFilesAndGitIgnoreByPath(runArgs.LocalPath)
	if err != nil {
		log.Fatal(err)
	}
	after = time.Now()
	log.Printf("allFiles=%v, gitIgnoreByPath=%v, duration=%v", len(allFiles), len(gitIgnoreByPath), after.Sub(before))

	before = time.Now()
	allFolders, err := syncer.FilterFolders(allFiles)
	if err != nil {
		log.Fatal(err)
	}
	after = time.Now()
	log.Printf("allFolders=%v, gitIgnoreByPath=%v, duration=%v", len(allFolders), len(gitIgnoreByPath), after.Sub(before))

	before = time.Now()
	files, err := syncer.FilterFiles(allFiles, gitIgnoreByPath)
	if err != nil {
		log.Fatal(err)
	}
	after = time.Now()
	log.Printf("files=%v, duration=%v", len(files), after.Sub(before))

	before = time.Now()
	folders, err := syncer.FilterFolders(files)
	if err != nil {
		log.Fatal(err)
	}
	after = time.Now()
	log.Printf("folders=%v, duration=%v", len(folders), after.Sub(before))

	before = time.Now()
	fileByPath, err := syncer.GetFileByPathFromFiles(files)
	if err != nil {
		log.Fatal(err)
	}
	after = time.Now()
	log.Printf("fileByPath=%v, duration=%v", len(fileByPath), after.Sub(before))

	before = time.Now()
	folderByPath, err := syncer.GetFileByPathFromFiles(folders)
	if err != nil {
		log.Fatal(err)
	}
	after = time.Now()
	log.Printf("folderByPath=%v, duration=%v", len(folderByPath), after.Sub(before))
}
