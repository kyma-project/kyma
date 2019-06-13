package apicontroller

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
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
		client:   client,
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

func (svc *apiService) Create(in gqlschema.APICreateInput) (*v1alpha2.Api, error) {
	api := v1alpha2.Api{
		TypeMeta: v1.TypeMeta{
			APIVersion: "authentication.kyma-project.io/v1alpha2",
			Kind:       "API",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      in.Name,
			Namespace: in.Namespace,
		},
		Spec: v1alpha2.ApiSpec{
			Service: v1alpha2.Service{
				Name: in.ServiceName,
				Port: in.ServicePort,
			},
			Hostname: in.Hostname,
			Authentication: []v1alpha2.AuthenticationRule{
				{
					Jwt: v1alpha2.JwtAuthentication{
						JwksUri: in.JwksURI,
						Issuer:  in.Issuer,
					},
					Type: v1alpha2.AuthenticationType("JWT"),
				},
			},
			DisableIstioAuthPolicyMTLS: in.DisableIstioAuthPolicyMTLS,
			AuthenticationEnabled:      in.AuthenticationEnabled,
		},
	}

	return svc.client.GatewayV1alpha2().Apis(in.Namespace).Create(&api)
}

func (svc *apiService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *apiService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}

func (svc *apiService) Update(in gqlschema.APICreateInput) (*v1alpha2.Api, error) {

	oldApi, err := svc.Find(in.Name, in.Namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "while finding API %s", in.Name)
	}

	if oldApi == nil {
		return nil, apiErrors.NewNotFound(schema.GroupResource{
			Group:    "authentication.kyma-project.io/v1alpha2",
			Resource: "API",
		}, in.Name)
	}

	api := v1alpha2.Api{
		TypeMeta: v1.TypeMeta{
			APIVersion: "authentication.kyma-project.io/v1alpha2",
			Kind:       "API",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:            in.Name,
			Namespace:       in.Namespace,
			ResourceVersion: oldApi.ObjectMeta.ResourceVersion,
		},
		Spec: v1alpha2.ApiSpec{
			Service: v1alpha2.Service{
				Name: in.ServiceName,
				Port: in.ServicePort,
			},
			Hostname: in.Hostname,
			Authentication: []v1alpha2.AuthenticationRule{
				{
					Jwt: v1alpha2.JwtAuthentication{
						JwksUri: in.JwksURI,
						Issuer:  in.Issuer,
					},
					Type: v1alpha2.AuthenticationType("JWT"),
				},
			},
			DisableIstioAuthPolicyMTLS: in.DisableIstioAuthPolicyMTLS,
			AuthenticationEnabled:      in.AuthenticationEnabled,
		},
	}

	return svc.client.GatewayV1alpha2().Apis(in.Namespace).Update(&api)
}

func (svc *apiService) Delete(name string, namespace string) error {
	return svc.client.GatewayV1alpha2().Apis(namespace).Delete(name, nil)
}
