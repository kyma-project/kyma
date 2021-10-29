package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fsnotify/fsnotify"
	_ "github.com/pkg/errors"
	"go.uber.org/zap"
)

const dataSourcesPath = "/etc/grafana/provisioning/datasources"
const grafanaPsName = "grafana-server"
const grafanaRun = "./run.sh"

var (
	logger *zap.SugaredLogger
)

type watcher interface {
	stop() error
	startGrafana() error
	attributes() *grafanaAttributes
	killProcess() error
}

type grafanaWatcher struct {
	attr *grafanaAttributes
}

type grafanaAttributes struct {
	path    string
	process string
	grafana *fsnotify.Watcher
	cmd     *exec.Cmd
}

func main() {
	done := make(chan bool)

	if err := initLogger(); err != nil {
		return
	}
	defer func(logger *zap.SugaredLogger) {
		err := logger.Sync()
		if err != nil {
			return
		}
	}(logger)

	watcher := &grafanaWatcher{&grafanaAttributes{path: dataSourcesPath, process: grafanaPsName}}
	err := start(watcher)
	if err != nil {
		logger.Error("error occurred in grafana-watcher:", err)
		watcher.attributes().cmd = nil
		return
	}

	<-done
}

func start(w watcher) error {

	if err := w.startGrafana(); err != nil {
		return err
	}
	logger.Info("Start watching Grafana datasource directory")
	var err error
	w.attributes().grafana, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := w.attributes().grafana.Add(w.attributes().path); err != nil {
		logger.Error("Error watching path: ", err)
	}

	go func() {
		for {
			select {
			case event := <-w.attributes().grafana.Events:
				switch {
				case event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write:
					if err := w.killProcess(); err != nil {
						logger.Errorf("Error killing process: %s", err)
					} else {
						if err := w.startGrafana(); err != nil {
							logger.Errorf("Error restarting Grafana: %s", err)
						}
					}
					logger.Infof("Datasource directory modified: %s", event.String())
				}
			case err := <-w.attributes().grafana.Errors:
				fmt.Printf("Error %s", err)
			}
		}
	}()
	return nil
}

func (g *grafanaWatcher) attributes() *grafanaAttributes {
	return g.attr
}

func (g *grafanaWatcher) startGrafana() error {
	logger.Info("Starting Grafana...")
	g.attr.cmd = exec.Command(grafanaRun)
	g.attr.cmd.Stdout = os.Stdout
	g.attr.cmd.Stderr = os.Stderr
	if err := g.attr.cmd.Start(); err != nil {
		logger.Error("error occurred in grafana start:", err)
		g.attr.cmd = nil
		return err
	}
	logger.Info("Grafana successfully started")
	return nil
}

func (g *grafanaWatcher) killProcess() error {
	psName := g.attributes().process
	g.attr.cmd = exec.Command("pkill", psName)
	g.attr.cmd.Stdout = os.Stdout
	g.attr.cmd.Stderr = os.Stderr
	if err := g.attr.cmd.Start(); err != nil {
		logger.Error("error occurred in process shutdown:", err)
		g.attr.cmd = nil
		return err
	}
	logger.Info("Grafana successfully stopped")
	return nil
}

func (g *grafanaWatcher) stop() error {
	return g.attr.grafana.Close()
}

func initLogger() error {
	rawLogger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	logger = rawLogger.Sugar()
	return nil
}
