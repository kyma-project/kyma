package function

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources"
	"github.com/kyma-project/kyma/tests/function-controller/internal/utils"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
)

type Function struct {
	resCli      *resources.Resource
	function    *serverlessv1alpha2.Function
	FunctionURL *url.URL
	waitTimeout time.Duration
	log         *logrus.Entry
	verbose     bool
}

func NewFunction(name, namespace string, proxyEnabled bool, c utils.Container) *Function {
	function := &serverlessv1alpha2.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       serverlessv1alpha2.FunctionKind,
			APIVersion: serverlessv1alpha2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	fnURL, err := utils.GetSvcURL(name, namespace, proxyEnabled)
	if err != nil {
		panic(err)
	}

	return &Function{
		resCli:      resources.New(c.DynamicCli, serverlessv1alpha2.GroupVersion.WithResource("functions"), namespace, c.Log, c.Verbose),
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		verbose:     c.Verbose,
		function:    function,
		FunctionURL: fnURL,
	}
}

func (f *Function) Create(spec serverlessv1alpha2.FunctionSpec) error {
	f.function.Spec = spec
	_, err := f.resCli.Create(f.function)
	if err != nil {
		return errors.Wrapf(err, "while creating %s", f.toString())
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
	err = resources.WaitUntilConditionSatisfied(ctx, f.resCli.ResCli, condition)
	if err == nil {
		return nil
	}
	return err
}

func (f *Function) Delete() error {
	err := f.resCli.Delete(f.function.Name)
	if err != nil {
		return errors.Wrapf(err, "while deleting %s", f.toString())
	}

	return nil
}

func (f *Function) Update(spec serverlessv1alpha2.FunctionSpec) error {
	err := utils.RetryOnConflict(utils.DefaultRetry, func() error {
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
	return errors.Wrapf(err, "while updating %s", f.toString())
}

func (f *Function) Get() (*serverlessv1alpha2.Function, error) {
	u, err := f.resCli.Get(f.function.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting %s", f.toString())
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

	out, err := utils.PrettyMarshall(fn)
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
		f.logReadiness(false)
		return false
	}

	ready := conditions[0].Type == serverlessv1alpha2.ConditionRunning && conditions[0].Status == corev1.ConditionTrue

	f.logReadiness(ready)

	return ready
}

func convertFromUnstructuredToFunction(u *unstructured.Unstructured) (serverlessv1alpha2.Function, error) {
	fn := serverlessv1alpha2.Function{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &fn)
	return fn, err
}

func (f Function) toString() string {
	return fmt.Sprintf("Function %s in namespace %s", f.function.Name, f.function.Namespace)
}

func (f Function) logReadiness(ready bool) {
	if ready {
		f.log.Infof("%s is ready", f.toString())
	} else {
		f.log.Infof("%s is not ready", f.toString())
	}

	if f.verbose {
		f.log.Infof("%+v", f.function)
	}
}
