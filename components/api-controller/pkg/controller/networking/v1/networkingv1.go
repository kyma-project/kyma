package v1

import (
	"fmt"
	"reflect"

	istioNetworkingTyped "github.com/kyma-project/kyma/components/api-controller/pkg/clients/networking.istio.io/clientset/versioned/typed/networking.istio.io/v1alpha3"
	log "github.com/sirupsen/logrus"
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

	virtualService := toVirtualService(dto, a.istioGateway)

	log.Infof("Creating virtual service: %+v", virtualService)

	created, err := a.istioVirtualServiceInterface(dto.MetaDto).Create(virtualService)
	if err != nil {
		return nil, commons.HandleError(err, "error while creating virtual service")
	}

	log.Infof("Virtual service %s/%s ver: %s created.", created.Namespace, created.Name, created.ResourceVersion)

	return gatewayResourceFrom(created), nil
}

func (a *istioImpl) Update(oldDto, newDto *Dto) (*kymaMeta.GatewayResource, error) {

	newVirtualService := toVirtualService(newDto, a.istioGateway)
	oldVirtualService := toVirtualService(oldDto, a.istioGateway)

	log.Infof("Trying to create or update virtual service: %+v", newVirtualService)

	if a.isEqual(oldVirtualService, newVirtualService) {

		log.Infof("Update skipped: virtual service %s/%s has not changed.", oldVirtualService.Namespace, oldVirtualService.Name)
		return gatewayResourceFrom(oldVirtualService), nil
	}

	newVirtualService.ObjectMeta.ResourceVersion = oldDto.Status.Resource.Version

	log.Infof("Updating virtual service: %s/%s", newVirtualService.Namespace, newVirtualService.Name)

	updated, err := a.istioVirtualServiceInterface(newDto.MetaDto).Update(newVirtualService)
	if err != nil {
		return nil, commons.HandleError(err, "error while updating virtual service")
	}

	log.Infof("Virtual service %s/%s ver: %s updated.", updated.Namespace, updated.Name, updated.ResourceVersion)
	return gatewayResourceFrom(updated), nil
}

func (a *istioImpl) Delete(dto *Dto) error {

	if dto == nil {
		log.Infof("Delete skipped: no virtual service to delete for: %s/%s.", dto.MetaDto.Namespace, dto.MetaDto.Name)
		return nil
	}
	return a.deleteByName(dto.MetaDto)
}

func (a *istioImpl) deleteByName(meta meta.Dto) error {

	// if there is no virtual service to delete, just skip it
	if meta.Name == "" {
		log.Infof("Delete skipped: no virtual service to delete.")
		return nil
	}
	log.Infof("Deleting virtual service: %s/%s", meta.Namespace, meta.Name)

	err := a.istioVirtualServiceInterface(meta).Delete(meta.Name, &k8sMeta.DeleteOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return commons.HandleError(err, "error while deleting virtual service: virtual service not found")
		}
		return commons.HandleError(err, "error while deleting virtual service")
	}

	log.Infof("Virtual service deleted: %s/%s", meta.Namespace, meta.Name)
	return nil
}

func (a *istioImpl) istioVirtualServiceInterface(metaDto meta.Dto) istioNetworkingTyped.VirtualServiceInterface {
	return a.istioNetworkingInterface.NetworkingV1alpha3().VirtualServices(metaDto.Namespace)
}

func (a *istioImpl) isEqual(oldRule, newRule *istioNetworkingApi.VirtualService) bool {
	return reflect.DeepEqual(oldRule.Spec, newRule.Spec)
}

func toVirtualService(dto *Dto, istioGateway string) *istioNetworkingApi.VirtualService {

	objectMeta := k8sMeta.ObjectMeta{
		Name:            dto.MetaDto.Name,
		Namespace:       dto.MetaDto.Namespace,
		Labels:          dto.MetaDto.Labels,
		UID:             dto.Status.Resource.Uid,
		ResourceVersion: dto.Status.Resource.Version,
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
