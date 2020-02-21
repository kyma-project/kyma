package eventmesh

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
)

func CreateServiceInstance(serviceCatalogInterface servicecatalog.Interface, serviceId, namespace string) error {
	serviceInstance := &v1beta1.ServiceInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceInstance",
			APIVersion: "servicecatalog.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceId,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ServiceClassExternalName: serviceId,
				ServicePlanExternalName:  "default",
			},
		},
	}
	_, err := serviceCatalogInterface.ServicecatalogV1beta1().ServiceInstances(namespace).Create(serviceInstance)
	return err
}
