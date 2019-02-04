package testkit

import (
	acV1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	ac "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	subscriptionV1 "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	subscription "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type EventingClient interface {
	CreateMapping(appName string) (*acV1.ApplicationMapping, error)
	CreateEventActivation(appName string) (*acV1.EventActivation, error)
	CreateSubscription(appName string) (*subscriptionV1.Subscription, error)
}

type eventingClient struct {
	subscriptionClientSet *subscription.Clientset
	appConnClientSet      *ac.Clientset
	namespace             string
}

func NewEventingClient(namespace string) (EventingClient, error) {
	kubeconfig := os.Getenv("KUBECONFIG")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	subscriptionClientSet, err := subscription.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	appConnClientSet, err := ac.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &eventingClient{
		subscriptionClientSet: subscriptionClientSet,
		appConnClientSet:      appConnClientSet,
		namespace:             namespace,
	}, nil
}

func (c *eventingClient) CreateMapping(appName string) (*acV1.ApplicationMapping, error) {
	am := &acV1.ApplicationMapping{
		TypeMeta:   metav1.TypeMeta{Kind: "ApplicationMapping", APIVersion: acV1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: c.namespace},
	}

	return c.appConnClientSet.ApplicationconnectorV1alpha1().ApplicationMappings(c.namespace).Create(am)
}

func (c *eventingClient) CreateEventActivation(appName string) (*acV1.EventActivation, error) {
	eaSpec := acV1.EventActivationSpec{
		DisplayName: "Commerce-events",
		SourceID:    appName,
	}

	ea := &acV1.EventActivation{
		TypeMeta:   metav1.TypeMeta{Kind: "EventActivation", APIVersion: acV1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: c.namespace},

		Spec: eaSpec,
	}

	return c.appConnClientSet.ApplicationconnectorV1alpha1().EventActivations(c.namespace).Create(ea)
}

func (c *eventingClient) CreateSubscription(appName string) (*subscriptionV1.Subscription, error) {
	subSpec := subscriptionV1.SubscriptionSpec{
		Endpoint:                      "http://e2e-lambda.production:8080",
		IncludeSubscriptionNameHeader: true,
		MaxInflight:                   400,
		PushRequestTimeoutMS:          2000,
		EventType:                     "order.created",
		EventTypeVersion:              "v1",
		SourceID:                      appName,
	}

	sub := &subscriptionV1.Subscription{
		TypeMeta:   metav1.TypeMeta{Kind: "Subscription", APIVersion: subscriptionV1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: appName, Namespace: c.namespace, Labels: map[string]string{"Function": appName}},

		SubscriptionSpec: subSpec,
	}

	return c.subscriptionClientSet.EventingV1alpha1().Subscriptions(c.namespace).Create(sub)
}
