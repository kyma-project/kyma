package resourceskit

import (
	scv1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	acv1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	sbuv1 "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	sbuClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

type ServiceCatalogClient interface {
	CreateServiceInstance(id, siName string, service *acv1.Service) (*scv1.ServiceInstance, error)
	DeleteServiceInstance(siName string) error
	CreateServiceBinding(id, sbName, svcInstanceName, secretName string) (*scv1.ServiceBinding, error)
	DeleteServiceBinding(sbName string) error
	CreateServiceBindingUsage(sbuName, lambdaName, sbName string) (*sbuv1.ServiceBindingUsage, error)
	DeleteServiceBindingUsage(sbuName string) error
}

type serviceCatalogClient struct {
	scClient  *scClient.Clientset
	sbuClient *sbuClient.Clientset
	namespace string
}

func NewServiceCatalogClient(config *rest.Config, namespace string) (ServiceCatalogClient, error) {
	scClientSet, err := scClient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	sbuClientSet, err := sbuClient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &serviceCatalogClient{scClient: scClientSet, sbuClient: sbuClientSet, namespace: namespace}, nil
}

func (c *serviceCatalogClient) CreateServiceInstance(id, siName string, service *acv1.Service) (*scv1.ServiceInstance, error) {
	serviceInstance := &scv1.ServiceInstance{
		TypeMeta: v1.TypeMeta{Kind: "ServiceInstance", APIVersion: scv1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{
			Name:       siName,
			Namespace:  c.namespace,
			Finalizers: []string{"kubernetes-incubator/service-catalog"},
		},
		Spec: scv1.ServiceInstanceSpec{
			ExternalID: id,
			Parameters: &runtime.RawExtension{},
			ServiceClassRef: &scv1.LocalObjectReference{
				Name: service.ID,
			},
			ServicePlanRef: &scv1.LocalObjectReference{
				Name: service.ID + "-plan",
			},
			UpdateRequests: 0,
			UserInfo: &scv1.UserInfo{
				Groups: []string{
					"system:serviceaccounts",
					"system:serviceaccounts:kyma-system",
					"system:authenticated",
				},
				UID:      "",
				Username: "system:serviceaccount:kyma-system:core-console-backend-service",
			},
		},
	}

	return c.scClient.ServicecatalogV1beta1().ServiceInstances(c.namespace).Create(serviceInstance)
}

func (c *serviceCatalogClient) DeleteServiceInstance(siName string) error {
	return c.scClient.ServicecatalogV1beta1().ServiceInstances(c.namespace).Delete(siName, &v1.DeleteOptions{})
}

func (c *serviceCatalogClient) CreateServiceBinding(id, sbName, svcInstanceName, secretName string) (*scv1.ServiceBinding, error) {
	serviceBinding := &scv1.ServiceBinding{
		TypeMeta:   v1.TypeMeta{Kind: "ServiceBinding", APIVersion: scv1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: sbName, Namespace: c.namespace},
		Spec: scv1.ServiceBindingSpec{
			ExternalID: id,
			ServiceInstanceRef: scv1.LocalObjectReference{
				Name: svcInstanceName,
			},
			SecretName: secretName,
			UserInfo: &scv1.UserInfo{
				Groups: []string{
					"system:authenticated",
				},
				UID:      "",
				Username: "adimn@kyma.cx",
			},
		},
	}

	return c.scClient.ServicecatalogV1beta1().ServiceBindings(c.namespace).Create(serviceBinding)
}

func (c *serviceCatalogClient) DeleteServiceBinding(sbName string) error {
	return c.scClient.ServicecatalogV1beta1().ServiceBindings(c.namespace).Delete(sbName, &v1.DeleteOptions{})
}

func (c *serviceCatalogClient) CreateServiceBindingUsage(sbuName, lambdaName, sbName string) (*sbuv1.ServiceBindingUsage, error) {
	serviceBindingUsage := &sbuv1.ServiceBindingUsage{
		TypeMeta: v1.TypeMeta{Kind: "ServiceBindingUsage", APIVersion: scv1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{
			Name:      sbuName,
			Namespace: c.namespace,
			Labels:    map[string]string{"Function": lambdaName, "ServiceBinding": sbName},
		},
		Spec: sbuv1.ServiceBindingUsageSpec{
			Parameters: &sbuv1.Parameters{
				EnvPrefix: &sbuv1.EnvPrefix{
					Name: "",
				},
			},
			ServiceBindingRef: sbuv1.LocalReferenceByName{
				Name: sbName,
			},
			UsedBy: sbuv1.LocalReferenceByKindAndName{
				Kind: "Function",
				Name: lambdaName,
			},
		},
	}

	return c.sbuClient.ServicecatalogV1alpha1().ServiceBindingUsages(c.namespace).Create(serviceBindingUsage)
}

func (c *serviceCatalogClient) DeleteServiceBindingUsage(sbuName string) error {
	return c.sbuClient.ServicecatalogV1alpha1().ServiceBindingUsages(c.namespace).Delete(sbuName, &v1.DeleteOptions{})
}
