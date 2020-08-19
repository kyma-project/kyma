package helpers

import (
	"fmt"

	"github.com/avast/retry-go"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	messagingv1alpha1clientset "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"

	appbrokerv1alpha1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	appbrokerclientset "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appconnectorclientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/typed/sources/v1alpha1"
)

const (
	integrationNamespace = "kyma-integration"
)

type ApplicationOption func(*appconnectorv1alpha1.Application)

func CreateApplication(appConnectorInterface appconnectorclientset.Interface, name string, applicationOptions ...ApplicationOption) error {
	application := &appconnectorv1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appconnectorv1alpha1.ApplicationSpec{
			SkipInstallation: false,
			Services:         []appconnectorv1alpha1.Service{},
		},
	}
	for _, option := range applicationOptions {
		option(application)
	}
	_, err := appConnectorInterface.ApplicationconnectorV1alpha1().Applications().Create(application)
	return err
}

func WithAccessLabel(label string) ApplicationOption {
	return func(application *appconnectorv1alpha1.Application) {
		application.Spec.AccessLabel = label
	}
}

func WithEventService(id string) ApplicationOption {
	return func(application *appconnectorv1alpha1.Application) {
		application.Spec.Services = append(application.Spec.Services,
			appconnectorv1alpha1.Service{
				ID:          id,
				Name:        id,
				DisplayName: "Some Events",
				Description: "Application Service Class with Events",
				Labels: map[string]string{
					"connected-app": "app-name",
				},
				ProviderDisplayName: "provider",
				Entries: []appconnectorv1alpha1.Entry{
					{
						Type: "Events",
					},
				},
			})
	}
}

func WaitForApplication(appConnector appconnectorclientset.Interface, messaging messagingv1alpha1clientset.MessagingV1alpha1Interface, sources sourcesv1alpha1.SourcesV1alpha1Interface, name string, retryOptions ...retry.Option) error {
	application, err := appConnector.ApplicationconnectorV1alpha1().Applications().Get(name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("cannot get application: %v", err)
	}
	if !application.Spec.SkipInstallation {
		if err := WaitForChannel(messaging, name, integrationNamespace, retryOptions...); err != nil {
			return fmt.Errorf("waiting for application failed: %v, application: %+v", err, application)
		}
		if err := WaitForHttpSource(sources, name, integrationNamespace, retryOptions...); err != nil {
			return fmt.Errorf("waiting for application failed: %v, application: %+v", err, application)
		}
	}
	return nil
}

func WaitForServiceInstance(serviceCatalog servicecatalogclientset.Interface, name, namespace string, retryOptions ...retry.Option) error {
	return retry.Do(
		func() error {
			sc, err := serviceCatalog.ServicecatalogV1beta1().ServiceInstances(namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if sc.Status.ProvisionStatus != scv1beta1.ServiceInstanceProvisionStatusProvisioned {
				return fmt.Errorf("service instance %v not provisioned yet: %v", name, sc.Status.ProvisionStatus)
			}
			return nil
		}, retryOptions...)
}

func WaitForChannel(messaging messagingv1alpha1clientset.MessagingV1alpha1Interface, name, namespace string, retryOptions ...retry.Option) error {
	return retry.Do(
		func() error {
			ch, err := messaging.Channels(namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("cannot get channel: %v", err)
			}
			if !ch.Status.IsReady() {
				return fmt.Errorf("channel %v not ready", ch.Name)
			}
			return nil
		}, retryOptions...)
}

func WaitForHttpSource(sources sourcesv1alpha1.SourcesV1alpha1Interface, name, namespace string, retryOptions ...retry.Option) error {
	return retry.Do(
		func() error {
			src, err := sources.HTTPSources(namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("cannot get HttpSource: %v", err)
			}
			if !src.Status.IsReady() {
				return fmt.Errorf("httpsource %v for HTTPSource not ready, src: %+v", src.Name, src)
			}
			return nil
		}, retryOptions...)
}

func CreateApplicationMapping(appBroker appbrokerclientset.Interface, name, namespace string) error {
	_, err := appBroker.ApplicationconnectorV1alpha1().ApplicationMappings(namespace).Create(&appbrokerv1alpha1.ApplicationMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApplicationMapping",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
	return err
}
