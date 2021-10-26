package main

import (
	"fmt"
	"github.com/docker/docker/pkg/integration/cmd"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/process"
	"go.uber.org/zap"
	"os"
	"os/exec"
)

const dataSourcesPath = "/etc/grafana/provisioning/datasources"
const grafanaPsName = "grafana-server"
const grafanaRun = "./run.sh"

var (
	logger *zap.SugaredLogger
)

type watcher struct {
	grafana    *fsnotify.Watcher
	path       string
	eventCount int
	cmd *exec.Cmd
}

func main() {
	done := make(chan bool)

	rawLogger, _ := zap.NewProduction()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			return
		}
	}(rawLogger)
	logger = rawLogger.Sugar()

	watcher := watcher{path: dataSourcesPath}
	err := watcher.start()
	if err != nil {
		logger.Error("error occurred in grafana-watcher:", err)
		watcher.cmd = nil
		return
	}

	<-done
}

func (w *watcher) start() error {
	logger.Info("Starting Grafana...")
	if err := w.startGrafana(); err != nil {
		return err
	}

	logger.Info("Start watching Grafana datasource directory")
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
				case event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Remove == fsnotify.Remove:
					if err := killProcess(grafanaPsName); err != nil {
						logger.Errorf("Error killing process: %s", err)
					} else {
						if err := w.startGrafana(); err != nil {
							logger.Errorf("Error restarting Grafana: %s", err)
						}
					}
					w.eventCount++ // TODO: Make it more testable
					logger.Infof("Datasource directory modified: %s", event.String())
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

func (w *watcher) startGrafana() error{
	w.cmd = exec.Command(grafanaRun)
	w.cmd.Stdout = os.Stdout
	w.cmd.Stderr = os.Stderr
	if err := w.cmd.Start(); err != nil {
		logger.Error("error occurred in grafana start:", err)
		w.cmd = nil
		return err
	}
	logger.Info("Grafana successfully started")
	return nil
}

func killProcess(name string) error {
	processes, err := process.Processes()
	if err != nil {
		return err
	}
	for _, p := range processes {
		n, err := p.Name()
		if err != nil {
			return err
		}
		if n == name {
			logger.Infof("Killing process: %s", name)
			return p.Kill()
		}
	}
	return fmt.Errorf("process not found: %s", name)
}