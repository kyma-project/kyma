package resource

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ServiceCreator func(factory *GenericServiceFactory) (*GenericService, error)
type ServiceCreators map[schema.GroupVersionResource]ServiceCreator

type Module struct {
	*module.Pluggable
	serviceCreators ServiceCreators
	services        map[schema.GroupVersionResource]*GenericService

	// Module should receive ServiceFactory in Enable, as it is needed (and should be needed) only there,
	// but I don't want to touch module.PluggableModule interface right now
	factory *GenericServiceFactory
}

func NewModule(name string, factory *GenericServiceFactory, serviceCreators ServiceCreators) *Module {
	return &Module{
		Pluggable:       module.NewPluggable(name),
		serviceCreators: serviceCreators,
		services:        make(map[schema.GroupVersionResource]*GenericService),
		factory:         factory,
	}
}

func (m *Module) Enable() error {
	newServices := make(map[schema.GroupVersionResource]*GenericService)
	for resource, creator := range m.serviceCreators {
		var err error
		newServices[resource], err = creator(m.factory)
		if err != nil {
			return err
		}
	}

	m.Pluggable.Enable()
	m.factory.InformerFactory.Start(make(chan struct{}))
	m.factory.InformerFactory.WaitForCacheSync(make(chan struct{}))
	m.services = newServices
	return nil
}

func (m *Module) Disable() error {
	m.Pluggable.Disable(func(err error) {
		newServices := make(map[schema.GroupVersionResource]*GenericService)
		for resource := range m.serviceCreators {
			newServices[resource] = &GenericService{ServiceBase: disabledServiceBase{gvr: resource, err: err}}
		}
		m.services = newServices
	})
	return nil
}

func (m *Module) Service(gvr schema.GroupVersionResource) *GenericService {
	return m.services[gvr]
}
