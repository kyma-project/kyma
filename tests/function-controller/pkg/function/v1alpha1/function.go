package v1alpha1

import (
	"context"
	"fmt"
	"net/url"
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

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

type Function struct {
	resCli      *resource.Resource
	function    *serverlessv1alpha1.Function
	FunctionURL *url.URL
	waitTimeout time.Duration
	log         *logrus.Entry
	verbose     bool
}

func NewFunction(name string, proxyEnabled bool, c shared.Container) *Function {
	function := &serverlessv1alpha1.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.Namespace,
		},
	}

	fnURL, err := helpers.GetSvcURL(name, c.Namespace, proxyEnabled)
	if err != nil {
		panic(err)
	}

	return &Function{
		resCli:      resource.New(c.DynamicCli, serverlessv1alpha1.GroupVersion.WithResource("functions"), c.Namespace, c.Log, c.Verbose),
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		verbose:     c.Verbose,
		function:    function,
		FunctionURL: fnURL,
	}
}

func (f *Function) Create(spec serverlessv1alpha1.FunctionSpec) error {
	f.function.Spec = spec
	_, err := f.resCli.Create(f.function)
	if err != nil {
		return errors.Wrapf(err, "while creating %s", f.toString)
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
	err := f.resCli.Delete(f.function.Name)
	if err != nil {
		return errors.Wrapf(err, "while deleting %s", f.toString)
	}

	return nil
}

func (f *Function) Update(spec serverlessv1alpha1.FunctionSpec) error {
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
	return errors.Wrapf(err, "while updating %s", f.toString)

}

func (f *Function) Get() (*serverlessv1alpha1.Function, error) {
	u, err := f.resCli.Get(f.function.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting %s", f.toString)
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

func (f Function) isConditionReady(fn serverlessv1alpha1.Function) bool {
	conditions := fn.Status.Conditions
	if len(conditions) == 0 {
		f.LogReadiness(false)
		return false
	}

	ready := conditions[0].Type == serverlessv1alpha1.ConditionRunning && conditions[0].Status == corev1.ConditionTrue

	f.LogReadiness(ready)

	return ready
}

func convertFromUnstructuredToFunction(u *unstructured.Unstructured) (serverlessv1alpha1.Function, error) {
	fn := serverlessv1alpha1.Function{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &fn)
	return fn, err
}

func (f Function) toString() string {
	return fmt.Sprintf("Function %s in namespace %s", f.function.Name, f.function.Namespace)
}

func (f Function) LogReadiness(ready bool) {
	if ready {
		f.log.Infof("%s is ready", f.toString())
	} else {
		f.log.Infof("%s is not ready", f.toString())
	}

	if f.verbose {
		f.log.Infof("%+v", f.function)
	}
}
