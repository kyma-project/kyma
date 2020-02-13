package apirule

import (
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule/types"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"

	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type APIRule struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
}

func New(dynamicCli dynamic.Interface, name, namespace string, waitTimeout time.Duration, logFn func(format string, args ...interface{})) *APIRule {
	return &APIRule{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  types.GroupVersion.Version,
			Group:    types.GroupVersion.Group,
			Resource: "apirules",
		}, namespace, logFn),
		name:        name,
		namespace:   namespace,
		waitTimeout: waitTimeout,
	}
}

func (a *APIRule) Create(name, host string, port uint32, callbacks ...func(...interface{})) (string, error) {
	gateway := "kyma-gateway.kyma-system.svc.cluster.local"

	rule := &types.APIRule{
		TypeMeta: v1.TypeMeta{
			Kind:       "APIRule",
			APIVersion: types.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      a.name,
			Namespace: a.namespace,
		},
		Spec:
		types.APIRuleSpec{
			Gateway: &gateway,
			Rules: []types.Rule{
				{
					Path:    "/.*",
					Methods: []string{"GET"},
					AccessStrategies: []*types.Authenticator{
						{
							Handler: &types.Handler{
								Name:   "noop",
								Config: &runtime.RawExtension{},
							},
						},
					},
				},
			},
			Service: &types.Service{
				Name: &name,
				Port: &port,
				Host: &host,
			},
		},
	}

	resourceVersion, err := a.resCli.Create(rule, callbacks...)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while creating APIRule %s in namespace %s", a.name, a.namespace)
	}
	return resourceVersion, err
}

func (a *APIRule) Delete(callbacks ...func(...interface{})) error {
	err := a.resCli.Delete(a.name, a.waitTimeout, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while deleting APIRule %s in namespace %s", a.name, a.namespace)
	}

	return nil
}
