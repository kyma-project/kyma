package knative

import (
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	k8sclientset "k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	eventingclientset "knative.dev/eventing/pkg/client/clientset/versioned"
	eventingv1alpha1client "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	messagingv1alpha1client "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
)

type Client interface {
	GetChannelByLabels(ns string, labels map[string]string) (*messagingv1alpha1.Channel, error)
	GetSubscriptionByLabels(ns string, labels map[string]string) (*messagingv1alpha1.Subscription, error)
	CreateSubscription(*messagingv1alpha1.Subscription) (*messagingv1alpha1.Subscription, error)
	UpdateSubscription(*messagingv1alpha1.Subscription) (*messagingv1alpha1.Subscription, error)
	DeleteSubscription(*messagingv1alpha1.Subscription) error
	GetDefaultBroker(ns string) (*eventingv1alpha1.Broker, error)
	DeleteBroker(*eventingv1alpha1.Broker) error
	GetNamespace(name string) (*corev1.Namespace, error)
	UpdateNamespace(*corev1.Namespace) (*corev1.Namespace, error)
}

type client struct {
	messagingClient messagingv1alpha1client.MessagingV1alpha1Interface
	eventingClient  eventingv1alpha1client.EventingV1alpha1Interface
	coreClient      corev1client.CoreV1Interface
}

// compile time contract check
var _ Client = &client{}

func NewClient(knClientSet eventingclientset.Interface, k8sClientSet k8sclientset.Interface) Client {
	return &client{
		messagingClient: knClientSet.MessagingV1alpha1(),
		eventingClient:  knClientSet.EventingV1alpha1(),
		coreClient:      k8sClientSet.CoreV1(),
	}
}

// GetChannelByLabels return a knative Channel fetched via label selectors
// and based on the labels, the list of channels should have only one item.
func (c *client) GetChannelByLabels(ns string, labels map[string]string) (*messagingv1alpha1.Channel, error) {
	// check there are labels
	if len(labels) == 0 {
		return nil, errors.New("no labels provided")
	}

	// list Channels
	channels, err := c.messagingClient.Channels(ns).List(metav1.ListOptions{
		LabelSelector: k8slabels.SelectorFromSet(labels).String(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "getting Channels from cluster")
	}

	// check Channel list length to be 1
	l := len(channels.Items)
	switch {
	case l == 0:
		return nil, apierrors.NewNotFound(messagingv1alpha1.Resource("channels"), "")

	case l > 1:
		names := make([]string, l)
		for _, c := range channels.Items {
			names = append(names, c.Name)
		}

		return nil, errors.Errorf("expected 1 Channel with labels %s in namespace %s, found %d (%s)",
			labels, ns, l, names)
	}

	// return the single Channel found on the list
	return &channels.Items[0], nil
}

// GetSubscriptionByLabels return a knative Subscription fetched via label selectors
// and based on the labels, the list of subscriptions should have only one item.
func (c *client) GetSubscriptionByLabels(ns string, labels map[string]string) (*messagingv1alpha1.Subscription, error) {
	// check there are labels
	if len(labels) == 0 {
		return nil, errors.New("no labels provided")
	}

	// list Subscriptions
	subscriptions, err := c.messagingClient.Subscriptions(ns).List(metav1.ListOptions{
		LabelSelector: k8slabels.SelectorFromSet(labels).String(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "getting Subscription from cluster")
	}

	// check Subscription list length to be 1
	l := len(subscriptions.Items)
	switch {
	case l == 0:
		return nil, apierrors.NewNotFound(eventingv1alpha1.Resource("subscriptions"), "")

	case l > 1:
		names := make([]string, l)
		for _, s := range subscriptions.Items {
			names = append(names, s.Name)
		}

		return nil, errors.Errorf("expected 1 Subscription with labels %s in namespace %s, found %d (%s)",
			labels, ns, l, names)
	}

	// return the single Subscription found on the list
	return &subscriptions.Items[0], nil
}

// CreateSubscription creates the given Subscription.
func (c *client) CreateSubscription(subscription *messagingv1alpha1.Subscription) (*messagingv1alpha1.Subscription, error) {
	return c.messagingClient.Subscriptions(subscription.Namespace).Create(subscription)
}

// UpdateSubscription updates the given Subscription.
func (c *client) UpdateSubscription(subscription *messagingv1alpha1.Subscription) (*messagingv1alpha1.Subscription, error) {
	return c.messagingClient.Subscriptions(subscription.Namespace).Update(subscription)
}

// DeleteSubscription deletes the given Subscription from the cluster.
func (c *client) DeleteSubscription(s *messagingv1alpha1.Subscription) error {
	bgDeletePolicy := metav1.DeletePropagationBackground

	return c.messagingClient.
		Subscriptions(s.Namespace).
		Delete(s.Name, &metav1.DeleteOptions{PropagationPolicy: &bgDeletePolicy})
}

// GetDefaultBroker gets the default Broker in the given Namespace.
func (c *client) GetDefaultBroker(ns string) (*eventingv1alpha1.Broker, error) {
	return c.eventingClient.Brokers(ns).Get("default", metav1.GetOptions{})
}

// DeleteBroker deletes the given Broker from the cluster.
func (c *client) DeleteBroker(b *eventingv1alpha1.Broker) error {
	bgDeletePolicy := metav1.DeletePropagationBackground

	return c.eventingClient.
		Brokers(b.Namespace).
		Delete(b.Name, &metav1.DeleteOptions{PropagationPolicy: &bgDeletePolicy})
}

// GetNamespace gets a Namespace by name from the cluster.
func (c *client) GetNamespace(name string) (*corev1.Namespace, error) {
	return c.coreClient.Namespaces().Get(name, metav1.GetOptions{})
}

// UpdateNamespace updates the given Namespace in the cluster.
func (c *client) UpdateNamespace(ns *corev1.Namespace) (*corev1.Namespace, error) {
	return c.coreClient.Namespaces().Update(ns)
}
