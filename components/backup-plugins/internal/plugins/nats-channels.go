package plugins

import (
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/heptio/velero/pkg/plugin/velero"
	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	evclientset "github.com/knative/eventing/pkg/client/clientset/versioned"
	messagingv1alpha1Client "github.com/knative/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	serviceNameSuffix = "-kn-channel"
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
	var messagingChannelInterface messagingv1alpha1Client.MessagingV1alpha1Interface
	messagingChannelInterface = evClient.MessagingV1alpha1()
	channelName := strings.TrimSuffix(metadata.GetName(), serviceNameSuffix)
	p.Log.Infof("NATS Channel Owner: %s", channelName)
	natsChannel, err := messagingChannelInterface.Channels(metadata.GetNamespace()).Get(channelName, metav1.GetOptions{})
	switch {
	case err == nil:
		// Exit since the NATS channel has been found
	case apierrors.IsNotFound(err):
		// there's no NATS Channel with such name
		p.Log.Errorf("Couldn't get NATS Channel %s for service: %s\n %v", channelName, metadata.GetName(), err)
		return velero.NewRestoreItemActionExecuteOutput(input.Item), nil
	default:
		return velero.NewRestoreItemActionExecuteOutput(input.Item), err
	}
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
