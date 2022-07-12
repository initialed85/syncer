package syncer

import "sort"

func SortFilesInPlace(files []*File) {
	sort.SliceStable(
		files,
		func(i, j int) bool {
			return files[i].Path < files[j].Path
		},
	)
}

func CopyFileByPath(fileByPath map[string]*File) map[string]*File {
	copiedFileByPath := make(map[string]*File)

	for path, file := range fileByPath {
		copiedFileByPath[path] = &*file
	}

	return copiedFileByPath
}
