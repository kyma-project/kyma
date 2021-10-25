package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
)

const dataSourcesPath = "etc/grafana/provsioning/datasources"

type watcher struct{
	grafana *fsnotify.Watcher
	path string

	eventCount int
}
func main() {

	done := make(chan bool)

	watcher := watcher{path: dataSourcesPath}
	err := watcher.start()
	if err != nil {
		// TODO Err Handle
	}
	<-done

}

func (w *watcher) start() error {
	var err error
	w.grafana, err = fsnotify.NewWatcher()
	if err != nil{
		return err
	}

	if err := w.grafana.Add(w.path); err != nil {
		fmt.Println("ERROR", err)
	}

	go func(){
		for {
			select {
			case event := <-w.grafana.Events:
				switch {
				//case event.Op&fsnotify.Write == fsnotify.Write:
				//	log.Printf("Write:  %s: %s", event.Op, event.Name)
				case event.Op&fsnotify.Create == fsnotify.Create:
					w.eventCount++
				case event.Op&fsnotify.Remove == fsnotify.Remove:
				 	w.eventCount++
				}
			case err := <-w.grafana.Errors:
				fmt.Printf("Error %s", err)
			}
		}

	}()
	return nil
}

func (w *watcher) stop() error {
	return w.grafana.Close()
}
