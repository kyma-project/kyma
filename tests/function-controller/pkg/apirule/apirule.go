package apirule

import (
	"time"

	apiruleTypes "github.com/kyma-project/kyma/tests/function-controller/pkg/apirule/types"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type APIRule struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         shared.Logger
}

func New(dynamicCli dynamic.Interface, name, namespace string, waitTimeout time.Duration, log shared.Logger) *APIRule {
	return &APIRule{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  apiruleTypes.GroupVersion.Version,
			Group:    apiruleTypes.GroupVersion.Group,
			Resource: "apirules",
		}, namespace, log),
		name:        name,
		namespace:   namespace,
		waitTimeout: waitTimeout,
		log:         log,
	}
}

func (a *APIRule) Create(name, host string, port uint32) (string, error) {
	gateway := "kyma-gateway.kyma-system.svc.cluster.local"

	rule := &apiruleTypes.APIRule{
		TypeMeta: v1.TypeMeta{
			Kind:       "APIRule",
			APIVersion: apiruleTypes.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      a.name,
			Namespace: a.namespace,
		},
		Spec: apiruleTypes.APIRuleSpec{
			Gateway: &gateway,
			Rules: []apiruleTypes.Rule{
				{
					Path:    "/.*",
					Methods: []string{"GET"},
					AccessStrategies: []*apiruleTypes.Authenticator{
						{
							Handler: &apiruleTypes.Handler{
								Name: "noop",
							},
						},
					},
				},
			},
			Service: &apiruleTypes.Service{
				Name: &name,
				Port: &port,
				Host: &host,
			},
		},
	}

	resourceVersion, err := a.resCli.Create(rule)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while creating APIRule %s in namespace %s", a.name, a.namespace)
	}
	return resourceVersion, err
}

func (a *APIRule) Delete() error {
	err := a.resCli.Delete(a.name, a.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while deleting APIRule %s in namespace %s", a.name, a.namespace)
	}

	return nil
}
