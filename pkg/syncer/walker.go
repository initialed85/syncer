package syncer

import (
	"fmt"
	"github.com/MichaelTJones/walk"
	"github.com/kalafut/imohash"
	ignore "github.com/sabhiram/go-gitignore"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
)

func GetFilesAndGitIgnoreByPath(path string) ([]*File, map[string]*ignore.GitIgnore, error) {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	files := make([]*File, 0)
	gitIgnoreByPath := make(map[string]*ignore.GitIgnore)

	// note: walk.Walk makes uses goroutines to walk as fast as possible (so it's heavy)
	err := walk.Walk(
		path,
		func(path string, info os.FileInfo, walkErr error) error {
			wg.Add(1)
			defer wg.Done()

			if walkErr != nil {
				return walkErr
			}

			// apply the folder regex while we're here for efficiency
			if folderIgnoreExp.MatchString(path) {
				return nil
			}

			// apply the file regex while we're here for efficiency
			if fileIgnoreExp.MatchString(path) {
				return nil
			}

			file, walkErr := GetFileWithInfo(path, info)
			if walkErr != nil {
				return walkErr
			}

			// build GitIgnores while we're here for efficiency; note: compilation is done in goroutines so walk.Walk can go fast
			if !file.IsDir && file.Name == ".gitignore" {
				wg.Add(1)
				go func(gitIgnoreFile *File) {
					defer wg.Done()

					gitIgnore, compileErr := ignore.CompileIgnoreFile(gitIgnoreFile.Path)
					if compileErr != nil {
						log.Printf("warning: attempt to parse %v caused %v", gitIgnoreFile.Path, walkErr)
						return
					}

					mu.Lock()
					gitIgnoreByPath[gitIgnoreFile.ParentPath] = gitIgnore
					mu.Unlock()
				}(file)
				runtime.Gosched()
			}

			_ = imohash.SumFile
			// TODO
			// if !(file.IsDir || file.IsSymlink) {
			// 	wg.Add(1)
			// 	go func(sumFile *File) {
			// 		defer wg.Done()
			//
			// 		Sum, err := imohash.SumFile(file.Path)
			// 		if err != nil {
			// 			log.Printf("warning: attempt to imohash.SumFile %v returned %v", sumFile.Path, err)
			// 		}
			//
			// 		file.Sum = Sum
			// 	}(file)
			// 	runtime.Gosched()
			// }

			mu.Lock()
			files = append(files, file)
			mu.Unlock()

			return nil
		},
	)

	runtime.Gosched()

	if err != nil && err != walk.SkipDir {
		return nil, nil, err
	}

	wg.Wait()

	return files, gitIgnoreByPath, nil
}

func FilterFiles(files []*File, gitIgnoreByPath map[string]*ignore.GitIgnore) ([]*File, error) {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	gitIgnoreFilteredFiles := make([]*File, 0)

	for _, file := range files {
		wg.Add(1)

		go func(f *File) {
			defer wg.Done()

			gitIgnored := false

			for gitIgnorePath, gitIgnore := range gitIgnoreByPath {
				if !strings.HasPrefix(f.Path, gitIgnorePath) {
					continue
				}

				if !gitIgnore.MatchesPath(f.Path) {
					continue
				}

				gitIgnored = true
				break
			}

			if gitIgnored {
				return
			}

			mu.Lock()
			gitIgnoreFilteredFiles = append(gitIgnoreFilteredFiles, f)
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	return gitIgnoreFilteredFiles, nil
}

func FilterFolders(files []*File) ([]*File, error) {
	folders := make([]*File, 0)

	for _, file := range files {
		if !file.IsDir {
			continue
		}

		folders = append(folders, file)
	}

	return folders, nil
}

func GetFileByPathFromFiles(files []*File) (map[string]*File, error) {
	fileByPath := make(map[string]*File)

	for i, file := range files {
		duplicateFile, ok := fileByPath[file.Path]
		if ok {
			return nil, fmt.Errorf("duplicate i=%v, file=%v, duplicateFile=%v", i, file, duplicateFile)
		}

		fileByPath[file.Path] = file
	}

	return fileByPath, nil
}

func GetFileByPathAndFolderByPathAndGitIgnoreByPathForPath(path string) (map[string]*File, map[string]*File, map[string]*ignore.GitIgnore, error) {
	allFiles, gitIgnoreByPath, err := GetFilesAndGitIgnoreByPath(path)
	if err != nil {
		return nil, nil, nil, err
	}

	files, err := FilterFiles(allFiles, gitIgnoreByPath)
	if err != nil {
		return nil, nil, nil, err
	}

	folders, err := FilterFolders(files)
	if err != nil {
		return nil, nil, nil, err
	}

	fileByPath, err := GetFileByPathFromFiles(files)
	if err != nil {
		return nil, nil, nil, err
	}

	folderByPath, err := GetFileByPathFromFiles(folders)
	if err != nil {
		return nil, nil, nil, err
	}

	return fileByPath, folderByPath, gitIgnoreByPath, nil
}

func GetFilesFromFileByPath(fileByPath map[string]*File) ([]*File, error) {
	files := make([]*File, 0)

	for _, file := range fileByPath {
		files = append(files, file)
	}

	return files, nil
}
