package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"
)

func Test_watcher_watch(t *testing.T) {

	if err := initLogger(); err != nil{
		return
	}
	defer logger.Sync()

	wd, _ := os.Getwd()
	directory := filepath.Join(wd, "testdata")
	file := "testdata.txt"

	err := os.MkdirAll(directory, os.ModePerm)
	assert.NoError(t, err)



	t.Run("Creating file event notified", func(t *testing.T) {
		var testWatcher = &grafanaWatcherMock{1, &grafanaAttributes{path:directory}}
		err = start(testWatcher)
		assert.NoError(t, err)

		_, err = os.Create(filepath.Join(directory, file))
		assert.NoError(t, err)
		time.Sleep(1 * time.Second) // TODO mach mal hübsch
		assert.Equal(t, 2, testWatcher.eventCount)

		err = os.Remove(path.Join(directory, file))
		assert.NoError(t, err)
		time.Sleep(1 * time.Second)
		assert.Equal(t, 3, testWatcher.eventCount)

		err = testWatcher.stop() // TODO defer
		if err != nil {
		}
	})

	t.Cleanup(func() {
		err := os.RemoveAll(directory)
		if err != nil{
			// TODO: Do something, research
		}
	})


	// create file
	// assert testWatcher event count
	// stop watcher
}

type grafanaWatcherMock struct {
	eventCount int
	attr *grafanaAttributes
}

func (g *grafanaWatcherMock) stop() error{
	return nil
}
func (g *grafanaWatcherMock) startGrafana() error{
	return nil
}
func (g *grafanaWatcherMock) attributes() *grafanaAttributes{
	return g.attr
}

func (g *grafanaWatcherMock) killProcess() error{
	g.eventCount++
	return nil
}