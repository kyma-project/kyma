package plugins

import (
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/heptio/velero/pkg/plugin/velero"
	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	evclientset "github.com/knative/eventing/pkg/client/clientset/versioned"
	messagingv1alpha1Client "github.com/knative/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	channelNameSuffix = "-kn-channel"
	channelKind       = "NatssChannel"
)

// SetOwnerReference is a plugin for velero to set new value of UID in metadata.ownerReference field
// based on new values of other restored objects
type SetNatsChannelOwnerReference struct {
	Log logrus.FieldLogger
}

// AppliesTo return list of resource kinds which should be handled by this plugin
func (p *SetNatsChannelOwnerReference) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedResources: []string{"service"},
		LabelSelector:     "messaging.knative.dev/role=natss-channel",
	}, nil
}

// Execute contains main logic for plugin
// nolint
func (p *SetNatsChannelOwnerReference) Execute(input *velero.RestoreItemActionExecuteInput) (
	*velero.RestoreItemActionExecuteOutput, error) {
	metadata, err := meta.Accessor(input.Item)
	if err != nil {
		return nil, err
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		p.Log.Infof("ERROR: getting cluster config: %v", err)
		return nil, err
	}
	p.Log.Infof("Setting owner reference for %s %s in namespace %s", input.Item.GetObjectKind(),
		metadata.GetName(), metadata.GetNamespace())

	evClient, err := evclientset.NewForConfig(config)
	evClient.MessagingV1alpha1()
	var messagingChannel messagingv1alpha1Client.MessagingV1alpha1Interface
	messagingChannel = evClient.MessagingV1alpha1()
	channelName := metadata.GetName() + channelNameSuffix
	p.Log.Infof("NATS Channel Owner: %s", channelName)
	natsChannel, err := messagingChannel.Channels(metadata.GetNamespace()).Get(channelName, metav1.GetOptions{})
	var natsChannelKind = v1alpha1.SchemeGroupVersion.WithKind(channelKind)

	ownerReferences := []metav1.OwnerReference{
		*metav1.NewControllerRef(natsChannel, natsChannelKind),
	}
	metadata.SetOwnerReferences(ownerReferences)

	return velero.NewRestoreItemActionExecuteOutput(input.Item), nil
}

func (p *SetNatsChannelOwnerReference) inClusterClient() (*kubernetes.Clientset, error) {
	config, err := p.inClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		p.Log.Errorf("Error in creating the clientset: %v", err.Error())
		return nil, err
	}

	return clientset, nil
}

func (p *SetNatsChannelOwnerReference) inClusterConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		p.Log.Errorf("Error getting InClusterConfig: %v", err.Error())
		return nil, err
	}
	return config, nil
}
