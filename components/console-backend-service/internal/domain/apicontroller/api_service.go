package apicontroller

import (
	"fmt"
	"github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"k8s.io/client-go/tools/cache"
)

type apiService struct {
	informer cache.SharedIndexInformer
	client   versioned.Interface
	notifier resource.Notifier
}

func newApiService(informer cache.SharedIndexInformer, client versioned.Interface) *apiService {
	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)

	return &apiService{
		informer: informer,
		client: client,
		notifier: notifier,
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

func (svc *apiService) Find(name string, namespace string) (*v1alpha2.Api, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil {
		return nil, errors.Wrapf(err, "while finding API %s", name)
	}

	if !exists {
		return nil, nil
	}

	res, ok := item.(*v1alpha2.Api)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *v1alpha2.Api", res)
	}

	return res, nil
}

func (svc *apiService) Create(name string, namespace string, hostname string, serviceName string, servicePort int, jwksUri string, issuer string, disableIstioAuthPolicyMTLS *bool, authenticationEnabled *bool) (*v1alpha2.Api, error) {
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
						JwksUri: jwksUri,
						Issuer:  issuer,
					},
					Type: v1alpha2.AuthenticationType("JWT"),
				},
			},
			DisableIstioAuthPolicyMTLS: disableIstioAuthPolicyMTLS,
			AuthenticationEnabled:      authenticationEnabled,
		},
	}

	return svc.client.GatewayV1alpha2().Apis(namespace).Create(&api)
}

func (svc *apiService) Subscribe(listener resource.Listener) {
 	svc.notifier.AddListener(listener)
}

func (svc *apiService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}

func (svc *apiService) Update(name string, namespace string, hostname string, serviceName string, servicePort int, jwksUri string, issuer string, disableIstioAuthPolicyMTLS *bool, authenticationEnabled *bool) (*v1alpha2.Api, error) {

	oldApi, err := svc.Find(name, namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "while finding API %s", name)
	}

	if oldApi == nil {
		return nil, errors.New("API not found") //fix this error
	}

	api := v1alpha2.Api{
		TypeMeta: v1.TypeMeta{
			APIVersion: "authentication.kyma-project.io/v1alpha2",
			Kind:       "API",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:   name,
			Namespace: namespace,
			ResourceVersion: oldApi.ObjectMeta.ResourceVersion,
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
						JwksUri: jwksUri,
						Issuer:  issuer,
					},
					Type: v1alpha2.AuthenticationType("JWT"),
				},
			},
			DisableIstioAuthPolicyMTLS: disableIstioAuthPolicyMTLS,
			AuthenticationEnabled:      authenticationEnabled,
		},
	}

	return svc.client.GatewayV1alpha2().Apis(namespace).Update(&api)
}

func (svc *apiService) Delete(name string, namespace string) error {
	return svc.client.GatewayV1alpha2().Apis(namespace).Delete(name, nil)
}
