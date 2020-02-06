/*
Copyright 2020 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package eventmesh

import (
	"fmt"

	"github.com/avast/retry-go"

	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	serviceCatalogCS "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"

	appBrokerv1alpha1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	appBroker "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appConnectorCS "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	messagingCS "knative.dev/eventing/pkg/client/clientset/versioned"
	servingCS "knative.dev/serving/pkg/client/clientset/versioned"
)

const (
	integrationNamespace = "kyma-integration"
)

type ApplicationOption func(*appconnectorv1alpha1.Application)

func CreateApplication(appConnectorInterface appConnectorCS.Interface, name string, applicationOptions ...ApplicationOption) error {
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

func WithoutInstallation() ApplicationOption {
	return func(application *appconnectorv1alpha1.Application) {
		application.Spec.SkipInstallation = true
	}
}

func WithAPIService(id, gatewayUrl string) ApplicationOption {
	return func(application *appconnectorv1alpha1.Application) {
		application.Spec.Services = append(application.Spec.Services,
			appconnectorv1alpha1.Service{
				ID:          id,
				Name:        id,
				DisplayName: "Some API",
				Description: "Application Service Class with API",
				Labels: map[string]string{
					"connected-app": "app-name",
				},
				ProviderDisplayName: "provider",
				Entries: []appconnectorv1alpha1.Entry{
					{
						Type:        "API",
						AccessLabel: "accessLabel",
						GatewayUrl:  gatewayUrl,
					},
				},
			})
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

func WaitForApplication(appConnector appConnectorCS.Interface, messaging messagingCS.Interface, serving servingCS.Interface, name string, retryOptions ...retry.Option) error {
	application, err := appConnector.ApplicationconnectorV1alpha1().Applications().Get(name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("cannt get application: %w", err)
	}
	if !application.Spec.SkipInstallation {
		if err := WaitForChannel(messaging, name, integrationNamespace, retryOptions...); err != nil {
			return fmt.Errorf("waiting for application failed: %w, application: %+v", err, application)
		}
		if err := WaitForHttpSource(serving, name, integrationNamespace, retryOptions...); err != nil {
			return fmt.Errorf("waiting for application failed: %w, application: %+v", err, application)
		}
	}
	return nil
}

func WaitForServiceInstance(serviceCatalog serviceCatalogCS.Interface, name, namespace string, retryOptions ...retry.Option) error {
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
		})
}

func WaitForChannel(messaging messagingCS.Interface, name, namespace string, retryOptions ...retry.Option) error {
	return retry.Do(
		func() error {
			ch, err := messaging.MessagingV1alpha1().Channels(namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("cannot get channel: %w", err)
			}
			if !ch.Status.IsReady() {
				return fmt.Errorf("channel %v not ready", ch.Name)
			}
			return nil
		}, retryOptions...)
}

func WaitForHttpSource(serving servingCS.Interface, name, namespace string, retryOptions ...retry.Option) error {
	return retry.Do(
		func() error {
			ksvc, err := serving.ServingV1alpha1().Services(namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("cannot get HttpSource: %w", err)
			}
			if !ksvc.Status.IsReady() {
				return fmt.Errorf("knative-service %v for HTTPSource not ready, ksvc: %+v", ksvc.Name, ksvc)
			}
			return nil
		}, retryOptions...)
}

func CreateApplicationMapping(appBroker appBroker.Interface, name, namespace string) error {
	_, err := appBroker.ApplicationconnectorV1alpha1().ApplicationMappings(namespace).Create(&appBrokerv1alpha1.ApplicationMapping{
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
