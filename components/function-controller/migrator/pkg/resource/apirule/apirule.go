package apirule

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource"
	apiruleTypes "github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/apirule/types"
	apiTypes "github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/apis/types"
	"k8s.io/apimachinery/pkg/runtime"

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

var (
	servicePort uint32 = 80
)

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

func (a *APIRule) Create(api apiTypes.Api, domain string, callbacks ...func(...interface{})) error {
	rule := migrateApisToApirule(api, domain)

	_, err := a.resCli.Create(rule, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while creating APIRule %s in namespace %s", a.name, a.namespace)
	}
	return err
}

func migrateApisToApirule(api apiTypes.Api, domain string) *apiruleTypes.APIRule {
	gateway := "kyma-gateway.kyma-system.svc.cluster.local"
	var handler apiruleTypes.Handler

	if len(api.Spec.Authentication) == 0 {
		handler = apiruleTypes.Handler{
			Name:   "noop",
			Config: nil,
		}
	} else {
		handler = apiruleTypes.Handler{
			Name: "jwt",
			Config: &runtime.RawExtension{
				Raw: []byte(
					fmt.Sprintf(`{
  						 "jwks_urls": [
                                "http://dex-service.kyma-system.svc.cluster.local:5556/keys"
                            ],
                            "trusted_issuers": [
                                "https://dex.%s"
                            ]
					}`, domain),
				),
			},
		}
	}

	rule := &apiruleTypes.APIRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIRule",
			APIVersion: apiruleTypes.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      api.Name,
			Namespace: api.Namespace,
		},
		Spec: apiruleTypes.APIRuleSpec{
			Gateway: &gateway,
			Rules: []apiruleTypes.Rule{
				{
					Path:    "/.*",
					Methods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
					AccessStrategies: []*apiruleTypes.Authenticator{
						{
							Handler: &handler,
						},
					},
				},
			},
			Service: &apiruleTypes.Service{
				Name: &api.Spec.Service.Name,
				Port: &servicePort,
				Host: &api.Spec.Hostname,
			},
		},
	}
	return rule
}
