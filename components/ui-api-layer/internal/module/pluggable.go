package module

import (
	"reflect"
)

type Pluggable struct {
	name      string
	isEnabled bool
	SyncCh    chan bool
	stopCh    chan struct{}
}

func NewPluggable(name string) *Pluggable {
	return &Pluggable{name: name}
}

func (p *Pluggable) Name() string {
	return p.name
}

func (p *Pluggable) IsEnabled() bool {
	return p.isEnabled
}

func (p *Pluggable) Enable() {
	p.isEnabled = true
	p.stopCh = make(chan struct{})
	p.SyncCh = make(chan bool)
}

func (p *Pluggable) EnableAndSyncCache(sync func(stopCh chan struct{})) {
	p.Enable()

	go func(stopCh chan struct{}, syncCh chan bool) {
		sync(stopCh)
		syncCh <- true
	}(p.stopCh, p.SyncCh)
}

func (p *Pluggable) EnableAndSyncInformerFactory(informerFactory SharedInformerFactory, onSync func()) {
	p.Enable()

	go func(informerFactory SharedInformerFactory, onSyncFn func(), syncCh chan bool) {
		informerFactory.Start(p.stopCh)
		informerFactory.WaitForCacheSync(p.stopCh)
		onSyncFn()
		syncCh <- true
	}(informerFactory, onSync, p.SyncCh)
}

func (p *Pluggable) Disable(disableModule func(disabledErr error)) {
	p.isEnabled = false

	if p.stopCh != nil {
		close(p.stopCh)
	}

	disabledErr := NewDisabledModuleError(p.name)
	disableModule(disabledErr)
}

func (p *Pluggable) StopCacheSyncOnClose(stopCh <-chan struct{}) {
	go func() {
		<-stopCh

		if p.stopCh == nil {
			return
		}
		close(p.stopCh)
	}()
}

type SharedInformerFactory interface {
	Start(stopCh <-chan struct{})
	WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool
}
