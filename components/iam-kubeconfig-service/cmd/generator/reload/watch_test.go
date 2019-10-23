package reload

import (
	"context"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/howeyc/fsnotify"
)

type TestAgent struct {
	configCh chan interface{}
}

func (ta *TestAgent) Restart(c interface{}) {
	ta.configCh <- c
}

func (ta *TestAgent) Run(ctx context.Context) {
	<-ctx.Done()
}

func TestEventsBatchingDelay(t *testing.T) {

	lock := sync.Mutex{}
	called := 0

	//This function is invoked by Watcher after processing events in configured minDelaySeconds time window
	notifyFunc := func() {
		lock.Lock()
		defer lock.Unlock()
		called++
	}

	var minDelaySeconds uint8 = 1 //collect events for one second before notifying

	initialDelay := 900 * time.Millisecond
	nextDelay := 200 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	wch := make(chan *fsnotify.FileEvent, 10)

	w := watcher{
		"test",
		[]string{}, //We don't really look for any files, hence the empty slice
		minDelaySeconds,
		notifyFunc,
	}
	go w.watchFileEvents(ctx, wch)

	// fire off multiple events
	wch <- &fsnotify.FileEvent{Name: "event1"}
	wch <- &fsnotify.FileEvent{Name: "event2"}
	wch <- &fsnotify.FileEvent{Name: "event3"}

	// sleep for less than a second
	time.Sleep(initialDelay)

	// Expect no events to be delivered within initialDelay.
	lock.Lock()
	if called != 0 {
		t.Fatalf("Called %d times, want 0", called)
	}
	lock.Unlock()

	// wait for long enough to ensure notification
	time.Sleep(nextDelay)

	// Expect exactly 1 event to be delivered.
	lock.Lock()
	defer lock.Unlock()
	if called != 1 {
		t.Fatalf("Called %d times, want 1", called)
	}

	cancel()
}

func TestWatchSingleFile(t *testing.T) {
	// create a temp dir
	tmpDir, err := ioutil.TempDir(os.TempDir(), "certs")
	if err != nil {
		t.Fatalf("failed to create a temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("failed to remove temp dir: %v", err)
		}
	}()

	// create a temp file
	tmpFile, err := ioutil.TempFile(tmpDir, "test.file")
	if err != nil {
		t.Fatalf("failed to create a temp file in testdata/certs: %v", err)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			t.Errorf("failed to close file %s: %v", tmpFile.Name(), err)
		}
	}()

	called := make(chan bool)
	callbackFunc := func() {
		called <- true
	}

	// test modify file event
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watcher := NewWatcher("test", []string{tmpFile.Name()}, 1, callbackFunc)
	go watcher.Run(ctx)

	// sleep for a bit to make sure the watcher is set up before change is made
	time.Sleep(time.Millisecond * 500)

	// modify file
	if _, err := tmpFile.Write([]byte("foo")); err != nil {
		t.Fatalf("failed to update file %s: %v", tmpFile.Name(), err)
	}

	if err := tmpFile.Sync(); err != nil {
		t.Fatalf("failed to sync file %s: %v", tmpFile.Name(), err)
	}

	t.Logf("Waiting for notification after file data change")
	select {
	case <-called:
		// expected
		break
	case <-time.After(time.Millisecond * 1100):
		t.Fatalf("The callback is not called within time limit " + time.Now().String() + " when file was modified")
	}

	// test delete file event
	// delete the file
	err = os.Remove(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to delete file %s: %v", tmpFile.Name(), err)
	}

	t.Logf("Waiting for notification after file is deleted")
	select {
	case <-called:
		// expected
		break
	case <-time.After(time.Millisecond * 1100):
		t.Fatalf("The callback is not called within time limit " + time.Now().String() + " when file was deleted")
	}
}

func TestWatchMultipleFiles(t *testing.T) {
	// create a temp dir
	tmpDir, err := ioutil.TempDir(os.TempDir(), "certs")
	if err != nil {
		t.Fatalf("failed to create a temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("failed to remove temp dir: %v", err)
		}
	}()

	// create a temp file
	tmpFile1, err := ioutil.TempFile(tmpDir, "test1.file")
	if err != nil {
		t.Fatalf("failed to create a temp file in testdata/certs: %v", err)
	}
	defer func() {
		if err := tmpFile1.Close(); err != nil {
			t.Errorf("failed to close file %s: %v", tmpFile1.Name(), err)
		}
	}()

	// create a temp file
	tmpFile2, err := ioutil.TempFile(tmpDir, "test2.file")
	if err != nil {
		t.Fatalf("failed to create a temp file in testdata/certs: %v", err)
	}
	defer func() {
		if err := tmpFile2.Close(); err != nil {
			t.Errorf("failed to close file %s: %v", tmpFile2.Name(), err)
		}
	}()

	called := make(chan bool)
	callbackFunc := func() {
		called <- true
	}

	// test modify file event
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watcher := NewWatcher("test", []string{tmpFile1.Name(), tmpFile2.Name()}, 1, callbackFunc)
	go watcher.Run(ctx)

	// sleep for a bit to make sure the watcher is set up before change is made
	time.Sleep(time.Millisecond * 500)

	// modify first file
	if _, err := tmpFile1.Write([]byte("foo")); err != nil {
		t.Fatalf("failed to update file %s: %v", tmpFile1.Name(), err)
	}

	if err := tmpFile1.Sync(); err != nil {
		t.Fatalf("failed to sync file %s: %v", tmpFile1.Name(), err)
	}

	t.Logf("Waiting for notification after the first file data change")
	select {
	case <-called:
		// expected
		break
	case <-time.After(time.Millisecond * 1100):
		t.Fatalf("The callback is not called within time limit " + time.Now().String() + " when file was modified")
	}

	// modify second file
	if _, err := tmpFile2.Write([]byte("bar")); err != nil {
		t.Fatalf("failed to update file %s: %v", tmpFile2.Name(), err)
	}

	if err := tmpFile2.Sync(); err != nil {
		t.Fatalf("failed to sync file %s: %v", tmpFile2.Name(), err)
	}

	t.Logf("Waiting for notification after the second file data change")
	select {
	case <-called:
		// expected
		break
	case <-time.After(time.Millisecond * 1100):
		t.Fatalf("The callback is not called within time limit " + time.Now().String() + " when file was modified")
	}

	// test delete file event
	// delete the file
	err = os.Remove(tmpFile1.Name())
	if err != nil {
		t.Fatalf("failed to delete file %s: %v", tmpFile1.Name(), err)
	}

	t.Logf("Waiting for notification after the first file is deleted")
	select {
	case <-called:
		// expected
		break
	case <-time.After(time.Millisecond * 1100):
		t.Fatalf("The callback is not called within time limit " + time.Now().String() + " when file was deleted")
	}
}
