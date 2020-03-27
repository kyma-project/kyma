package function

import (
	"time"

	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Function struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
}

type Data struct {
	Body        string
	Deps        string
	Annotations map[string]string
	Labels      map[string]string
	Resources   corev1.ResourceRequirements
	EnvVar      []corev1.EnvVar
	Timeout     int32
}

func New(dynamicCli dynamic.Interface, name, namespace string, waitTimeout time.Duration, logFn func(format string, args ...interface{})) *Function {
	return &Function{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  serverlessv1alpha1.GroupVersion.Version,
			Group:    serverlessv1alpha1.GroupVersion.Group,
			Resource: "functions",
		}, namespace, logFn),
		name:        name,
		namespace:   namespace,
		waitTimeout: waitTimeout,
	}
}

func (f *Function) Create(data *Data, callbacks ...func(...interface{})) error {
	function := &serverlessv1alpha1.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: serverlessv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        f.name,
			Namespace:   f.namespace,
			Annotations: data.Annotations,
			Labels:      data.Labels,
		},
		Spec: serverlessv1alpha1.FunctionSpec{
			Function:            data.Body,
			FunctionContentType: "plaintext",
			Deps:                data.Deps,
			Size:                "L",
			Runtime:             "nodejs8",
			Env:                 data.EnvVar,
			Timeout:             data.Timeout,
		},
	}

	_, err := f.resCli.Create(function, callbacks...)

	if err != nil {
		return errors.Wrapf(err, "while creating Function %s in namespace %s", f.name, f.namespace)
	}
	return err
}

func (f *Function) Get() (*serverlessv1alpha1.Function, error) {
	u, err := f.resCli.Get(f.name)
	if err != nil {
		return nil, err
	}

	var res serverlessv1alpha1.Function
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting unstructured to function %s in namespace %s", f.name, f.namespace)
	}

	return &res, nil
}
