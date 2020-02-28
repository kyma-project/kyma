package service_instance_controller

import (
	"fmt"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/gateway"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const serviceInstanceSuffix = "%s-service-instance-controller"

func InitServiceInstanceController(mgr manager.Manager, appName string, gatewayManager gateway.GatewayManager) error {
	logger := log.WithField("controller", "Service Instance")

	reconciler := NewReconciler(mgr.GetClient(), gatewayManager, logger)

	controllerName := fmt.Sprintf(serviceInstanceSuffix, appName)

	return startServiceInstanceController(controllerName, mgr, reconciler)
}

func startServiceInstanceController(controllerName string, mgr manager.Manager, reconciler ServiceInstanceReconciler) error {
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &v1beta1.ServiceInstance{}}, &handler.EnqueueRequestForObject{})
}
