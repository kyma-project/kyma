package controller

import (
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/remote-environment-controller/pkg/kymahelm"
	restclient "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	overridesTemplate = `global:
  domainName: {{ .DomainName }}
  proxyServiceImage: {{ .ProxyServiceImage }}
  eventServiceImage: {{ .EventServiceImage }}
  eventServiceTestsImage: {{ .EventServiceTestsImage }}`
)

type OverridesData struct {
	DomainName             string
	ProxyServiceImage      string
	EventServiceImage      string
	EventServiceTestsImage string
}

func InitRemoteEnvironmentController(mgr manager.Manager, overridesData OverridesData, namespace string, appName string, tillerUrl string) error {
	overrides, err := kymahelm.ParseOverrides(overridesData, overridesTemplate)
	if err != nil {
		return err
	}

	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return err
	}

	helmClient := kymahelm.NewClient(tillerUrl)

	reClient, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		return err
	}

	reconciler := NewReconciler(mgr.GetClient(), helmClient, reClient.ApplicationconnectorV1alpha1().RemoteEnvironments(), overrides, namespace)

	return startRemoteEnvController(appName, mgr, reconciler)
}

func startRemoteEnvController(appName string, mgr manager.Manager, reconciler RemoteEnvironmentReconciler) error {
	c, err := controller.New(appName, mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &v1alpha1.RemoteEnvironment{}}, &handler.EnqueueRequestForObject{})
}
