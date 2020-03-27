package apis

import (
	"time"

	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource"
	apiType "github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/apis/types"
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Api struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
}

func New(dynamicCli dynamic.Interface, name, namespace string, waitTimeout time.Duration, logFn func(format string, args ...interface{})) *Api {
	return &Api{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  apiType.SchemeGroupVersion.Version,
			Group:    apiType.SchemeGroupVersion.Group,
			Resource: "apis",
		}, namespace, logFn),
		name:        name,
		namespace:   namespace,
		waitTimeout: waitTimeout,
	}
}

func (api *Api) List(callbacks ...func(...interface{})) ([]*apiType.Api, error) {
	ul, err := api.resCli.List(callbacks...)
	if err != nil {
		return nil, err
	}

	var apis []*apiType.Api

	for _, u := range ul.Items {
		var res apiType.Api
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), &res)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting Api %s in namespace %s", u.GetName(), u.GetNamespace())
		}

		apis = append(apis, &res)
	}

	return apis, nil
}

func (api *Api) Delete(callbacks ...func(...interface{})) error {
	err := api.resCli.Delete(api.name, api.waitTimeout, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while deleting Api %s in namespace %s", api.name, api.namespace)
	}

	return nil
}
