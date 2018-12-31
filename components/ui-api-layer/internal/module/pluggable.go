package module

import (
	"reflect"
)

type Pluggable struct {
	name      string
	isEnabled bool
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
}

func (p *Pluggable) EnableAndSyncCache(sync func(stopCh chan struct{})) {
	p.Enable()

	go sync(p.stopCh)
}

func (p *Pluggable) EnableAndSyncInformerFactory(informerFactory SharedInformerFactory, onSync func()) {
	p.Enable()

	go func(informerFactory SharedInformerFactory) {
		informerFactory.Start(p.stopCh)
		informerFactory.WaitForCacheSync(p.stopCh)
		onSync()
	}(informerFactory)
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
		close(p.stopCh)
	}()
}

type SharedInformerFactory interface {
	Start(stopCh <-chan struct{})
	WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool
}
