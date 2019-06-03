package apicontroller

import (
	"fmt"
	"github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"k8s.io/client-go/tools/cache"
)

type apiService struct {
	informer cache.SharedIndexInformer
	client   versioned.Interface
}

func newApiService(informer cache.SharedIndexInformer, client versioned.Interface) *apiService {
	return &apiService{
		informer: informer,
		client: client,
	}
}

func (svc *apiService) List(namespace string, serviceName *string, hostname *string) ([]*v1alpha2.Api, error) {
	items, err := svc.informer.GetIndexer().ByIndex("namespace", namespace)
	if err != nil {
		return nil, err
	}

	var apis []*v1alpha2.Api
	for _, item := range items {

		api, ok := item.(*v1alpha2.Api)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: *Api", item)
		}

		match := true
		if serviceName != nil {
			if *serviceName != api.Spec.Service.Name {
				match = false
			}
		}
		if hostname != nil {
			if *hostname != api.Spec.Hostname {
				match = false
			}
		}

		if match {
			apis = append(apis, api)
		}
	}

	return apis, nil
}

func (svc *apiService) Create(name string, namespace string, hostname string, serviceName string, servicePort int, disableIstioAuthPolicyMTLS *bool, authenticationEnabled *bool) (*v1alpha2.Api, error) {


	api := v1alpha2.Api{
		TypeMeta: v1.TypeMeta{
			APIVersion: "authentication.kyma-project.io/v1alpha2",
			Kind:       "API",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:   name,
			Namespace: namespace,
		},
		Spec: v1alpha2.ApiSpec{
			Service:                    v1alpha2.Service{
				Name: serviceName,
				Port: servicePort,
			},
			Hostname:                   hostname,
			Authentication: []v1alpha2.AuthenticationRule{
				{
					Jwt: v1alpha2.JwtAuthentication{
						JwksUri: "https://test",
						Issuer:  "test",
					},
					Type: "",
				},
			},
			DisableIstioAuthPolicyMTLS: disableIstioAuthPolicyMTLS,
			AuthenticationEnabled:      authenticationEnabled,
		},
	}

	return svc.client.GatewayV1alpha2().Apis(namespace).Create(&api)
}
