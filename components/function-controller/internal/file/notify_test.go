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
	t.Run("react to file deletion", func(t *testing.T) {
		file, err := ioutil.TempFile(os.TempDir(), "test-*")
		assert.NoError(t, err)

		notifyErr := make(chan error)
		go func() {
			notifyErr <- NotifyModification(context.Background(), file.Name())
		}()

		time.Sleep(1 * time.Second)
		err = os.Remove(file.Name())
		assert.NoError(t, err)

		assert.NoError(t, <-notifyErr)
	})

	t.Run("file does not exist", func(t *testing.T) {
		notifyErr := make(chan error)
		go func() {
			notifyErr <- NotifyModification(context.Background(), "/path/does/not/exist")
		}()

		time.Sleep(1 * time.Second)

		assert.Error(t, <-notifyErr)
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

		time.Sleep(1 * time.Second)
		cancel()

		assert.Equal(t, <-notifyErr, context.Canceled)
	})
}
