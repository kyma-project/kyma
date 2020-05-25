package testing

import (
	"k8s.io/client-go/dynamic/dynamicinformer"
	"testing"
	"time"

	"k8s.io/client-go/tools/cache"
)

func WaitForInformerStartAtMost(t *testing.T, timeout time.Duration, informer cache.SharedIndexInformer) {
	stop := make(chan struct{})
	syncedDone := make(chan struct{})

	go func() {
		if !cache.WaitForCacheSync(stop, informer.HasSynced) {
			t.Fatalf("timeout occurred when waiting to sync informer")
		}
		close(syncedDone)
	}()

	go informer.Run(stop)

	select {
	case <-time.After(timeout):
		close(stop)
	case <-syncedDone:
	}
}

func WaitForInformerFactoryStartAtMost(t *testing.T, timeout time.Duration, informer dynamicinformer.DynamicSharedInformerFactory) {
	stop := make(chan struct{})
	syncedDone := make(chan struct{})

	go func() {
		informer.WaitForCacheSync(stop)
		close(syncedDone)
	}()

	go informer.Start(stop)

	select {
	case <-time.After(timeout):
		close(stop)
	case <-syncedDone:
	}
}
