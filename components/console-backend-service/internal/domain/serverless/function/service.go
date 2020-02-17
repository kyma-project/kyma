package function

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Service struct {
	*resource.Service
}

func NewService(serviceFactory *resource.ServiceFactory) *Service {
	return &Service{
		Service: serviceFactory.ForResource(schema.GroupVersionResource{
			Version:  v1alpha1.SchemeGroupVersion.Version,
			Group:    v1alpha1.SchemeGroupVersion.Group,
			Resource: "functions",
		}),
	}
}

func (svc *Service) List(namespace string) ([]*v1alpha1.Function, error) {
	results := make([]*v1alpha1.Function, 0)
	err := svc.ListInIndex("namespace", namespace, &results)
	return results, err
}

func (svc *Service) Find(name, namespace string) (*v1alpha1.Function, error) {
	var result *v1alpha1.Function
	err := svc.FindInNamespace(name, namespace, &result)
	return result, err
}

func (svc *Service) Delete(name, namespace string) error {
	return svc.Client.Namespace(namespace).Delete(name, &metav1.DeleteOptions{})
}

func (svc *Service) Create(name, namespace string, labels gqlschema.Labels, size, runtime string) (*v1alpha1.Function, error) {
	function, err := toUnstructured(&v1alpha1.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "serverless.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1alpha1.FunctionSpec{
			Size:    size,
			Runtime: runtime,
		},
	})
	if err != nil {
		return &v1alpha1.Function{}, err
	}

	created, err := svc.Client.Namespace(namespace).Create(function, metav1.CreateOptions{})
	if err != nil {
		return &v1alpha1.Function{}, err
	}

	return fromUnstructured(created)
}

func (svc *Service) Update(name, namespace string, params gqlschema.FunctionUpdateInput) (*v1alpha1.Function, error) {

	var oldFunction *v1alpha1.Function
	err := svc.FindInNamespace(name, namespace, &oldFunction)
	if err != nil {
		return nil, err
	}

	if oldFunction == nil {
		return nil, errors.NewNotFound(schema.GroupResource{
			Group:    v1alpha1.SchemeGroupVersion.Group,
			Resource: "functions",
		}, name)
	}

	resourceVersion := oldFunction.ObjectMeta.ResourceVersion

	newFunction, err := toUnstructured(&v1alpha1.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "serverless.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          params.Labels,
			ResourceVersion: resourceVersion,
		},
		Spec: v1alpha1.FunctionSpec{
			Size:     params.Size,
			Runtime:  params.Runtime,
			Function: params.Content,
			Deps:     params.Dependencies,
		},
	})
	if err != nil {
		return &v1alpha1.Function{}, err
	}

	updated, err := svc.Client.Namespace(namespace).Update(newFunction, metav1.UpdateOptions{})
	if err != nil {
		return &v1alpha1.Function{}, err
	}

	return fromUnstructured(updated)
}
