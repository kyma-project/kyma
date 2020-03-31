package kubeless

import (
	"time"

	kubelessv1beta1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource"

	"github.com/pkg/errors"
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

func New(dynamicCli dynamic.Interface, name, namespace string, waitTimeout time.Duration, logFn func(format string, args ...interface{})) *Function {
	return &Function{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  kubelessv1beta1.SchemeGroupVersion.Version,
			Group:    kubelessv1beta1.SchemeGroupVersion.Group,
			Resource: "functions",
		}, namespace, logFn),
		name:        name,
		waitTimeout: waitTimeout,
	}
}

func (f *Function) List(callbacks ...func(...interface{})) ([]*kubelessv1beta1.Function, error) {
	ul, err := f.resCli.List(callbacks...)
	if err != nil {
		return nil, err
	}

	var fns []*kubelessv1beta1.Function

	for _, u := range ul.Items {
		var res kubelessv1beta1.Function
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), &res)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting kubeless Function %s in namespace %s", u.GetName(), u.GetNamespace())
		}

		fns = append(fns, &res)
	}

	return fns, nil
}

func (f *Function) Delete(callbacks ...func(...interface{})) error {
	err := f.resCli.Delete(f.name, f.waitTimeout, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while deleting Function %s in namespace %s", f.name, f.namespace)
	}

	return nil
}
