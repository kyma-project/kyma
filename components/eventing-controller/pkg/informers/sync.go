package informers

import (
	"context"
	"log"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
)

const (
	DefaultResyncPeriod = 10 * time.Second
)

type waitForCacheSyncFunc func(stopCh <-chan struct{}) map[schema.GroupVersionResource]bool

// WaitForCacheSyncOrDie waits for the cache to sync. If sync fails everything stops
func WaitForCacheSyncOrDie(ctx context.Context, dc dynamicinformer.DynamicSharedInformerFactory) {
	dc.Start(ctx.Done())

	ctx, cancel := context.WithTimeout(context.Background(), DefaultResyncPeriod)
	defer cancel()

	err := hasSynced(ctx, dc.WaitForCacheSync)
	if err != nil {
		log.Fatalf("Failed to sync informer caches: %v", err)
	}
}

func hasSynced(ctx context.Context, fn waitForCacheSyncFunc) error {
	// synced gets closed as soon as fn returns
	synced := make(chan struct{})
	// closing stopWait forces fn to return, which happens whenever ctx
	// gets canceled
	stopWait := make(chan struct{})
	defer close(stopWait)

	// close the synced channel if the `WaitForCacheSync()` finished the execution cleanly
	go func() {
		informersCacheSync := fn(stopWait)
		res := true
		for _, sync := range informersCacheSync {
			if !sync {
				res = false
			}
		}
		if res {
			close(synced)
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-synced:
	}

	return nil
}
