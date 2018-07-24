package resource

import "sync"

//go:generate mockery -name=Listener -output=automock -outpkg=automock -case=underscore
type Listener interface {
	OnAdd(object interface{})
	OnUpdate(newObject, oldObject interface{})
	OnDelete(object interface{})
}

type Notifier interface {
	OnAdd(object interface{})
	OnUpdate(newObject, oldObject interface{})
	OnDelete(object interface{})
	AddListener(observer Listener)
	DeleteListener(observer Listener)
}

type notifier struct {
	sync.RWMutex
	// TODO: change to map for better performance
	listeners []Listener
}

func NewNotifier() Notifier {
	return new(notifier)
}

func (n *notifier) AddListener(listener Listener) {
	if listener == nil {
		return
	}

	n.Lock()
	defer n.Unlock()

	n.listeners = append(n.listeners, listener)
}

func (n *notifier) DeleteListener(listener Listener) {
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

func (n *notifier) OnAdd(obj interface{}) {
	n.RLock()
	defer n.RUnlock()

	for _, listener := range n.listeners {
		listener.OnAdd(obj)
	}
}

func (n *notifier) OnUpdate(oldObj, newObj interface{}) {
	n.RLock()
	defer n.RUnlock()

	for _, listener := range n.listeners {
		listener.OnUpdate(oldObj, newObj)
	}
}

func (n *notifier) OnDelete(obj interface{}) {
	n.RLock()
	defer n.RUnlock()

	for _, listener := range n.listeners {
		listener.OnDelete(obj)
	}
}
