package resourceskit

import (
	acV1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	ac "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	subscriptionV1 "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	subscription "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/consts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type EventingClient interface {
	CreateMapping() (*acV1.ApplicationMapping, error)
	DeleteMapping() error
	CreateEventActivation() (*acV1.EventActivation, error)
	DeleteEventActivation() error
	CreateSubscription() (*subscriptionV1.Subscription, error)
	DeleteSubscription() error
}

type eventingClient struct {
	subscriptionClientSet *subscription.Clientset
	appConnClientSet      *ac.Clientset
	namespace             string
}

func NewEventingClient(config *rest.Config, namespace string) (EventingClient, error) {
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

func (c *eventingClient) CreateMapping() (*acV1.ApplicationMapping, error) {
	am := &acV1.ApplicationMapping{
		TypeMeta:   metav1.TypeMeta{Kind: "ApplicationMapping", APIVersion: acV1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: consts.AppName, Namespace: c.namespace},
	}

	return c.appConnClientSet.ApplicationconnectorV1alpha1().ApplicationMappings(c.namespace).Create(am)
}

func (c *eventingClient) DeleteMapping() error {
	return c.appConnClientSet.ApplicationconnectorV1alpha1().ApplicationMappings(c.namespace).Delete(consts.AppName, &metav1.DeleteOptions{})
}

func (c *eventingClient) CreateEventActivation() (*acV1.EventActivation, error) {
	eaSpec := acV1.EventActivationSpec{
		DisplayName: "Commerce-events",
		SourceID:    consts.AppName,
	}

	ea := &acV1.EventActivation{
		TypeMeta:   metav1.TypeMeta{Kind: "EventActivation", APIVersion: acV1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: consts.AppName, Namespace: c.namespace},

		Spec: eaSpec,
	}

	return c.appConnClientSet.ApplicationconnectorV1alpha1().EventActivations(c.namespace).Create(ea)
}

func (c *eventingClient) DeleteEventActivation() error {
	return c.appConnClientSet.ApplicationconnectorV1alpha1().EventActivations(c.namespace).Delete(consts.AppName, &metav1.DeleteOptions{})
}

func (c *eventingClient) CreateSubscription() (*subscriptionV1.Subscription, error) {
	subSpec := subscriptionV1.SubscriptionSpec{
		Endpoint:                      consts.LambdaEndpoint,
		IncludeSubscriptionNameHeader: true,
		MaxInflight:                   400,
		PushRequestTimeoutMS:          2000,
		EventType:                     consts.EventType,
		EventTypeVersion:              consts.EventVersion,
		SourceID:                      consts.AppName,
	}

	sub := &subscriptionV1.Subscription{
		TypeMeta:   metav1.TypeMeta{Kind: "Subscription", APIVersion: subscriptionV1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: consts.AppName, Namespace: c.namespace, Labels: map[string]string{"Function": consts.AppName}},

		SubscriptionSpec: subSpec,
	}

	return c.subscriptionClientSet.EventingV1alpha1().Subscriptions(c.namespace).Create(sub)
}

func (c *eventingClient) DeleteSubscription() error {
	return c.subscriptionClientSet.EventingV1alpha1().Subscriptions(c.namespace).Delete(consts.AppName, &metav1.DeleteOptions{})
}
