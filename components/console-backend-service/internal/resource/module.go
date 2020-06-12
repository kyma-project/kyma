package resource

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ServiceCreator func(factory *ServiceFactory) (*Service, error)
type ServiceCreators map[schema.GroupVersionResource]ServiceCreator

type Module struct {
	*module.Pluggable
	serviceCreators ServiceCreators
	services map[schema.GroupVersionResource]*Service

	// Module should receive ServiceFactory in Enable, as it is needed (and should be needed) only there,
	// but I don't want to touch module.PluggableModule interface right now
	factory *ServiceFactory
}

func NewModule(name string, factory *ServiceFactory, serviceCreators ServiceCreators) *Module {
	return &Module{
		Pluggable: module.NewPluggable(name),
		serviceCreators: serviceCreators,
		services: make(map[schema.GroupVersionResource]*Service),
		factory: factory,
	}
}

func (m *Module) Enable() error {
	newServices := make(map[schema.GroupVersionResource]*Service)
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
		newServices := make(map[schema.GroupVersionResource]*Service)
		for resource := range m.serviceCreators {
			newServices[resource] = &Service{ServiceBase: disabledServiceBase{gvr: resource, err: err}}
		}
		m.services = newServices
	})
	return nil
}

func (m *Module) Service(gvr schema.GroupVersionResource) *Service {
	return m.services[gvr]
}