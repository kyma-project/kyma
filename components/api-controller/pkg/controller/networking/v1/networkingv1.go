package v1

import (
	"fmt"
	"reflect"

	istioNetworkingTyped "github.com/kyma-project/kyma/components/api-controller/pkg/clients/networking.istio.io/clientset/versioned/typed/networking.istio.io/v1alpha3"
	"istio.io/istio/pkg/log"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"

	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/meta/v1"
	istioNetworkingApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/networking.istio.io/v1alpha3"
	istioNetworking "github.com/kyma-project/kyma/components/api-controller/pkg/clients/networking.istio.io/clientset/versioned"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/commons"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
)

type istioImpl struct {
	istioNetworkingInterface istioNetworking.Interface
	istioGateway             string
}

func New(i istioNetworking.Interface, istioGateway string) Interface {
	return &istioImpl{
		istioNetworkingInterface: i,
		istioGateway:             istioGateway,
	}
}

func (a *istioImpl) Create(dto *Dto) (*kymaMeta.GatewayResource, error) {

	virtualService := toIstioVirtualService(dto, a.istioGateway)

	log.Debugf("Creating virtual service: %v", virtualService)

	created, err := a.istioVirtualServiceInterface(dto.MetaDto).Create(virtualService)
	if err != nil {
		return nil, commons.HandleError(err, "error while creating virtual service")
	}

	log.Debugf("Virtual service created: %v", virtualService)

	return gatewayResourceFrom(created), nil
}

func (a *istioImpl) Update(oldDto, newDto *Dto) (*kymaMeta.GatewayResource, error) {

	newVirtualService := toIstioVirtualService(newDto, a.istioGateway)
	oldVirtualService := toIstioVirtualService(oldDto, a.istioGateway)

	log.Debugf("Trying to create or update virtual service: %v", newVirtualService)

	if a.isEqual(oldVirtualService, newVirtualService) {

		log.Debugf("Update skipped: virtual service has not changed.")
		return nil, nil
	}

	newVirtualService.ObjectMeta.ResourceVersion = oldDto.Status.Resource.Version

	log.Debugf("Updating virtual service: %v", newVirtualService)

	updated, err := a.istioVirtualServiceInterface(newDto.MetaDto).Update(newVirtualService)
	if err != nil {
		return nil, commons.HandleError(err, "error while updating virtual service")
	}

	log.Debugf("Virtual service updated: %v", updated)
	return gatewayResourceFrom(updated), nil
}

func (a *istioImpl) Delete(dto *Dto) error {

	if dto == nil {
		log.Debug("Delete skipped: no virtual service to delete.")
		return nil
	}
	return a.deleteByName(dto.MetaDto)
}

func (a *istioImpl) deleteByName(meta meta.Dto) error {

	// if there is no virtual service to delete, just skip it
	if meta.Name == "" {
		log.Debug("Delete skipped: no virtual service to delete.")
		return nil
	}
	log.Debugf("Deleting virtual service: %s", meta.Name)

	err := a.istioVirtualServiceInterface(meta).Delete(meta.Name, &k8sMeta.DeleteOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return commons.HandleError(err, "error while deleting virtual service: virtual service not found")
		}
		return commons.HandleError(err, "error while deleting virtual service")
	}

	log.Debugf("Virtual service deleted: %+v", meta.Name)
	return nil
}

func (a *istioImpl) istioVirtualServiceInterface(metaDto meta.Dto) istioNetworkingTyped.VirtualServiceInterface {
	return a.istioNetworkingInterface.NetworkingV1alpha3().VirtualServices(metaDto.Namespace)
}

func (a *istioImpl) isEqual(oldRule, newRule *istioNetworkingApi.VirtualService) bool {
	return reflect.DeepEqual(oldRule.Spec, newRule.Spec)
}

func toIstioVirtualService(dto *Dto, istioGateway string) *istioNetworkingApi.VirtualService {

	objectMeta := k8sMeta.ObjectMeta{
		Name:      dto.MetaDto.Name,
		Namespace: dto.MetaDto.Namespace,
		Labels:    dto.MetaDto.Labels,
	}

	spec := &istioNetworkingApi.VirtualServiceSpec{
		Hosts:    []string{dto.Hostname},
		Gateways: []string{istioGateway},
		Http: []*istioNetworkingApi.HTTPRoute{
			&istioNetworkingApi.HTTPRoute{
				Match: []*istioNetworkingApi.HTTPMatchRequest{
					&istioNetworkingApi.HTTPMatchRequest{
						Uri: &istioNetworkingApi.StringMatch{
							Regex: "/.*",
						},
					},
				},
				Route: []*istioNetworkingApi.DestinationWeight{
					{
						Destination: &istioNetworkingApi.Destination{
							Host: fmt.Sprintf("%s.%s.svc.cluster.local", dto.ServiceName, dto.MetaDto.Namespace),
							Port: &istioNetworkingApi.PortSelector{
								Number: uint32(dto.ServicePort),
							},
						},
					},
				},
			},
		},
	}

	return &istioNetworkingApi.VirtualService{
		ObjectMeta: objectMeta,
		Spec:       spec,
	}
}

func gatewayResourceFrom(vscv *istioNetworkingApi.VirtualService) *kymaMeta.GatewayResource {
	return &kymaMeta.GatewayResource{
		Name:    vscv.Name,
		Uid:     vscv.UID,
		Version: vscv.ResourceVersion,
	}
}
