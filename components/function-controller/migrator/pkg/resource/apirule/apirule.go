package apirule

import (
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource"
	apiruleTypes "github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/apirule/types"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			Version:  apiruleTypes.GroupVersion.Version,
			Group:    apiruleTypes.GroupVersion.Group,
			Resource: "apirules",
		}, namespace, logFn),
		name:        name,
		namespace:   namespace,
		waitTimeout: waitTimeout,
	}
}

func (a *APIRule) Create(name, host string, port uint32, callbacks ...func(...interface{})) error {
	gateway := "kyma-gateway.kyma-system.svc.cluster.local"

	rule := &apiruleTypes.APIRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIRule",
			APIVersion: apiruleTypes.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.name,
			Namespace: a.namespace,
		},
		Spec: apiruleTypes.APIRuleSpec{
			Gateway: &gateway,
			Rules: []apiruleTypes.Rule{
				{
					Path:    "/.*",
					Methods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
					AccessStrategies: []*apiruleTypes.Authenticator{
						{
							Handler: &apiruleTypes.Handler{
								Name:   "noop",
								Config: nil,
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

	_, err := a.resCli.Create(rule, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while creating APIRule %s in namespace %s", a.name, a.namespace)
	}
	return err
}
