package syncer

import (
	"io/fs"
	"time"
)

type fakeFileInfo struct {
	name  string
	isDir bool
}

func (f *fakeFileInfo) Name() string {
	return f.name
}

func (f *fakeFileInfo) Size() int64 {
	return 0
}

func (f *fakeFileInfo) Mode() fs.FileMode {
	return fs.ModeIrregular
}

func (f *fakeFileInfo) ModTime() time.Time {
	return time.Now()
}

func (f *fakeFileInfo) IsDir() bool {
	return f.isDir
}

func (f *fakeFileInfo) Sys() any {
	return nil
}
