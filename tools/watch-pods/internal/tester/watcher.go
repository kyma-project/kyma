package tester

import (
	"errors"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

var eventHandlerResyncPeriod = time.Second * 30

// NewPodWatcher creates pod watcher.
//
// onPodUpdate function will be called every time update for pod occurs, and onPodDelete on each delete event.
// Both functions should be fast, and O(1) as they are blocking event ingestion by shared informer.
func NewPodWatcher(podListWatcher cache.ListerWatcher, onPodUpdate func(pod *v1.Pod), onPodDelete func(ns, name string)) *PodWatcher {
	sharedInformer := cache.NewSharedIndexInformer(podListWatcher, &v1.Pod{}, eventHandlerResyncPeriod, cache.Indexers{})

	w := &PodWatcher{
		sharedInformer: sharedInformer,
		stopCh:         make(chan struct{}),
		onPodUpdate:    onPodUpdate,
		onPodDelete:    onPodDelete,
	}

	sharedInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(_, new interface{}) {
			p, ok := new.(*v1.Pod)
			if !ok {
				return
			}
			pCpy := p.DeepCopy()
			w.onPodUpdate(pCpy)
		},
		DeleteFunc: func(old interface{}) {
			p, ok := old.(*v1.Pod)
			if !ok {
				return
			}
			w.onPodDelete(p.GetNamespace(), p.GetName())
		},
		AddFunc: func(obj interface{}) {
			p, ok := obj.(*v1.Pod)
			if !ok {
				return
			}
			pCpy := p.DeepCopy()
			w.onPodUpdate(pCpy)
		},
	})

	return w
}

// PodWatcher allows to listen for changes in pods
type PodWatcher struct {
	sharedInformer cache.SharedIndexInformer
	stopCh         chan struct{}
	onPodUpdate    func(pod *v1.Pod)
	onPodDelete    func(ns, pod string)
}

// StartListeningToEvents starts listening on pod events
func (w *PodWatcher) StartListeningToEvents() error {
	go w.sharedInformer.Run(w.stopCh)

	if !cache.WaitForCacheSync(w.stopCh, w.sharedInformer.HasSynced) {
		return errors.New("cannot sync cache")
	}

	return nil
}

// Stop stops watching on pod events. It has to be called to properly free up all resources.
// Be aware that resource release is only signaled but there is no way to wait till resources are actually released.
func (w *PodWatcher) Stop() {
	close(w.stopCh)
}
