package servicecatalog

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// ClientInterface exposes functions to interact with ServiceCatalog
type ClientInterface interface {
	GetServiceBindings(ns string) (*v1beta1.ServiceBindingList, error)
	GetServiceInstances(ns string) (*v1beta1.ServiceInstanceList, error)
	DeleteBinding(namespace, name string) error
	DeleteInstance(namespace, name string) error
}

// Client provides functions to interact with Service Catalog.
// It wraps the actual service catalog client which is instantiated lazily (so _this_ Client can be instantiated before Service Catalog is available)
type client struct {
	wrapper catalogServiceWrapper
}

// NewClient returns a pointer to a Client instance configured with provided config.
func NewClient(config *rest.Config) ClientInterface {
	res := client{
		wrapper: newDefaultCatalogServiceWrapper(config),
	}
	return &res
}

// GetServiceBindings returns all ServiceBinding objects from provided namespace.
// Use empty string to return objects from all namespaces.
func (c *client) GetServiceBindings(ns string) (*v1beta1.ServiceBindingList, error) {
	api, err := c.wrapper.getCatalogAPI()
	if err != nil {
		return nil, err
	}

	return api.ServiceBindings(ns).List(v1.ListOptions{})
}

// GetServiceInstances returns all ServiceInstance objects from provided namespace.
// Use empty string to return objects from all namespaces.
func (c *client) GetServiceInstances(ns string) (*v1beta1.ServiceInstanceList, error) {
	api, err := c.wrapper.getCatalogAPI()
	if err != nil {
		return nil, err
	}

	return api.ServiceInstances(ns).List(v1.ListOptions{})
}

// DeleteBinding deletes ServiceBinding with given name from given namespace
func (c *client) DeleteBinding(namespace, name string) error {
	api, err := c.wrapper.getCatalogAPI()
	if err != nil {
		return err
	}

	err = api.ServiceBindings(namespace).Delete(name, &v1.DeleteOptions{})

	if err != nil {
		return err
	}

	return nil
}

// DeleteBinding deletes ServiceBinding with given name from given namespace
func (c *client) DeleteInstance(namespace, name string) error {
	api, err := c.wrapper.getCatalogAPI()
	if err != nil {
		return err
	}

	err = api.ServiceInstances(namespace).Delete(name, &v1.DeleteOptions{})

	if err != nil {
		return err
	}

	return nil
}
