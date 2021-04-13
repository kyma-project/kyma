package helpers

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
)

func CreateServiceInstance(serviceCatalogInterface servicecatalogclientset.Interface, serviceId, namespace string) error {
	serviceInstance := &scv1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceId,
		},
		Spec: scv1beta1.ServiceInstanceSpec{
			PlanReference: scv1beta1.PlanReference{
				ServiceClassExternalName: serviceId,
				ServicePlanExternalName:  "default",
			},
		},
	}
	_, err := serviceCatalogInterface.ServicecatalogV1beta1().ServiceInstances(namespace).Create(serviceInstance)
	return err
}
