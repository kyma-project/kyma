package resource

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/retry"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	watchtools "k8s.io/client-go/tools/watch"
)

type logger interface {
	Logf(format string, args ...interface{})
}

type Resource struct {
	ResCli    dynamic.ResourceInterface
	namespace string
	kind      string

	log logger
}

func New(dynamicCli dynamic.Interface, s schema.GroupVersionResource, namespace string, log logger) *Resource {
	resCli := dynamicCli.Resource(s).Namespace(namespace)
	return &Resource{ResCli: resCli, namespace: namespace, kind: s.Resource, log: log}
}

func (r *Resource) Create(res interface{}) (string, error) {
	var resourceVersion string
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(res)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while converting resource %s %s to unstructured", r.kind, res)
	}
	unstructuredObj := &unstructured.Unstructured{
		Object: u,
	}
	err = retry.OnTimeout(retry.DefaultBackoff, func() error {
		var resource *unstructured.Unstructured
		r.log.Logf("[CREATE]: %s", unstructuredObj)
		resource, err = r.ResCli.Create(unstructuredObj, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		resourceVersion = resource.GetResourceVersion()
		return nil
	}, r.log)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while creating resource %s", unstructuredObj.GetKind())
	}
	return resourceVersion, nil
}

func (r *Resource) Get(name string) (*unstructured.Unstructured, error) {
	var result *unstructured.Unstructured
	err := retry.OnTimeout(retry.DefaultBackoff, func() error {
		var err error
		result, err = r.ResCli.Get(name, metav1.GetOptions{})
		return err
	}, r.log)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting resource %s '%s'", r.kind, name)
	}
	namespace := "-"
	if r.namespace != "" {
		namespace = r.namespace
	}
	r.log.Logf("GET %s: namespace:%s kind:%s\n%v", name, namespace, r.kind, result)
	return result, nil
}

func (r *Resource) Delete(name string, timeout time.Duration) error {
	var initialResourceVersion string
	err := retry.OnTimeout(retry.DefaultBackoff, func() error {
		u, err := r.ResCli.Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		initialResourceVersion = u.GetResourceVersion()
		return nil
	}, r.log)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	err = retry.WithIgnoreOnNotFound(retry.DefaultBackoff, func() error {
		namespace := "-"
		if r.namespace != "" {
			namespace = r.namespace
		}
		r.log.Logf("DELETE %s: namespace:%s name:%s", r.kind, namespace, name)
		return r.ResCli.Delete(name, &metav1.DeleteOptions{})
	}, r.log)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	condition := func(event watch.Event) (bool, error) {
		if event.Type != watch.Deleted {
			return false, nil
		}
		u, ok := event.Object.(*unstructured.Unstructured)
		if !ok || u.GetName() != name {
			return false, nil
		}
		return true, nil
	}
	_, err = watchtools.Until(ctx, initialResourceVersion, r.ResCli, condition)
	if err != nil {
		return err
	}
	return nil
}
