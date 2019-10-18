package resource

import (
	"fmt"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/retry"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Resource struct {
	ResCli    dynamic.ResourceInterface
	namespace string
	kind      string

	log func(format string, args ...interface{})
}

func New(dynamicCli dynamic.Interface, s schema.GroupVersionResource, namespace string, logFn func(format string, args ...interface{})) *Resource {
	resCli := dynamicCli.Resource(s).Namespace(namespace)
	return &Resource{ResCli: resCli, namespace: namespace, kind: s.Resource, log: logFn}
}

func (r *Resource) Create(res interface{}, callbacks ...func(...interface{})) (string, error) {
	var resourceVersion string
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(res)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while converting resource %s %s to unstructured", r.kind, res)
	}
	unstructuredObj := &unstructured.Unstructured{
		Object: u,
	}
	err = retry.OnCreateError(retry.DefaultBackoff, func() error {
		var resource *unstructured.Unstructured
		resource, err = r.ResCli.Create(unstructuredObj, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		resourceVersion = resource.GetResourceVersion()
		return nil
	}, callbacks...)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while creating resource %s ", unstructuredObj.GetKind())
	}
	return resourceVersion, nil
}

func (r *Resource) Get(name string, callbacks ...func(...interface{})) (*unstructured.Unstructured, error) {
	var result *unstructured.Unstructured
	err := retry.OnGetError(retry.DefaultBackoff, func() error {
		var err error
		result, err = r.ResCli.Get(name, metav1.GetOptions{})
		return err
	}, callbacks...)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, err
		}
		return nil, errors.Wrapf(err, "while getting resource %s '%s'", r.kind, name)
	}
	for _, callback := range callbacks {
		namespace := "-"
		if r.namespace != "" {
			namespace = r.namespace
		}
		callback(fmt.Sprintf("GET %s: namespace:%s kind:%s\n%v", name, namespace, r.kind, result))
	}
	return result, nil
}

func (r *Resource) Delete(name string, callbacks ...func(...interface{})) error {
	err := retry.OnDeleteError(retry.DefaultBackoff, func() error {
		for _, callback := range callbacks {
			namespace := "-"
			if r.namespace != "" {
				namespace = r.namespace
			}
			callback(fmt.Sprintf("DELETE %s: namespace:%s name:%s", r.kind, namespace, name))
		}
		return r.ResCli.Delete(name, &metav1.DeleteOptions{})
	}, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while deleting resource %s '%s'", r.kind, name)
	}
	return nil
}
