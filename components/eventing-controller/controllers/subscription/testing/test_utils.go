package testing

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/avast/retry-go/v3"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/events"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	smallTimeout           = 10 * time.Second
	smallPollingInterval   = 1 * time.Second
	subscriptionNameFormat = "nats-sub-%d"
)

type TestEnsemble struct {
	testID                    int
	Cfg                       *rest.Config
	K8sClient                 client.Client
	TestEnv                   *envtest.Environment
	NatsServer                *natsserver.Server
	DefaultSubscriptionConfig env.DefaultSubscriptionConfig
	SubscriberSvc             *v1.Service
	Cancel                    context.CancelFunc
	Ctx                       context.Context
	G                         *gomega.GomegaWithT
}

type Want struct {
	K8sSubscription  []gomegatypes.GomegaMatcher
	K8sEvents        []v1.Event
	NatsSubscription []gomegatypes.GomegaMatcher
}

func UpdateSubscriptionOnK8s(ens *TestEnsemble, subscription *eventingv1alpha1.Subscription) {
	g := ens.G

	err := ens.K8sClient.Update(ens.Ctx, subscription)
	g.Expect(err).Should(gomega.BeNil())
}

func CreateSubscription(ens *TestEnsemble, subscriptionOpts ...reconcilertesting.SubscriptionOpt) *eventingv1alpha1.Subscription {
	subscriptionName := fmt.Sprintf(subscriptionNameFormat, ens.testID)
	ens.testID++
	subscription := reconcilertesting.NewSubscription(subscriptionName, ens.SubscriberSvc.Namespace, subscriptionOpts...)
	subscription = createSubscriptionInK8s(ens, subscription)
	return subscription
}

func TestSubscriptionOnK8s(ens *TestEnsemble, subscription *eventingv1alpha1.Subscription, expectations ...gomegatypes.GomegaMatcher) {
	subExpectations := append(expectations, reconcilertesting.HaveSubscriptionName(subscription.Name))
	getSubscriptionOnK8S(ens, subscription).Should(gomega.And(subExpectations...))
}

func TestEventsOnK8s(ens *TestEnsemble, expectations ...v1.Event) {
	for _, event := range expectations {
		getK8sEvents(ens).Should(reconcilertesting.HaveEvent(event))
	}
}

func ValidSinkURL(ens *TestEnsemble, additions ...string) string {
	url := reconcilertesting.ValidSinkURL(ens.SubscriberSvc.Namespace, ens.SubscriberSvc.Name)
	for _, a := range additions {
		url = fmt.Sprintf("%s%s", url, a)
	}
	return url
}

func NewUncleanEventType(ending string) string {
	return fmt.Sprintf("%s%s", reconcilertesting.OrderCreatedEventTypeNotClean, ending)
}

func NewCleanEventType(ending string) string {
	return fmt.Sprintf("%s%s", reconcilertesting.OrderCreatedEventType, ending)
}

// IsSubscriptionDeletedOnK8s checks a subscription is deleted and allows making assertions on it
func IsSubscriptionDeletedOnK8s(ens *TestEnsemble, subscription *eventingv1alpha1.Subscription) gomega.AsyncAssertion {
	g := ens.G

	return g.Eventually(func() bool {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := ens.K8sClient.Get(ens.Ctx, lookupKey, subscription); err != nil {
			return k8serrors.IsNotFound(err)
		}
		return false
	}, smallTimeout, smallPollingInterval)
}

func ConfigDefault(maxInFlightMsg int) *eventingv1alpha1.SubscriptionConfig {
	return &eventingv1alpha1.SubscriptionConfig{MaxInFlightMessages: maxInFlightMsg}
}

func ConditionInvalidSink(msg string) eventingv1alpha1.Condition {
	return eventingv1alpha1.MakeCondition(
		eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionNotActive,
		v1.ConditionFalse, msg)
}

func EventInvalidSink(msg string) v1.Event {
	return v1.Event{
		Reason:  string(events.ReasonValidationFailed),
		Message: msg,
		Type:    v1.EventTypeWarning,
	}
}

func StartTestEnv(ens *TestEnsemble) {
	g := ens.G

	var err error
	var k8sCfg *rest.Config

	err = retry.Do(func() error {
		defer func() {
			if r := recover(); r != nil {
				log.Println("panic recovered:", r)
			}
		}()

		k8sCfg, err = ens.TestEnv.Start()
		return err
	},
		retry.Delay(time.Minute),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(5),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("[%v] try failed to start testenv: %s", n, err)
			if stopErr := ens.TestEnv.Stop(); stopErr != nil {
				log.Printf("failed to stop testenv: %s", stopErr)
			}
		}),
	)

	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(k8sCfg).ToNot(gomega.BeNil())
	ens.Cfg = k8sCfg
}

func StopTestEnv(ens *TestEnsemble) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("panic recovered:", r)
		}
	}()

	if stopErr := ens.TestEnv.Stop(); stopErr != nil {
		log.Printf("failed to stop testenv: %s", stopErr)
	}
}

func StartSubscriberSvc(ens *TestEnsemble) {
	ens.SubscriberSvc = reconcilertesting.NewSubscriberSvc("test-subscriber", "test")
	createSubscriberSvcInK8s(ens)
}

func createSubscriberSvcInK8s(ens *TestEnsemble) {
	g := ens.G

	// if the namespace is not "default" create it on the cluster
	if ens.SubscriberSvc.Namespace != "default " {
		namespace := fixtureNamespace(ens.SubscriberSvc.Namespace)
		if namespace.Name != "default" {
			err := ens.K8sClient.Create(ens.Ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				g.Expect(err).ShouldNot(gomega.HaveOccurred())
			}
		}
	}

	// create subscriber svc on cluster
	err := ens.K8sClient.Create(ens.Ctx, ens.SubscriberSvc)
	g.Expect(err).Should(gomega.BeNil())
}

// createSubscriptionInK8s creates a Subscription on the K8s client of the testEnsemble. All the reconciliation
// happening will be reflected in the subscription.
func createSubscriptionInK8s(ens *TestEnsemble, subscription *eventingv1alpha1.Subscription) *eventingv1alpha1.Subscription {
	g := ens.G

	if subscription.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(subscription.Namespace)
		if namespace.Name != "default" {
			err := ens.K8sClient.Create(ens.Ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				g.Expect(err).ShouldNot(gomega.HaveOccurred())
			}
		}
	}

	// create subscription on cluster
	err := ens.K8sClient.Create(ens.Ctx, subscription)
	g.Expect(err).Should(gomega.BeNil())
	return subscription
}

func fixtureNamespace(name string) *v1.Namespace {
	namespace := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "natsNamespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return &namespace
}

//
// getSubscription fetches a subscription using the lookupKey and allows making assertions on it
func getSubscriptionOnK8S(ens *TestEnsemble, subscription *eventingv1alpha1.Subscription, intervals ...interface{}) gomega.AsyncAssertion {
	g := ens.G

	if len(intervals) == 0 {
		intervals = []interface{}{smallTimeout, smallPollingInterval}
	}
	return g.Eventually(func() *eventingv1alpha1.Subscription {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := ens.K8sClient.Get(ens.Ctx, lookupKey, subscription); err != nil {
			return &eventingv1alpha1.Subscription{}
		}
		return subscription
	}, intervals...)
}

// getK8sEvents returns all kubernetes events for the given namespace.
// The result can be used in a gomega assertion.
func getK8sEvents(ens *TestEnsemble) gomega.AsyncAssertion {
	g := ens.G

	eventList := v1.EventList{}
	return g.Eventually(func() v1.EventList {
		err := ens.K8sClient.List(ens.Ctx, &eventList, client.InNamespace(ens.SubscriberSvc.Namespace))
		if err != nil {
			return v1.EventList{}
		}
		return eventList
	}, smallTimeout, smallPollingInterval)
}
