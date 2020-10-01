package resource

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sync"
)

//go:generate mockery -name=Listener -output=automock -outpkg=automock -case=underscore

type NotifierXX struct {
	sync.RWMutex
	// TODO: change to map for better performance
	listeners []*Listener
}

func NewNotifierXX() *NotifierXX {
	return &NotifierXX{}
}

func (n *NotifierXX) AddListener(listener *Listener) {
	if listener == nil {
		return
	}

	n.Lock()
	defer n.Unlock()

	n.listeners = append(n.listeners, listener)
}

func (n *NotifierXX) DeleteListener(listener *Listener) {
	if listener == nil {
		return
	}

	n.Lock()
	defer n.Unlock()

	filtered := n.listeners[:0]
	for _, l := range n.listeners {
		if l != listener {
			filtered = append(filtered, l)
		}
	}

	n.listeners = filtered
}

func (n *NotifierXX) OnAdd(obj runtime.Object) {
	n.RLock()
	defer n.RUnlock()

	for _, listener := range n.listeners {
		listener.OnAdd(obj)
	}
}

func (n *NotifierXX) OnUpdate(oldObj, newObj runtime.Object) {
	n.RLock()
	defer n.RUnlock()

	for _, listener := range n.listeners {
		listener.OnUpdate(oldObj, newObj)
	}
}

func (n *NotifierXX) OnDelete(obj runtime.Object) {
	n.RLock()
	defer n.RUnlock()

	for _, listener := range n.listeners {
		listener.OnDelete(obj)
	}
}
