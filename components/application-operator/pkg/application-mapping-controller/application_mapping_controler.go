package application_mapping_controller

import (
	v1alpha12 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func InitApplicationMappingController(controllerName string, mgr manager.Manager, deployer GatewayDeployer) error {
	reconciler := NewReconciler(mgr.GetClient(), deployer)

	return startApplicationMappingController(controllerName, mgr, reconciler)
}

func startApplicationMappingController(controllerName string, mgr manager.Manager, reconciler AppMappingReconciler) error {
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &v1alpha12.ApplicationMapping{}}, &handler.EnqueueRequestForObject{})
}
