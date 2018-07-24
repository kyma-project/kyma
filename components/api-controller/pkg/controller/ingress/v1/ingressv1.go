package v1

import (
	"fmt"
	"reflect"

	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/meta/v1"
	k8sApiExtensions "k8s.io/api/extensions/v1beta1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/runtime"
	k8s "k8s.io/client-go/kubernetes"
)

type ingress struct {
	kubeClient k8s.Interface
}

// New returns initialized controller for ingresses
func New(kubeClient k8s.Interface) Interface {
	return &ingress{
		kubeClient: kubeClient,
	}
}

func (i *ingress) Get(dto *Dto) (k8sApiExtensions.Ingress, error) {

	ingName := ingressNameFrom(dto)
	ing, err := i.kubeClient.ExtensionsV1beta1().Ingresses(dto.MetaDto.Namespace).Get(ingName, k8sMeta.GetOptions{})

	return *ing, err
}

func (i *ingress) Create(dto *Dto) (*kymaMeta.GatewayResource, error) {

	ing := ingressFrom(dto)

	createdIng, err := i.kubeClient.ExtensionsV1beta1().Ingresses(dto.MetaDto.Namespace).Create(&ing)
	if err != nil {
		runtime.HandleError(fmt.Errorf("failed to create Ingress '%s' in namespace %s", ing.Name, dto.MetaDto.Namespace))
		return nil, err
	}

	return gatewayResourceFrom(createdIng), nil
}

func (i *ingress) Update(oldDto, newDto *Dto) (*kymaMeta.GatewayResource, error) {

	newIng := ingressFrom(newDto)
	oldIng := ingressFrom(oldDto)

	var err error
	var updatedIng *k8sApiExtensions.Ingress

	if newIng.Name != oldIng.Name {

		err := i.Delete(oldDto)
		if err != nil {
			return nil, fmt.Errorf("error while deleting old ingress. Root cause: %v", err)
		}
		createdIng, err := i.Create(newDto)
		if err != nil {
			return nil, fmt.Errorf("error while creating new ingress. Root cause: %v", err)
		}
		return createdIng, nil

	} else if !reflect.DeepEqual(newIng.Spec, oldIng.Spec) {

		updatedIng, err = i.kubeClient.ExtensionsV1beta1().Ingresses(newDto.MetaDto.Namespace).Update(&newIng)
		if err != nil {
			runtime.HandleError(fmt.Errorf("could not update Ingress '%s' in namespace '%s'. Details : %s", newIng.Name, newIng.Namespace, err.Error()))
			return nil, err
		}
		return gatewayResourceFrom(updatedIng), nil
	}
	// if there were no changes in ingress
	return nil, nil
}

func (i *ingress) Delete(dto *Dto) error {

	err := i.kubeClient.ExtensionsV1beta1().Ingresses(dto.MetaDto.Namespace).Delete(ingressNameFrom(dto), &k8sMeta.DeleteOptions{})

	if err != nil && !apiErrors.IsNotFound(err) {
		runtime.HandleError(fmt.Errorf("could not delete Ingress '%s' in namespace '%s'. Details : %s", dto.MetaDto.Name, dto.MetaDto.Namespace, err.Error()))
	}
	return err
}

func ingressNameFrom(dto *Dto) string {
	return dto.ServiceName + "-ing"
}

func ingressFrom(dto *Dto) k8sApiExtensions.Ingress {

	ing := k8sApiExtensions.Ingress{}

	ing.Labels = dto.MetaDto.Labels

	annotations := make(map[string]string)
	annotations["kubernetes.io/ingress.class"] = "istio"
	ing.Annotations = annotations

	ing.APIVersion = "v1"
	ing.Name = ingressNameFrom(dto)
	ing.Spec = k8sApiExtensions.IngressSpec{
		TLS: []k8sApiExtensions.IngressTLS{
			{
				SecretName: "istio-ingress-certs",
			},
		},
		Rules: []k8sApiExtensions.IngressRule{
			{
				Host: dto.Hostname,
				IngressRuleValue: k8sApiExtensions.IngressRuleValue{
					HTTP: &k8sApiExtensions.HTTPIngressRuleValue{
						Paths: []k8sApiExtensions.HTTPIngressPath{
							{
								Path: "/.*",
								Backend: k8sApiExtensions.IngressBackend{
									ServiceName: dto.ServiceName,
									ServicePort: intstr.FromInt(dto.ServicePort),
								},
							},
						},
					},
				},
			},
		},
	}
	return ing
}

func gatewayResourceFrom(api *k8sApiExtensions.Ingress) *kymaMeta.GatewayResource {
	return &kymaMeta.GatewayResource{
		Name:    api.Name,
		Uid:     api.UID,
		Version: api.ResourceVersion,
	}
}
