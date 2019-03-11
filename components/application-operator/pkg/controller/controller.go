package controller

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/application"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func InitApplicationController(mgr manager.Manager, releaseManager application.ReleaseManager, appName string) error {
	reconciler := NewReconciler(mgr.GetClient(), releaseManager)

	return startApplicationController(appName, mgr, reconciler)
}

func startApplicationController(appName string, mgr manager.Manager, reconciler ApplicationReconciler) error {
	c, err := controller.New(appName, mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &v1alpha1.Application{}}, &handler.EnqueueRequestForObject{})
}
