package resource

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	watchtools "k8s.io/client-go/tools/watch"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/retry"
)

type Resource struct {
	ResCli    dynamic.ResourceInterface
	namespace string
	kind      string
	verbose   bool
	log       *logrus.Entry
}

func New(dynamicCli dynamic.Interface, s schema.GroupVersionResource, namespace string, log *logrus.Entry, verbose bool) *Resource {
	resCli := dynamicCli.Resource(s).Namespace(namespace)
	return &Resource{ResCli: resCli, namespace: namespace, kind: s.Resource, log: log, verbose: verbose}
}

func (r *Resource) Create(res interface{}) (string, error) {
	var resourceVersion string
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(res)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while converting resource %s %+v to unstructured", r.kind, res)
	}
	unstructuredObj := &unstructured.Unstructured{
		Object: u,
	}
	err = retry.OnTimeout(retry.DefaultBackoff, func() error {
		var resource *unstructured.Unstructured

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

	if r.verbose {
		r.log.Infof("[CREATE]: name %s, namespace %s, resource %v", unstructuredObj.GetName(), unstructuredObj.GetNamespace(), unstructuredObj)
	}

	return resourceVersion, nil
}

func (r *Resource) List(set map[string]string) (*unstructured.UnstructuredList, error) {
	var result *unstructured.UnstructuredList

	selector := labels.SelectorFromSet(set).String()

	err := retry.OnTimeout(retry.DefaultBackoff, func() error {
		var err error
		result, err = r.ResCli.List(metav1.ListOptions{
			LabelSelector: selector,
		})
		return err
	}, r.log)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing resource %s in namespace %s", r.kind, r.namespace)
	}
	namespace := "-"
	if r.namespace != "" {
		namespace = r.namespace
	}

	if r.verbose {
		r.log.Infof("LIST %s: namespace:%s kind:%s\n%v", selector, namespace, r.kind, result)
	}

	return result, nil
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

	if r.verbose {
		r.log.Infof("GET name:%s: namespace:%s kind:%s\n%v", name, namespace, r.kind, result)
	}

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

		if r.verbose {
			r.log.Infof("DELETE %s: namespace:%s name:%s", r.kind, namespace, name)
		}

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

func (r *Resource) Update(res interface{}) (*unstructured.Unstructured, error) {
	// https://github.com/kubernetes/client-go/blob/kubernetes-1.17.4/examples/dynamic-create-update-delete-deployment/main.go#L119-L166

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %s %s to unstructured", r.kind, res)
	}

	unstructuredObj := &unstructured.Unstructured{
		Object: u,
	}

	var result *unstructured.Unstructured
	err = retry.WithIgnoreOnConflict(retry.DefaultBackoff, func() error {
		var errUpdate error
		result, errUpdate = r.ResCli.Update(unstructuredObj, metav1.UpdateOptions{})
		return errUpdate
	}, r.log)
	if err != nil {
		return nil, errors.Wrapf(err, "while updating resource %s '%s'", r.kind, unstructuredObj.GetName())
	}

	namespace := "-"
	if r.namespace != "" {
		namespace = r.namespace
	}

	r.log.Infof("UPDATE %s: namespace:%s kind:%s", result.GetName(), namespace, r.kind)
	if r.verbose {
		r.log.Infof("%+v", result)
	}

	return result, nil
}
