package config

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

func TestRunOnConfigChange(t *testing.T) {
	t.Run("run on canfig change and cancel context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cfgFile := fixConfig(t)
		defer os.Remove(cfgFile.Name())

		callbackChan := make(chan bool)
		done := make(chan bool)
		go func() {
			RunOnConfigChange(ctx, zap.NewNop().Sugar(), cfgFile.Name(), func(c Config) {
				callbackChan <- true
			})

			done <- true
		}()

		quitChan := modifyFileEveryTick(t, cfgFile, 500*time.Millisecond)

		assert.Equal(t, true, <-callbackChan)

		quitChan <- true
		cancel()

		assert.Equal(t, true, <-done)
	})
}

func modifyFileEveryTick(t *testing.T, file *os.File, interval time.Duration) chan interface{} {
	ticker := time.NewTicker(interval)

	quit := make(chan interface{})
	go func() {
		for {
			select {
			case <-ticker.C:
				err := os.WriteFile(file.Name(), []byte("{}"), 0o644)
				assert.NoError(t, err)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return quit
}

func fixConfig(t *testing.T) *os.File {
	file, err := ioutil.TempFile(os.TempDir(), "test-*")
	assert.NoError(t, err)

	bytes, err := yaml.Marshal(&Config{
		LogLevel:  "debug",
		LogFormat: "json",
	})
	assert.NoError(t, err)

	_, err = file.Write(bytes)
	assert.NoError(t, err)

	return file
}
