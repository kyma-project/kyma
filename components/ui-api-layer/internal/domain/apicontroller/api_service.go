package apicontroller

import (
	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/v1alpha2"
	"k8s.io/client-go/tools/cache"

	"fmt"
)

type apiService struct {
	informer cache.SharedIndexInformer
}

func newApiService(informer cache.SharedIndexInformer) *apiService {
	return &apiService{
		informer: informer,
	}
}

func (svc *apiService) List(environment string, serviceName *string, hostname *string) ([]*v1alpha2.Api, error) {
	items, err := svc.informer.GetIndexer().ByIndex("namespace", environment)
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
