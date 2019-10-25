package reload

import (
	"context"
	"path/filepath"
	"time"

	"github.com/howeyc/fsnotify"

	log "github.com/sirupsen/logrus"
)

// Watcher is designed to provide notifications about changes to files mounted inside kubernetes Pod, like Secrets or ConfigMaps
// Because file watches breaks in kubernetes when mounts are updated, we watch for directories instead.
// This happens because files mounted in a Pod are actually symbolic links pointing to "real" files.
// On updating mounted files, kubernetes deletes the existing file, which sends a DELETE file event and breaks the watch
type Watcher interface {
	// Run start the watcher loop (blocking call)
	// context is used to terminate the loop
	Run(context.Context)
}

type watcher struct {
	name                   string   //name of the watcher to improve logging
	filePaths              []string //a single watcher can react to changes to many files
	eventBatchDelaySeconds uint8    //changes that occur within this time window are batched in a single notification
	notifyFunc             func()   //notification function invoked when changes are detected, but only after configured eventBatchDelaySeconds time
}

// NewWatcher creates a new watcher instance
// name is used in logging
// filePaths parameter is a list of file paths to watch
// notifyFunc is a function that is invoked after watcher detects changes to monitored files.
func NewWatcher(name string, filePaths []string, evBatchDelaySeconds uint8, notifyFunc func()) Watcher {
	return &watcher{
		name:                   name,
		filePaths:              filePaths,
		eventBatchDelaySeconds: evBatchDelaySeconds,
		notifyFunc:             notifyFunc,
	}
}

//Run implements Watcher interface
func (w *watcher) Run(ctx context.Context) {
	log.Infof("Watcher [%s] starts watching for files: %v", w.name, w.filePaths)

	watchFileEventsFunc := func(fEventChan <-chan *fsnotify.FileEvent) {
		w.watchFileEvents(ctx, fEventChan)
	}

	dirs := uniqeDirNames(w.filePaths)

	// monitor files
	go func() {
		w.watchForDirs(dirs, watchFileEventsFunc)
	}()

	<-ctx.Done()
	log.Infof("Watcher [%s] has successfully terminated", w.name)
}

// watchFileEvents watches for changes on a channel and notifies via notifyFn().
// The function batches changes so that related changes are processed together in a single step.
// The function ensures that notifyFn() is called no more than one time per eventBatchDelaySeconds.
// The function does not return until the the context is canceled.
func (w *watcher) watchFileEvents(ctx context.Context, wch <-chan *fsnotify.FileEvent) {
	minDelay := time.Second * time.Duration(w.eventBatchDelaySeconds)

	// timer and channel for managing minDelay.
	var timeChan <-chan time.Time
	var timer *time.Timer

	for {
		select {
		case ev := <-wch:
			log.Infof("Watcher[%s]: watchFileEvents: %s", w.name, ev.String())
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

			log.Infof("Watcher[%s]: watchFileEvents: notifying", w.name)
			w.notifyFunc()
		case <-ctx.Done():
			log.Infof("Watcher[%s]: watchFileEvents has successfully terminated", w.name)
			return
		}
	}
}

// watchForDirs configures a watch for every directory path in dirs.
// It then invokes provided watchFunc with configured Watchers.
// This function is blocking so it should be run in a goroutine.
func (w *watcher) watchForDirs(dirs []string, watchFunc func(fEventChan <-chan *fsnotify.FileEvent)) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		log.Warningf("Watcher[%s]: failed to create a watcher for certificate files: %v", w.name, err)
		return
	}
	defer func() {
		if err := fw.Close(); err != nil {
			log.Warningf("Watcher[%s]: closing watcher encounters an error %v", w.name, err)
		}
	}()

	for _, dir := range dirs {
		if err := fw.Watch(dir); err != nil {
			log.Warningf("Watcher[%s]: watching %s encountered an error %v", w.name, dir, err)
			return
		}
		log.Infof("Watcher[%s]: watching %s for changes", w.name, dir)
	}

	watchFunc(fw.Event)
}

//Extracts directory paths from provided filePaths list and returns a list of paths with duplicates removed.
func uniqeDirNames(filePaths []string) []string {
	if len(filePaths) == 0 {
		return []string{}
	}

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
