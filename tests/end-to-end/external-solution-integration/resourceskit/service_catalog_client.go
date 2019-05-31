package resourceskit

import (
	scv1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	sbuv1 "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	sbuClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/consts"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

type ServiceCatalogClient interface {
	CreateServiceInstance(siName, siID, serviceID string) (*scv1.ServiceInstance, error)
	DeleteServiceInstance(siName string) error
	CreateServiceBinding() (*scv1.ServiceBinding, error)
	DeleteServiceBinding() error
	CreateServiceBindingUsage() (*sbuv1.ServiceBindingUsage, error)
	DeleteServiceBindingUsage() error
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

func (c *serviceCatalogClient) CreateServiceInstance(siName, siID, serviceID string) (*scv1.ServiceInstance, error) {
	log.WithFields(log.Fields{"name": siName, "id": siID, "serviceID": serviceID}).Debug("Creating ServiceInstance")
	return c.scClient.ServicecatalogV1beta1().ServiceInstances(c.namespace).Create(&scv1.ServiceInstance{
		TypeMeta: v1.TypeMeta{Kind: "ServiceInstance", APIVersion: scv1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{
			Name:       siName,
			Namespace:  c.namespace,
			Finalizers: []string{"kubernetes-incubator/service-catalog"},
		},
		Spec: scv1.ServiceInstanceSpec{
			ExternalID: siID,
			Parameters: &runtime.RawExtension{},
			PlanReference: scv1.PlanReference{
				ServiceClassName: serviceID,
				ServicePlanName:  serviceID + "-plan",
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
	})
}

func (c *serviceCatalogClient) DeleteServiceInstance(siName string) error {
	log.WithFields(log.Fields{"ServiceInstanceName": siName}).Debug("Deleting ServiceInstance")
	return c.scClient.ServicecatalogV1beta1().ServiceInstances(c.namespace).Delete(siName, &v1.DeleteOptions{})
}

func (c *serviceCatalogClient) CreateServiceBinding() (*scv1.ServiceBinding, error) {
	log.WithFields(log.Fields{"ServiceBindingName": consts.ServiceBindingName}).Debug("Creating ServiceBinding")
	serviceBinding := &scv1.ServiceBinding{
		TypeMeta:   v1.TypeMeta{Kind: "ServiceBinding", APIVersion: scv1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: consts.ServiceBindingName, Namespace: c.namespace},
		Spec: scv1.ServiceBindingSpec{
			ExternalID: consts.ServiceBindingID,
			InstanceRef: scv1.LocalObjectReference{
				Name: consts.ServiceInstanceName,
			},
			SecretName: consts.ServiceBindingSecret,
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

func (c *serviceCatalogClient) DeleteServiceBinding() error {
	log.WithFields(log.Fields{"ServiceBindingName": consts.ServiceBindingName}).Debug("Deleting ServiceBinding")
	return c.scClient.ServicecatalogV1beta1().ServiceBindings(c.namespace).Delete(consts.ServiceBindingName, &v1.DeleteOptions{})
}

func (c *serviceCatalogClient) CreateServiceBindingUsage() (*sbuv1.ServiceBindingUsage, error) {
	log.WithFields(log.Fields{"ServiceBindingUsageName": consts.ServiceBindingUsageName}).Debug("Creating ServiceBindingUsage")
	serviceBindingUsage := &sbuv1.ServiceBindingUsage{
		TypeMeta: v1.TypeMeta{Kind: "ServiceBindingUsage", APIVersion: sbuv1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{
			Name:      consts.ServiceBindingUsageName,
			Namespace: c.namespace,
			Labels:    map[string]string{"Function": consts.AppName, "ServiceBinding": consts.ServiceBindingName},
		},
		Spec: sbuv1.ServiceBindingUsageSpec{
			Parameters: &sbuv1.Parameters{
				EnvPrefix: &sbuv1.EnvPrefix{
					Name: "",
				},
			},
			ServiceBindingRef: sbuv1.LocalReferenceByName{
				Name: consts.ServiceBindingName,
			},
			UsedBy: sbuv1.LocalReferenceByKindAndName{
				Kind: "function",
				Name: consts.AppName,
			},
		},
	}

	return c.sbuClient.ServicecatalogV1alpha1().ServiceBindingUsages(c.namespace).Create(serviceBindingUsage)
}

func (c *serviceCatalogClient) DeleteServiceBindingUsage() error {
	log.WithFields(log.Fields{"ServiceBindingUsageName": consts.ServiceBindingUsageName}).Debug("Deleting ServiceBindingUsage")
	return c.sbuClient.ServicecatalogV1alpha1().ServiceBindingUsages(c.namespace).Delete(consts.ServiceBindingUsageName, &v1.DeleteOptions{})
}
