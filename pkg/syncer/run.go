package syncer

import (
	"time"
)

func Run(
	localPath string,
	rate time.Duration,
	debounce time.Duration,
) (func(), error) {
	differ, err := GetDiffer()
	if err != nil {
		return nil, err
	}

	handler, err := GetHandler(localPath, differ)
	if err != nil {
		return nil, err
	}

	watcher, err := GetWatcher(localPath, rate, debounce, handler)
	if err != nil {
		return nil, err
	}

	return func() {
		watcher.Close()
	}, nil
}
