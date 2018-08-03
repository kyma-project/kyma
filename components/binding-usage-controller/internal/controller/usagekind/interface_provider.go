package usagekind

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type resourceInterfaceProvider struct {
	resourceInterface dynamic.Interface
	kind              string
}

func newResourceInterfaceProvider(cp dynamic.ClientPool, gvk schema.GroupVersionKind) (*resourceInterfaceProvider, error) {
	resInterface, err := cp.ClientForGroupVersionKind(gvk)

	if err != nil {
		return nil, errors.Wrapf(err, "while creating client for %s", gvk)
	}
	return &resourceInterfaceProvider{
		resourceInterface: resInterface,
		kind:              gvk.Kind,
	}, nil
}

// ResourceInterface provides concrete ResourceInterface for given namespace
func (p *resourceInterfaceProvider) ResourceInterface(namespace string) dynamic.ResourceInterface {
	return p.resourceInterface.Resource(&metav1.APIResource{
		Namespaced: true,
		Name:       p.kind + "s",
	}, namespace)
}
