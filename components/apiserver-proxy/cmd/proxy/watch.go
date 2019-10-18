package main

import (
	"context"
	"path/filepath"
	"time"

	"github.com/howeyc/fsnotify"

	"istio.io/pkg/log"
)

const (
	// defaultMinDelay is the minimum amount of time between delivery of two successive events via updateFunc.
	defaultMinDelay = 10 * time.Second
)

// Watcher provides notifications about changes files mounted inside kubernetes Pod, like Secrets or ConfigMaps
// Note: Because file watches breaks in kubernetes when mounts are updated, we watch for directories instead.
// This happens because files mounted to a Pod are actually symbolic links pointing to "real" files.
// On updating mounted files, kubernetes deletes the existing file, which sends a DELETE file event and breaks the watch
type Watcher interface {
	// Run start the watcher loop (blocking call)
	// context is used to terminate the loop
	Run(context.Context)
}

type watcher struct {
	filePaths       []string //a single watcher can react to changes to many files
	minDelaySeconds uint8
	notifyFunc      func()
}

// NewWatcher creates a new watcher instance
// filePaths parameter is a list of file paths to watch
// notifyFunc is a function that is invoked after watcher detects changes to monitored files.
func NewWatcher(filePaths []string, minDelaySeconds uint8, notifyFunc func()) Watcher {
	return &watcher{
		filePaths:       filePaths,
		minDelaySeconds: minDelaySeconds,
		notifyFunc:      notifyFunc,
	}
}

//Run implements Watcher interface
func (w *watcher) Run(ctx context.Context) {
	watchFileEventsFunc := func(fEventChan <-chan *fsnotify.FileEvent) {
		watchFileEvents(ctx, fEventChan, defaultMinDelay, w.notifyFunc)
	}

	dirs := uniqeDirNames(w.filePaths)

	// monitor files
	go func() {
		watchForDirs(dirs, watchFileEventsFunc)
	}()

	<-ctx.Done()
	log.Info("Watcher has successfully terminated")
}

// watchFileEvents watches for changes on a channel and notifies via notifyFn().
// The function batches changes so that related changes are processed together.
// The function ensures that notifyFn() is called no more than one time per minDelay.
// The function does not return until the the context is canceled.
func watchFileEvents(ctx context.Context, wch <-chan *fsnotify.FileEvent, minDelay time.Duration, notifyFunc func()) {
	// timer and channel for managing minDelay.
	var timeChan <-chan time.Time
	var timer *time.Timer

	for {
		select {
		case ev := <-wch:
			log.Infof("watchFileEvents: %s", ev.String())
			if timer != nil {
				continue
			}
			// create new timer
			timer = time.NewTimer(minDelay)
			timeChan = timer.C
		case <-timeChan:
			// reset timer
			timeChan = nil
			timer.Stop()
			timer = nil

			log.Info("watchFileEvents: notifying")
			notifyFunc()
		case <-ctx.Done():
			log.Infof("watchFileEvents for has successfully terminated")
			return
		}
	}
}

// watchForDirs configures a watch for every directory path in dirs.
// It then invokes provided watchFunc with configured Watcher.
// This function is expected to be blocking so it should be run as a goroutine.
func watchForDirs(dirs []string, watchFunc func(fEventChan <-chan *fsnotify.FileEvent)) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		log.Warnf("failed to create a watcher for certificate files: %v", err)
		return
	}
	defer func() {
		if err := fw.Close(); err != nil {
			log.Warnf("closing watcher encounters an error %v", err)
		}
	}()

	for _, dir := range dirs {
		if err := fw.Watch(dir); err != nil {
			log.Warnf("watching %s encountered an error %v", dir, err)
			return
		}
		log.Infof("watching %s for changes", dir)
	}

	watchFunc(fw.Event)
}

//Extracts directory paths from provided filePaths and returns a list of paths with removed duplicates.
func uniqeDirNames(filePaths []string) []string {
	dirMap := make(map[string]bool)
	for _, c := range filePaths {
		dirMap[filepath.Dir(c)] = true
	}

	i := 0
	res := make([]string, len(dirMap))
	for d := range dirMap {
		res[i] = d
		i++
	}
	return res
}

/*
func main() {
	files := []string{"/etc/tls-cert/tls.crt", "/etc/tls-cert/tls.key"}
	ctx := context.TODO()

	var notifyFunc func() = func() {
		log.Infof("Changes detected!, %v", files)
	}
	watcher := NewWatcher(files, 10, notifyFunc)
	watcher.Run(ctx)

	<-ctx.Done()
}
*/
