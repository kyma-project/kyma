package file

import (
	"context"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

func NotifyModification(ctx context.Context, path string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "unable to  create file watcher")
	}
	defer watcher.Close()

	if err = watcher.Add(path); err != nil {
		return errors.Wrap(err, "unable to add file to watcher")
	}

	return selectModification(ctx, watcher)
}

func selectModification(ctx context.Context, watcher *fsnotify.Watcher) error {
	select {
	case <-ctx.Done():
		return context.Canceled
	case <-watcher.Events:
		return nil
	case watcherError := <-watcher.Errors:
		return watcherError
	}
}
