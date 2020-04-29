package testsuite

import (
	"context"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	watchtools "k8s.io/client-go/tools/watch"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

var (
	ErrInvalidDataType = errors.New("invalid data type")
)

type function struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         shared.Logger
}

func newFunction(dynamicCli dynamic.Interface, name, namespace string, waitTimeout time.Duration, log shared.Logger) *function {
	return &function{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  serverlessv1alpha1.GroupVersion.Version,
			Group:    serverlessv1alpha1.GroupVersion.Group,
			Resource: "functions",
		}, namespace, log),
		name:        name,
		namespace:   namespace,
		waitTimeout: waitTimeout,
		log:         log,
	}
}

func (f *function) Create(data *functionData) (string, error) {
	function := &serverlessv1alpha1.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.name,
			Namespace: f.namespace,
		},
		Spec: serverlessv1alpha1.FunctionSpec{
			Source: data.Body,
			Deps:   data.Deps,
		},
	}

	resourceVersion, err := f.resCli.Create(function)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while creating Function %s in namespace %s", f.name, f.namespace)
	}
	return resourceVersion, err
}

func (f *function) WaitForStatusRunning(initialResourceVersion string) error {
	ctx, cancel := context.WithTimeout(context.Background(), f.waitTimeout)
	defer cancel()
	condition := f.isFunctionReady(f.name)
	_, err := watchtools.Until(ctx, initialResourceVersion, f.resCli.ResCli, condition)
	if err != nil {
		return err
	}
	return nil
}

func (f *function) Delete() error {
	err := f.resCli.Delete(f.name, f.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while deleting Function %s in namespace %s", f.name, f.namespace)
	}

	return nil
}

func (f *function) Get() error {
	fn, err := f.resCli.Get(f.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting Function %s in namespace %s", f.name, f.namespace)
	}

	return nil
}





func (f *function) Update() error {
	f.resCli.

	err := f.resCli.ResCli.Update(f.name, f.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while updating Function %s in namespace %s", f.name, f.namespace)
	}

	return nil
}

func (f *function) isFunctionReady(name string) func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		if event.Type != watch.Modified {
			return false, nil
		}
		u, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			return false, ErrInvalidDataType
		}
		if u.GetName() != name {
			return false, nil
		}

		function := serverlessv1alpha1.Function{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &function)
		if err != nil {
			return false, err
		}

		for _, condition := range function.Status.Conditions {
			if condition.Type == serverlessv1alpha1.ConditionRunning && condition.Status == corev1.ConditionTrue {
				f.log.Logf("%s is ready:\n%v", name, u)
				return true, nil
			}
		}

		f.log.Logf("%s is not ready:\n%v", name, u)
		return false, nil
	}
}
