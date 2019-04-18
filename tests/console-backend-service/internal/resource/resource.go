package resource

import (
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Resource struct {
	resCli    dynamic.ResourceInterface
	namespace string
	kind      string

	log func(format string, args ...interface{})
}

func New(dynamicCli dynamic.Interface, s schema.GroupVersionResource, namespace string, logFn func(format string, args ...interface{})) *Resource {
	resCli := dynamicCli.Resource(s).Namespace(namespace)

	return &Resource{resCli: resCli, namespace: namespace, kind: s.Resource, log: logFn}
}

func (r *Resource) Create(res interface{}) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(res)
	if err != nil {
		return errors.Wrapf(err, "while converting resource %s %s to unstructured", r.kind, res)
	}

	unstructuredObj := &unstructured.Unstructured{
		Object: u,
	}

	_, err = r.resCli.Create(unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			r.log("Cannot create. Resource %s with name '%s' already exist.", unstructuredObj.GetKind(), unstructuredObj.GetName())
			return nil
		}
		return errors.Wrapf(err, "while creating resource %s ", unstructuredObj.GetKind())
	}

	return nil
}

func (r *Resource) Get(name string) (*unstructured.Unstructured, error) {
	u, err := r.resCli.Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, err
		}

		return nil, errors.Wrapf(err, "while getting resource %s '%s'", r.kind, name)
	}

	return u, nil
}

func (r *Resource) Delete(name string) error {
	err := r.resCli.Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			r.log("Cannot delete. Resource %s with name '%s' is not found.", r.kind, name)
			return nil
		}
		return errors.Wrapf(err, "while deleting resource %s '%s'", r.kind, name)
	}

	return nil
}
