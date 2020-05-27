package resource

import (
	"sync"
)

//go:generate mockery -name=Listener -output=automock -outpkg=automock -case=underscore

type Notifier struct {
	sync.RWMutex
	// TODO: change to map for better performance
	listeners []*Listener
}

func NewNotifier() *Notifier {
	return &Notifier{}
}

func (n *Notifier) AddListener(listener *Listener) {
	if listener == nil {
		return
	}

	n.Lock()
	defer n.Unlock()

	n.listeners = append(n.listeners, listener)
}

func (n *Notifier) DeleteListener(listener *Listener) {
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

func (n *Notifier) OnAdd(obj interface{}) {
	n.RLock()
	defer n.RUnlock()

	for _, listener := range n.listeners {
		listener.OnAdd(obj)
	}
}

func (n *Notifier) OnUpdate(oldObj, newObj interface{}) {
	n.RLock()
	defer n.RUnlock()

	for _, listener := range n.listeners {
		listener.OnUpdate(oldObj, newObj)
	}
}

func (n *Notifier) OnDelete(obj interface{}) {
	n.RLock()
	defer n.RUnlock()

	for _, listener := range n.listeners {
		listener.OnDelete(obj)
	}
}
