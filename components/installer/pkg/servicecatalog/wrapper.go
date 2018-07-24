package servicecatalog

import (
	serviceCatalog "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	servicecatalogv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"k8s.io/client-go/rest"
)

// Wraps Service Catalog API so we can test the code without the actual API being invoked
type catalogServiceWrapper interface {
	getCatalogAPI() (servicecatalogv1beta1.ServicecatalogV1beta1Interface, error)
}

// Implements catalogServiceWrapper
type defaultCatalogServiceWrapper struct {
	config        *rest.Config
	catalogClient *serviceCatalog.Clientset
}

func (w *defaultCatalogServiceWrapper) getCatalogAPI() (servicecatalogv1beta1.ServicecatalogV1beta1Interface, error) {
	if err := w.lazyInit(); err != nil {
		return nil, err
	}
	return w.catalogClient.ServicecatalogV1beta1(), nil
}

func newDefaultCatalogServiceWrapper(config *rest.Config) catalogServiceWrapper {
	res := defaultCatalogServiceWrapper{
		config: config,
	}
	return &res
}

// Lazy initialization !
func (w *defaultCatalogServiceWrapper) lazyInit() error {

	if w.catalogClient == nil {
		//boostrap new service catalog client
		catalogClient, err := serviceCatalog.NewForConfig(w.config)
		if err != nil {
			return err
		}

		w.catalogClient = catalogClient
	}

	return nil
}
