package module

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
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

func (p *Pluggable) EnableAndSyncInformerFactories(onSync func(), informerFactories ...SharedInformerFactory) {
	p.Enable()

	go func(onSyncFn func(), syncCh chan bool, informerFactories ...SharedInformerFactory) {
		for _, i := range informerFactories {
			i.Start(p.stopCh)
		}
		for _, i := range informerFactories {
			i.WaitForCacheSync(p.stopCh)
		}

		onSyncFn()
		syncCh <- true
	}(onSync, p.SyncCh, informerFactories...)
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

func (p *Pluggable) EnableAndSyncDynamicInformerFactory(informerFactory DynamicSharedInformerFactory, onSync func()) {
	p.Enable()

	go func(informerFactory DynamicSharedInformerFactory, onSyncFn func(), syncCh chan bool) {
		fmt.Print(informerFactory)
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

type DynamicSharedInformerFactory interface {
	Start(stopCh <-chan struct{})
	ForResource(gvr schema.GroupVersionResource) informers.GenericInformer
	WaitForCacheSync(stopCh <-chan struct{}) map[schema.GroupVersionResource]bool
}
