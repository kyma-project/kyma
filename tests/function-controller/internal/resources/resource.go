package resources

import (
	"context"
	"github.com/kyma-project/kyma/tests/function-controller/internal/utils"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
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
	err = utils.OnTimeout(utils.DefaultBackoff, func() error {
		var resource *unstructured.Unstructured

		resource, err = r.ResCli.Create(context.Background(), unstructuredObj, metav1.CreateOptions{})
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

	err := utils.OnTimeout(utils.DefaultBackoff, func() error {
		var err error
		result, err = r.ResCli.List(context.Background(), metav1.ListOptions{
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
	err := utils.OnTimeout(utils.DefaultBackoff, func() error {
		var err error
		result, err = r.ResCli.Get(context.Background(), name, metav1.GetOptions{})
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

func WaitUntilConditionSatisfied(ctx context.Context, resCli dynamic.ResourceInterface, isReady func(event watch.Event) (bool, error)) error {
	watcher, err := resCli.Watch(ctx, metav1.ListOptions{})
	defer func() {
		if watcher != nil {
			watcher.Stop()
		}
	}()
	if err != nil {

		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case result := <-watcher.ResultChan():
			ready, err := isReady(result)
			if err != nil {
				return err
			}
			if ready {
				return nil
			}
		}
	}
}

func (r *Resource) Delete(name string) error {
	return utils.WithIgnoreOnNotFound(utils.DefaultBackoff, func() error {
		namespace := "-"
		if r.namespace != "" {
			namespace = r.namespace
		}

		if r.verbose {
			r.log.Infof("DELETE %s: namespace:%s name:%s", r.kind, namespace, name)
		}

		// if the DeletePropagationForeground is not enough then we'll need to somehow watch specified resource
		// and make sure that it was deleted manually
		// in the moment of writing this comment we do not have such a case in those tests
		// that's why we'll just hope DeletePropagationForeground is enough
		deletePropagationPolicy := metav1.DeletePropagationForeground
		return r.ResCli.Delete(context.Background(), name, metav1.DeleteOptions{
			PropagationPolicy: &deletePropagationPolicy,
		})
	}, r.log)
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

	result, err = r.ResCli.Update(context.Background(), unstructuredObj, metav1.UpdateOptions{})
	if err != nil {
		// upstream caller RetryOnConflict doesn't work with wrapped errors
		// https://github.com/kubernetes/client-go/blob/9927afa2880713c4332723b7f0865adee5e63a7b/util/retry/util.go#L89-L93
		return nil, err
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
