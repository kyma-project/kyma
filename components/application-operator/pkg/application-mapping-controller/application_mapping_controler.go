package application_mapping_controller

import (
	v1alpha12 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func InitApplicationMappingController(controllerName string, mgr manager.Manager, appConnClient v1alpha1.ApplicationconnectorV1alpha1Interface, deployer GatewayDeployer) error {
	reconciler := NewReconciler(appConnClient, deployer)

	return startApplicationMappingController(controllerName, mgr, reconciler)
}

func startApplicationMappingController(controllerName string, mgr manager.Manager, reconciler AppMappingReconciler) error {
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &v1alpha12.ApplicationMapping{}}, &handler.EnqueueRequestForObject{})
}
