package resource

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Resource struct {
	resCli dynamic.ResourceInterface
	namespace  string
	kind string
}

func New(dynamicCli dynamic.Interface, s schema.GroupVersionResource, namespace string) *Resource {
	resCli := dynamicCli.Resource(s).Namespace(namespace)

	return &Resource{resCli: resCli, namespace: namespace, kind: s.GroupResource().String()}
}

func (r *Resource) Create(res interface{}) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(res)
	if err != nil {
		return errors.Wrapf(err, "while converting resource %s to unstructured", r.kind, res)
	}

	unstructuredBucket := &unstructured.Unstructured{
		Object: u,
	}

	_, err = r.resCli.Create(unstructuredBucket, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrapf(err, "while creating resource %s ", unstructuredBucket.GetKind())
	}

	return nil
}

func (r *Resource) Delete(name string) error {
	err := r.resCli.Delete(name, &metav1.DeleteOptions{})
	if err != nil {
	 	return errors.Wrapf(err, "while deleting resource %s %s", r.kind, name)
	}

	return nil
}
