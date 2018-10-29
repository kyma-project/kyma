package v1

import (
	"fmt"
	"reflect"

	istioNetworkingTyped "github.com/kyma-project/kyma/components/api-controller/pkg/clients/networking.istio.io/clientset/versioned/typed/networking.istio.io/v1alpha3"
	log "github.com/sirupsen/logrus"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	k8sClient "k8s.io/client-go/kubernetes"

	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/meta/v1"
	istioNetworkingApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/networking.istio.io/v1alpha3"
	istioNetworking "github.com/kyma-project/kyma/components/api-controller/pkg/clients/networking.istio.io/clientset/versioned"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/commons"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
)

type istioImpl struct {
	istioNetworkingInterface istioNetworking.Interface
	kubernetesInterface      k8sClient.Interface
	istioGateway             string
}

func New(i istioNetworking.Interface, k k8sClient.Interface, istioGateway string) Interface {
	return &istioImpl{
		istioNetworkingInterface: i,
		kubernetesInterface:      k,
		istioGateway:             istioGateway,
	}
}

type HostnameNotAvailableError struct {
	Hostname string
}

func (e HostnameNotAvailableError) Error() string {
	return fmt.Sprintf("the hostname %s is already in use", e.Hostname)
}

func (a *istioImpl) Create(dto *Dto) (*kymaMeta.GatewayResource, error) {

	virtualService := toVirtualService(dto, a.istioGateway)

	log.Infof("Creating virtual service: %+v", virtualService)

	isHostnameAvailable, err := a.isHostnameAvailable(dto.MetaDto, dto.Hostname, dto.Status.Resource.Uid)
	if err != nil {
		return nil, commons.HandleError(err, "error while creating virtual service")

	} else if !isHostnameAvailable {
		return nil, HostnameNotAvailableError{dto.Hostname}
	}

	created, err := a.istioVirtualServiceInterface(dto.MetaDto.Namespace).Create(virtualService)
	if err != nil {
		return nil, commons.HandleError(err, "error while creating virtual service")
	}

	log.Infof("Virtual service %s/%s ver: %s created.", created.Namespace, created.Name, created.ResourceVersion)

	return gatewayResourceFrom(created), nil
}

func (a *istioImpl) Update(oldDto, newDto *Dto) (*kymaMeta.GatewayResource, error) {

	newVirtualService := toVirtualService(newDto, a.istioGateway)

	log.Infof("Trying to create or update virtual service: %+v", newVirtualService)

	// If old VirtualService deosn't exists
	if oldDto.Status.Resource.Name == "" {

		createdResource, err := a.Create(newDto)
		if err != nil && createdResource == nil {
			return nil, err
		}

		log.Infof("Virtual service created: %v", createdResource)
		return createdResource, nil
	}

	oldVirtualService := toVirtualService(oldDto, a.istioGateway)

	if a.isEqual(oldVirtualService, newVirtualService) {

		log.Infof("Update skipped: virtual service %s/%s has not changed.", oldVirtualService.Namespace, oldVirtualService.Name)
		return gatewayResourceFrom(oldVirtualService), nil
	}

	// if the new hostname is different than the old one
	if newDto.Hostname != oldDto.Hostname {
		// check if new hostname is available
		isHostnameAvailable, err := a.isHostnameAvailable(newDto.MetaDto, newDto.Hostname, oldDto.Status.Resource.Uid)
		if err != nil {
			return gatewayResourceFrom(oldVirtualService), commons.HandleError(err, "error while updating virtual service")

		} else if !isHostnameAvailable {
			return gatewayResourceFrom(oldVirtualService), HostnameNotAvailableError{newDto.Hostname}
		}
	}

	newVirtualService.ObjectMeta.ResourceVersion = oldDto.Status.Resource.Version

	log.Infof("Updating virtual service: %s/%s", newVirtualService.Namespace, newVirtualService.Name)

	updated, err := a.istioVirtualServiceInterface(newDto.MetaDto.Namespace).Update(newVirtualService)
	if err != nil {
		return nil, commons.HandleError(err, "error while updating virtual service")
	}

	log.Infof("Virtual service %s/%s ver: %s updated.", updated.Namespace, updated.Name, updated.ResourceVersion)
	return gatewayResourceFrom(updated), nil
}

func (a *istioImpl) Delete(dto *Dto) error {

	if dto == nil || dto.Status.Resource.Name == "" {
		log.Infof("Delete skipped: no virtual service to delete for: %s/%s.", dto.MetaDto.Namespace, dto.MetaDto.Name)
		return nil
	}
	return a.deleteByName(dto.Status.Resource.Name, dto.MetaDto.Namespace)
}

func (a *istioImpl) deleteByName(name, namespace string) error {

	// if there is no virtual service to delete, just skip it
	if name == "" {
		log.Infof("Delete skipped: no virtual service to delete.")
		return nil
	}
	log.Infof("Deleting virtual service: %s/%s", namespace, name)

	err := a.istioVirtualServiceInterface(namespace).Delete(name, &k8sMeta.DeleteOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			log.Infof("Delete skipped: no virtual service to delete was found.")
			return nil
		}
		return commons.HandleError(err, "error while deleting virtual service")
	}

	log.Infof("Virtual service deleted: %s/%s", namespace, name)
	return nil
}

func (a *istioImpl) istioVirtualServiceInterface(namespace string) istioNetworkingTyped.VirtualServiceInterface {
	return a.istioNetworkingInterface.NetworkingV1alpha3().VirtualServices(namespace)
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

func (a *istioImpl) isHostnameAvailable(metaDto meta.Dto, hostname string, assignedVirtualServiceUID types.UID) (bool, error) {

	nsList, err := a.kubernetesInterface.CoreV1().Namespaces().List(k8sMeta.ListOptions{})

	if err != nil {
		return false, err
	}

	for _, ns := range nsList.Items {

		vsList, err := a.istioVirtualServiceInterface(ns.GetName()).List(k8sMeta.ListOptions{})

		if err != nil {
			return false, err
		}

		for _, vs := range vsList.Items {
			for _, occupiedHostname := range vs.Spec.Hosts {
				if occupiedHostname == hostname {
					//if hostname is used by virtualservice actually assigned to CR
					if vs.UID == assignedVirtualServiceUID {
						return true, nil
					}
					return false, nil
				}
			}
		}
	}

	return true, nil
}
