package process

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"k8s.io/client-go/util/retry"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	appsv1 "k8s.io/api/apps/v1"

	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

var _ Step = &SaveCurrentState{}

type SaveCurrentState struct {
	name    string
	process *Process
}

func NewSaveCurrentState(p *Process) SaveCurrentState {
	return SaveCurrentState{
		name:    "Save current state to configmap",
		process: p,
	}
}

func (s SaveCurrentState) Do() error {

	// Get a list of triggers
	triggers, err := s.process.Clients.Trigger.List(corev1.NamespaceAll)
	if err != nil {
		return err
	}
	s.process.Logger.Debugf("Length of list of triggers: %d", len(triggers.Items))

	// Get a list of subscriptions
	subscriptions, err := s.process.Clients.Subscription.List(corev1.NamespaceAll)
	if err != nil {
		return err
	}
	s.process.Logger.Debugf("Length of list of subscriptions: %d", len(subscriptions.Items))

	// Get a list of channels
	channels, err := s.process.Clients.Channel.List(corev1.NamespaceAll, "")
	if err != nil {
		return err
	}
	s.process.Logger.Debugf("Length of list of channels: %d", len(channels.Items))

	brokers, err := s.process.Clients.Broker.List()
	if err != nil {
		return err
	}
	// Get a list of applications
	applications, err := s.process.Clients.Application.List()
	if err != nil {
		return err
	}
	s.process.Logger.Debugf("Length of list of applications: %d", len(applications.Items))

	eventSources, err := s.process.Clients.HttpSource.List()
	if err != nil {
		return err
	}
	s.process.Logger.Debugf("Length of list of HTTP event sources: %d", len(eventSources.Items))

	// Get a list of validator deployments
	validatorDeployments := &appsv1.DeploymentList{}
	if len(applications.Items) > 0 {
		validatorDeployments, err = s.getValidatorDeployments(applications)
		if err != nil {
			return err
		}
	}
	s.process.Logger.Debugf("Length of list of validator deployments: %d", len(channels.Items))

	// Get a list of event service deployments
	eventServiceDeployments := &appsv1.DeploymentList{}
	if len(applications.Items) > 0 {
		eventServiceDeployments, err = s.getEventServiceDeployments(applications)
		if err != nil {
			return err
		}
	}
	s.process.Logger.Debugf("Length of list of event service deployments: %d", len(channels.Items))

	// Get a list of all namespaces
	namespaces, err := s.process.Clients.Namespace.List()
	if err != nil {
		return err
	}
	// Filter out non-system namespaces
	nonSystemNamespaces := filterOutNamespaces(namespaces, systemNamespaces)
	// Filter out namespaces where eventing is not enabled
	filteredNamespaces := filterOutNonEventingNamespaces(nonSystemNamespaces)
	s.process.Logger.Debugf("Length of non-system with eventing namespaces: %d", len(filteredNamespaces.Items))

	// Save the objects to Process
	s.saveObjsToProcess(triggers,
		brokers,
		subscriptions,
		channels,
		applications,
		eventSources,
		validatorDeployments,
		eventServiceDeployments,
		filteredNamespaces)

	// Convert everything to JSON and then to data in ConfigMap
	backedUpData, err := convertRuntimeObjToJSON(triggers, subscriptions, channels, validatorDeployments, eventServiceDeployments, filteredNamespaces)
	if err != nil {
		return err
	}

	configMap := generateConfigMap(BackedUpConfigMapName, BackedUpConfigMapNamespace, s.process.ReleaseName, backedUpData)

	// Create the ConfigMap inside kyma-system
	err = retry.OnError(retry.DefaultBackoff, func(err error) bool {
		if err != nil {
			return true
		}
		return false
	}, func() (err error) {
		_, err = s.process.Clients.ConfigMap.Create(configMap)
		if err != nil {
			if k8serrors.IsAlreadyExists(err) {
				s.process.Logger.Errorf("retrying: configmap: %s in ns: %s already exists", configMap.Name, configMap.Namespace)
				delErr := s.process.Clients.ConfigMap.Delete(configMap.Namespace, configMap.Name)
				if delErr != nil {
					return errors.Wrapf(err, "failed to delete configmap: %s in ns: %s", configMap.Name, configMap.Namespace)
				}
				s.process.Logger.Infof("configmap: %s in ns: %s is deleted", configMap.Name, configMap.Namespace)
				return
			}
			return
		}
		return
	})

	if err != nil {
		return errors.Wrapf(err, "failed to create config map after retries")
	}
	s.process.Logger.Infof("Step: %s, create configmap %s/%s with current state", s.ToString(), configMap.Namespace, configMap.Name)

	return nil
}

func (s SaveCurrentState) saveObjsToProcess(triggers *eventingv1alpha1.TriggerList, brokers *eventingv1alpha1.BrokerList, subscriptions *messagingv1alpha1.SubscriptionList, channels *messagingv1alpha1.ChannelList, applications *applicationconnectorv1alpha1.ApplicationList, eventSources *sourcesv1alpha1.HTTPSourceList, validators *appsv1.DeploymentList, eventServices *appsv1.DeploymentList, namespaces *corev1.NamespaceList) {
	s.process.State.Triggers = triggers
	s.process.State.Brokers = brokers
	s.process.State.Subscriptions = subscriptions
	s.process.State.Channels = channels
	s.process.State.Applications = applications
	s.process.State.EventSources = eventSources
	s.process.State.ConnectivityValidators = validators
	s.process.State.EventServices = eventServices
	s.process.State.Namespaces = namespaces

}

func (s SaveCurrentState) getValidatorDeployments(apps *applicationconnectorv1alpha1.ApplicationList) (*appsv1.DeploymentList, error) {
	appLabels := make([]string, 0)
	labelFormat := "app=%s-connectivity-validator"
	for _, app := range apps.Items {
		appLabels = append(appLabels, fmt.Sprintf(labelFormat, app.Name))
	}
	appLabelsStr := strings.Join(appLabels, ",")
	return s.process.Clients.Deployment.List(KymaIntegrationNamespace, appLabelsStr)
}

func (s SaveCurrentState) getEventServiceDeployments(apps *applicationconnectorv1alpha1.ApplicationList) (*appsv1.DeploymentList, error) {
	appLabels := make([]string, 0)
	labelFormat := "app=%s-event-service"
	for _, app := range apps.Items {
		appLabels = append(appLabels, fmt.Sprintf(labelFormat, app.Name))
	}
	appLabelsStr := strings.Join(appLabels, ",")
	return s.process.Clients.Deployment.List(KymaIntegrationNamespace, appLabelsStr)
}

func (s SaveCurrentState) ToString() string {
	return s.name
}
