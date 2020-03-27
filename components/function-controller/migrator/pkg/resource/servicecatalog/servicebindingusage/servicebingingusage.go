package servicebindingusage

import (
	"time"

	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource"
	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/servicecatalog/types/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type ServiceBindingUsage struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
}

func New(dynamicCli dynamic.Interface, name, namespace string, waitTimeout time.Duration, logFn func(format string, args ...interface{})) *ServiceBindingUsage {
	return &ServiceBindingUsage{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha1.SchemeGroupVersion.Version,
			Group:    v1alpha1.SchemeGroupVersion.Group,
			Resource: "servicebindingusages",
		}, namespace, logFn),
		name:        name,
		namespace:   namespace,
		waitTimeout: waitTimeout,
	}
}

func (sbu *ServiceBindingUsage) List(callbacks ...func(...interface{})) ([]*v1alpha1.ServiceBindingUsage, error) {
	ul, err := sbu.resCli.List(callbacks...)
	if err != nil {
		return nil, err
	}

	var sbus []*v1alpha1.ServiceBindingUsage

	for _, u := range ul.Items {
		var res v1alpha1.ServiceBindingUsage
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), &res)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting ServiceBindingUsage %s in namespace %s", u.GetName(), u.GetNamespace())
		}

		sbus = append(sbus, &res)
	}

	return sbus, nil
}

func (sbu *ServiceBindingUsage) Update(resource *v1alpha1.ServiceBindingUsage, callbacks ...func(...interface{})) error {
	_, err := sbu.resCli.Update(resource, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while updating ServiceBindingUsage %s in namespace %s", sbu.name, sbu.namespace)
	}

	return nil
}
