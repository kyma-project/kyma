package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"os"
	"os/exec"
)

const dataSourcesPath = "/etc/grafana/provisioning/datasources"

var (
	logger       *zap.SugaredLogger
	cmd          *exec.Cmd
)

type watcher struct {
	grafana *fsnotify.Watcher
	path    string
	eventCount int
}

func main() {
	cmd = exec.Command("./run.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		logger.Error("grafana start error:", err)
		cmd = nil
		return
	}

	done := make(chan bool)

	watcher := watcher{path: dataSourcesPath}
	err := watcher.start()
	if err != nil {
		// TODO Err Handle
	}

	<-done
}

func (w *watcher) start() error {
	fmt.Println("Start watching grafana datasource directory")
	logger.Info("Start watching grafana datasource directory")
	var err error
	w.grafana, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := w.grafana.Add(w.path); err != nil {
		logger.Error("Error watching path: ", err)
	}

	go func() {
		for {
			select {
			case event := <-w.grafana.Events:
				switch {
				case event.Op&fsnotify.Create == fsnotify.Create:
					// TODO restart grafana: kill and run.sh

					w.eventCount++
					logger.Infof("File created: %s", event.String() )
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					// TODO restart grafana: kill and run.sh

					logger.Infof("File removed: %s", event.String() )
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
