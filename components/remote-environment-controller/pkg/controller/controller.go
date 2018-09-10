package controller

import (
	remoteenvironmentv1alpha1 "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-controller/pkg/kymahelm"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	overridesTemplate = `global:
  domainName: {{ .DomainName }}`
)

type OverridesData struct {
	DomainName string
}

func InitRemoteEnvironmentController(mgr manager.Manager, overridesData OverridesData, namespace string, appName string, tillerUrl string) error {
	overrides, err := kymahelm.ParseOverrides(overridesData, overridesTemplate)
	if err != nil {
		return err
	}

	helmClient := kymahelm.NewClient(tillerUrl)
	reconciler := NewReconciler(mgr.GetClient(), helmClient, overrides, namespace)

	return startRemoteEnvController(appName, mgr, reconciler)
}

func startRemoteEnvController(appName string, mgr manager.Manager, reconciler RemoteEnvironmentReconciler) error {
	c, err := controller.New(appName, mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &remoteenvironmentv1alpha1.RemoteEnvironment{}}, &handler.EnqueueRequestForObject{})
}
