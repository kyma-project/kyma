package process

import (
	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/application"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/broker"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/channel"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/configmap"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/deployment"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/httpsource"
	kymasubscription "github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/kyma-subscription"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/namespace"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/subscription"
	"github.com/kyma-project/kyma/components/event-sources/upgrade-job/clients/trigger"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
)

type PatchDeploymentSpec struct {
	Spec appsv1.DeploymentSpec `json:"spec"`
}

type Spec struct {
	Template Template `json:"template"`
}

type Template struct {
	Spec TemplateSpec `json:"spec"`
}

type TemplateSpec struct {
	Containers []corev1.Container `json:"containers"`
}

type Clients struct {
	Trigger          trigger.Client
	Broker           broker.Client
	Subscription     subscription.Client
	Channel          channel.Client
	Application      application.Client
	Deployment       deployment.Client
	ConfigMap        configmap.Client
	HttpSource       httpsource.Client
	Namespace        namespace.Client
	KymaSubscription kymasubscription.Client
}

type State struct {
	Triggers               *eventingv1alpha1.TriggerList
	Brokers                *eventingv1alpha1.BrokerList
	Subscriptions          *messagingv1alpha1.SubscriptionList
	Channels               *messagingv1alpha1.ChannelList
	Applications           *applicationconnectorv1alpha1.ApplicationList
	ConnectivityValidators *appsv1.DeploymentList
	EventServices          *appsv1.DeploymentList
	EventSources           *sourcesv1alpha1.HTTPSourceList
	Namespaces             *corev1.NamespaceList
}

type Data struct {
	Triggers               runtime.RawExtension
	Subscriptions          runtime.RawExtension
	Channels               runtime.RawExtension
	ConnectivityValidators runtime.RawExtension
	EventServices          runtime.RawExtension
	Namespaces             runtime.RawExtension
}

type Labels map[string]string
