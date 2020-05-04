package testsuite

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	watchtools "k8s.io/client-go/tools/watch"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

type function struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         shared.Logger
	verbose     bool
}

func newFunction(dynamicCli dynamic.Interface, name, namespace string, waitTimeout time.Duration, log shared.Logger, verbose bool) *function {
	return &function{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  serverlessv1alpha1.GroupVersion.Version,
			Group:    serverlessv1alpha1.GroupVersion.Group,
			Resource: "functions",
		}, namespace, log, verbose),
		name:        name,
		namespace:   namespace,
		waitTimeout: waitTimeout,
		log:         log,
		verbose:     verbose,
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

func (f *function) Update(data *functionData) error {
	// correct update must first perform get
	fn, err := f.Get()
	if err != nil {
		return err
	}

	fnCopy := fn.DeepCopy()

	fnCopy.Spec.MinReplicas = &data.MinReplicas
	fnCopy.Spec.MaxReplicas = &data.MaxReplicas
	fnCopy.Spec.Deps = data.Deps
	fnCopy.Spec.Source = data.Body

	_, err = f.resCli.Update(fnCopy)
	return err
}

func (f *function) Get() (*serverlessv1alpha1.Function, error) {
	u, err := f.resCli.Get(f.name)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Function %s in namespace %s", f.name, f.namespace)
	}

	function := serverlessv1alpha1.Function{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &function)
	if err != nil {
		return nil, err
	}

	return &function, nil
}

func (f *function) isFunctionReady(name string) func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		if event.Type != watch.Modified {
			return false, nil
		}

		function, err := f.Get()
		if err != nil {
			return false, err
		}
		return f.isConditionReady(*function), nil

	}
}

func (f function) isConditionReady(fn serverlessv1alpha1.Function) bool {
	conditions := fn.Status.Conditions
	if len(conditions) == 0 {
		return false
	}

	ready := conditions[0].Type == serverlessv1alpha1.ConditionRunning && conditions[0].Status == corev1.ConditionTrue
	if ready {
		f.log.Logf("Function %s is ready", f.name)

	} else {
		f.log.Logf("Function %s is not ready", f.name)
	}

	if f.verbose {
		f.log.Logf("%+v", fn)
	}

	return ready
}
