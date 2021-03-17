package process

import (
	"encoding/json"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
)

var (
	systemNamespaces = map[string]bool{
		"kyma-system":      true,
		"kyma-installer":   true,
		"knative-eventing": true,
		"istio-system":     true,
		"kube-node-lease":  true,
		"kube-system":      true,
		"natss":            true,
		"kube-public":      true,
		"kyma-integration": true,
	}
)

const (
	BackedUpConfigMapName      = "eventing-post-upgrade-backed-up-data"
	BackedUpConfigMapNamespace = "kyma-system"
	KymaIntegrationNamespace   = "kyma-integration"
)

func filterOutNamespaces(namespaces *corev1.NamespaceList, filterOutNamespaces map[string]bool) *corev1.NamespaceList {
	result := new(corev1.NamespaceList)

	for _, namespace := range namespaces.Items {
		if filterOutNamespaces[namespace.Name] {
			continue
		}
		result.Items = append(result.Items, namespace)
	}

	result.ListMeta = namespaces.ListMeta
	return result
}

func filterOutNonEventingNamespaces(namespaces *corev1.NamespaceList) *corev1.NamespaceList {
	result := new(corev1.NamespaceList)

	for _, namespace := range namespaces.Items {
		if len(namespace.Labels[knativeEventingLabelKey]) != 0 {
			result.Items = append(result.Items, namespace)
		}
	}

	result.ListMeta = namespaces.ListMeta
	return result
}

func convertRuntimeObjToJSON(triggers *eventingv1alpha1.TriggerList, subscriptions *messagingv1alpha1.SubscriptionList, channels *messagingv1alpha1.ChannelList, validators, eventServices *appsv1.DeploymentList, nonSystemNs *corev1.NamespaceList) ([]byte, error) {
	data := Data{}
	result := []byte{}
	triggersData, err := json.Marshal(*triggers)
	if err != nil {
		return result, nil
	}
	data.Triggers = runtime.RawExtension{
		Raw:    triggersData,
		Object: nil,
	}

	subscriptionsData, err := json.Marshal(*subscriptions)
	if err != nil {
		return result, nil
	}
	data.Subscriptions = runtime.RawExtension{
		Raw:    subscriptionsData,
		Object: nil,
	}

	channelsData, err := json.Marshal(*channels)
	if err != nil {
		return result, nil
	}
	data.Channels = runtime.RawExtension{
		Raw:    channelsData,
		Object: nil,
	}

	validatorsData, err := json.Marshal(*validators)
	if err != nil {
		return result, nil
	}
	data.ConnectivityValidators = runtime.RawExtension{
		Raw:    validatorsData,
		Object: nil,
	}

	eventServicesData, err := json.Marshal(*eventServices)
	if err != nil {
		return result, nil
	}
	data.EventServices = runtime.RawExtension{
		Raw:    eventServicesData,
		Object: nil,
	}

	nonSystemNsData, err := json.Marshal(*nonSystemNs)
	if err != nil {
		return result, nil
	}
	data.Namespaces = runtime.RawExtension{
		Raw:    nonSystemNsData,
		Object: nil,
	}

	result, err = json.Marshal(data)
	if err != nil {
		return []byte{}, err
	}

	return result, nil
}

func generateConfigMap(name, namespace, releaseName string, data []byte) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"release": releaseName,
			},
		},
		Data: map[string]string{
			"Data": string(data),
		},
	}
	return cm
}
