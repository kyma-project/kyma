package function

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/retry"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
)

type Function struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         *logrus.Entry
	verbose     bool
}

func NewFunction(name string, c shared.Container) *Function {
	return &Function{
		resCli:      resource.New(c.DynamicCli, serverlessv1alpha2.GroupVersion.WithResource("functions"), c.Namespace, c.Log, c.Verbose),
		name:        name,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		verbose:     c.Verbose,
	}
}

func (f *Function) Create(spec serverlessv1alpha2.FunctionSpec) error {
	function := &serverlessv1alpha2.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: serverlessv1alpha2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.name,
			Namespace: f.namespace,
		},
		Spec: spec,
	}

	_, err := f.resCli.Create(function)
	if err != nil {
		return errors.Wrapf(err, "while creating Function %s in namespace %s", f.name, f.namespace)
	}
	return err
}

func (f *Function) WaitForStatusRunning() error {
	fn, err := f.Get()
	if err != nil {
		return err
	}

	if f.isConditionReady(*fn) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), f.waitTimeout)
	defer cancel()
	condition := f.isFunctionReady()
	err = resource.WaitUntilConditionSatisfied(ctx, f.resCli.ResCli, condition)
	if err == nil {
		return nil
	}
	return err
}

func (f *Function) Delete() error {
	err := f.resCli.Delete(f.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting Function %s in namespace %s", f.name, f.namespace)
	}

	return nil
}

func (f *Function) Update(spec serverlessv1alpha2.FunctionSpec) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// correct update must first perform get
		fn, err := f.Get()
		if err != nil {
			// RetryOnConflict doesn't work with wrapped errors
			// https://github.com/kubernetes/client-go/blob/9927afa2880713c4332723b7f0865adee5e63a7b/util/retry/util.go#L89-L93
			return err
		}

		fnCopy := fn.DeepCopy()
		fnCopy.Spec = spec

		_, err = f.resCli.Update(fnCopy)
		// RetryOnConflict doesn't work with wrapped errors
		// https://github.com/kubernetes/client-go/blob/9927afa2880713c4332723b7f0865adee5e63a7b/util/retry/util.go#L89-L93
		return err
	}, f.log)
	return errors.Wrapf(err, "while updating Function %s in namespace %s", f.name, f.namespace)

}

func (f *Function) Get() (*serverlessv1alpha2.Function, error) {
	u, err := f.resCli.Get(f.name)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Function %s in namespace %s", f.name, f.namespace)
	}

	function, err := convertFromUnstructuredToFunction(u)
	if err != nil {
		return nil, err
	}

	return &function, nil
}

func (f Function) LogResource() error {
	fn, err := f.Get()
	if err != nil {
		return err
	}

	out, err := helpers.PrettyMarshall(fn)
	if err != nil {
		return err
	}

	f.log.Infof("%s", out)
	return nil
}

func (f *Function) isFunctionReady() func(event watch.Event) (bool, error) {
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

func (f Function) isConditionReady(fn serverlessv1alpha2.Function) bool {
	conditions := fn.Status.Conditions
	if len(conditions) == 0 {
		shared.LogReadiness(false, f.verbose, f.name, f.log, fn)
		return false
	}

	ready := conditions[0].Type == serverlessv1alpha2.ConditionRunning && conditions[0].Status == corev1.ConditionTrue

	shared.LogReadiness(ready, f.verbose, f.name, f.log, fn)

	return ready
}

func convertFromUnstructuredToFunction(u *unstructured.Unstructured) (serverlessv1alpha2.Function, error) {
	fn := serverlessv1alpha2.Function{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &fn)
	return fn, err
}
