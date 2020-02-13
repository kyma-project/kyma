package testsuite

import (
	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"time"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"

	"k8s.io/client-go/dynamic"
)

type function struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
}

func newFunction(dynamicCli dynamic.Interface, name, namespace string, waitTimeout time.Duration, logFn func(format string, args ...interface{})) *function {
	return &function{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version: serverlessv1alpha1.GroupVersion.Version,
			Group: serverlessv1alpha1.GroupVersion.Group,
			Resource: "functions",
		}, namespace, logFn),
		name:        name,
		namespace:   namespace,
		waitTimeout: waitTimeout,
	}
}

func (f *function) Create(data *functionData, callbacks ...func(...interface{})) (string, error) {
	function := &serverlessv1alpha1.Function{
		TypeMeta: v1.TypeMeta{
			Kind:       "Function",
			APIVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      f.name,
			Namespace: f.namespace,
		},
		Spec: serverlessv1alpha1.FunctionSpec{
			Function: data.Body,
			FunctionContentType: "plaintext",
			Deps: data.Deps,
			Size: "L",
			Runtime: "nodejs8",
		},
	}

	resourceVersion, err := f.resCli.Create(function, callbacks...)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while creating Function %s in namespace %s", f.name, f.namespace)
	}
	return resourceVersion, err
}

func (f *function) WaitForStatusReady(initialResourceVersion string, callbacks ...func(...interface{})) error {
	waitForStatusReady := buildWaitForStatusesRunning(f.resCli.ResCli, f.waitTimeout, f.name)
	err := waitForStatusReady(initialResourceVersion, callbacks...)
	return err
}

func(f *function) Get() (*serverlessv1alpha1.Function, error) {
	//TODO: do this :)
	return nil, nil
}

func(f *function) Delete(callbacks ...func(...interface{})) error {
	err := f.resCli.Delete(f.name, f.waitTimeout, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while deleting Function %s in namespace %s", f.name, f.namespace)
	}

	return nil
}
