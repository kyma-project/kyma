package usagekind

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type resourceInterfaceProvider struct {
	resourceInterface dynamic.NamespaceableResourceInterface
}

func newResourceInterfaceProvider(client dynamic.Interface, gvr schema.GroupVersionResource) (*resourceInterfaceProvider, error) {
	resInterface := client.Resource(gvr)

	return &resourceInterfaceProvider{
		resourceInterface: resInterface,
	}, nil
}
