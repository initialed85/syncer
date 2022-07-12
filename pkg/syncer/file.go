package syncer

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type File struct {
	Name       string
	Path       string
	ParentPath string
	HasInfo    bool
	Size       int64
	Modified   time.Time
	Mode       fs.FileMode
	IsDir      bool
	IsSymlink  bool
	Sum        [16]byte
}

func GetFileWithInfo(path string, info fs.FileInfo) (*File, error) {
	var err error

	if info == nil {
		info, err = os.Stat(path)
		if err != nil {
			return nil, err
		}
	}

	isDir := info.IsDir()
	isSymlink := info.Mode()&os.ModeSymlink == os.ModeSymlink

	parentPath, _ := filepath.Split(path)

	file := File{
		Name:       info.Name(),
		Path:       path,
		ParentPath: strings.TrimRight(parentPath, "/"),
		HasInfo:    true,
		Size:       info.Size(),
		Modified:   info.ModTime(),
		Mode:       info.Mode(),
		IsDir:      isDir,
		IsSymlink:  isSymlink,
	}

	return &file, nil
}

func GetFileWithoutInfo(path string) *File {
	parentPath, name := filepath.Split(path)

	return &File{
		Name:       name,
		Path:       path,
		ParentPath: strings.TrimRight(parentPath, "/"),
		HasInfo:    false,
		Size:       0,
		Modified:   time.Now(),
		Mode:       fs.ModeIrregular,
		IsDir:      false,
		IsSymlink:  false,
		Sum:        [16]byte{},
	}
}

func GetFileWithOptionalInfo(path string) *File {
	file, err := GetFileWithInfo(path, nil)
	if err != nil {
		file = GetFileWithoutInfo(path)
	}

	return file
}

func GetFolderWithFakeInfo(path string) *File {
	_, name := filepath.Split(path)

	file, err := GetFileWithInfo(path, &fakeFileInfo{
		name:  name,
		isDir: true,
	})
	if err != nil {
		file = GetFileWithoutInfo(path)
	}

	return file
}
