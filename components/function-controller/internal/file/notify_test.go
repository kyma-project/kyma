package file

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNotifyModification(t *testing.T) {
	t.Run("react to file modification", func(t *testing.T) {
		file, err := ioutil.TempFile(os.TempDir(), "test-*")
		assert.NoError(t, err)

		notifyErr := make(chan error)
		go func() {
			notifyErr <- NotifyModification(context.Background(), file.Name())
		}()

		quit := modifyFileEveryTick(t, file, 500*time.Millisecond)

		assert.NoError(t, <-notifyErr)
		quit <- true
	})

	t.Run("file does not exist", func(t *testing.T) {
		notifyErr := make(chan error)
		go func() {
			notifyErr <- NotifyModification(context.Background(), "/path/does/not/exist")
		}()

		err := <-notifyErr
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("cancel context", func(t *testing.T) {
		file, err := ioutil.TempFile(os.TempDir(), "test-*")
		assert.NoError(t, err)
		defer os.Remove(file.Name())

		ctx, cancel := context.WithCancel(context.Background())

		notifyErr := make(chan error)
		go func() {
			notifyErr <- NotifyModification(ctx, file.Name())
		}()

		cancel()

		assert.Equal(t, <-notifyErr, context.Canceled)
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
