package beb

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/config"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	subscriptionNamespacePrefix = "test-"
	bigPollingInterval          = 3 * time.Second
	bigTimeOut                  = 40 * time.Second
	smallTimeOut                = 5 * time.Second
	smallPollingInterval        = 1 * time.Second
	domain                      = "domain.com"
)

type FakeCommander struct {
	client  dynamic.Interface
	backend *handlers.Beb
}

func (c *FakeCommander) Init(mgr manager.Manager) error {
	return nil
}

func (c *FakeCommander) Start() error {
	return nil
}

func (c *FakeCommander) Stop() error {
	return nil
}

type FakeCleaner struct {
}

func (c *FakeCleaner) Clean(eventType string) (string, error) {
	// Cleaning is not needed in this test
	return eventType, nil
}

func TestCleanup(t *testing.T) {
	bebCommander := FakeCommander{}
	g := gomega.NewWithT(t)

	// When
	ctx := context.Background()
	log := ctrl.Log.WithName("test-cleaner-beb")

	// create a Kyma subscription
	subscription := fixtureValidSubscription("test", "test")
	subscription.Status.Emshash = 0
	subscription.Status.Ev2hash = 0

	// create an APIRule
	apiRule := controllertesting.NewAPIRule(subscription, controllertesting.WithPath)
	controllertesting.WithService("foo-host", "foo-svc", apiRule)
	subscription.Status.APIRuleName = apiRule.Name

	// start BEB Mock
	bebMock := startBebMock()
	envConf := env.Config{
		BebApiUrl:                bebMock.MessagingURL,
		ClientID:                 "client-id",
		ClientSecret:             "client-secret",
		TokenEndpoint:            bebMock.TokenURL,
		WebhookActivationTimeout: 0,
		WebhookClientID:          "webhook-client-id",
		WebhookClientSecret:      "webhook-client-secret",
		WebhookTokenEndpoint:     "webhook-token-endpoint",
		Domain:                   domain,
		EventTypePrefix:          controllertesting.EventTypePrefix,
		BEBNamespace:             "/default/ns",
		Qos:                      "AT_LEAST_ONCE",
	}

	// create a BEB handler to connect to BEB Mock
	bebHandler := &handlers.Beb{Log: log}
	err := bebHandler.Initialize(envConf)
	g.Expect(err).To(gomega.BeNil())
	bebCommander.backend = bebHandler

	// create fake Dynamic clients
	fakeClient, err := NewFakeClient(subscription)
	g.Expect(err).To(gomega.BeNil())
	bebCommander.client = fakeClient

	// Create ApiRule
	unstructuredApiRule, err := toUnstructuredApiRule(apiRule)
	g.Expect(err).To(gomega.BeNil())
	unstructuredApiRuleBeforeCleanup, err := bebCommander.client.Resource(handlers.APIRuleGroupVersionResource()).Namespace("test").Create(ctx, unstructuredApiRule, metav1.CreateOptions{})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(unstructuredApiRuleBeforeCleanup).ToNot(gomega.BeNil())

	// create a BEB subscription from Kyma subscription
	fakeCleaner := FakeCleaner{}
	_, err = bebCommander.backend.SyncSubscription(subscription, &fakeCleaner, apiRule)
	g.Expect(err).To(gomega.BeNil())

	//  check that the susbcription exist in bebMock
	getSubscriptionUrl := fmt.Sprintf(bebMock.BebConfig.GetURLFormat, subscription.Name)
	resp, err := http.Get(getSubscriptionUrl)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))

	// check that the Kyma subscription exists
	unstructuredSub, err := bebCommander.client.Resource(SubscriptionGroupVersionResource()).Namespace("test").Get(ctx, subscription.Name, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	_, err = toSubscription(unstructuredSub)
	g.Expect(err).To(gomega.BeNil())

	// check that the APIRule exists
	unstructuredApiRuleBeforeCleanup, err = bebCommander.client.Resource(handlers.APIRuleGroupVersionResource()).Namespace("test").Get(ctx, apiRule.Name, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(unstructuredApiRuleBeforeCleanup).ToNot(gomega.BeNil())

	// Then
	err = cleanup(bebCommander.backend, bebCommander.client)
	g.Expect(err).To(gomega.BeNil())

	// Expect
	// the BEB subscription should be deleted from BEB Mock
	resp, err = http.Get(getSubscriptionUrl)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusNotFound))

	// the Kyma subscription status should be empty
	unstructuredSub, err = bebCommander.client.Resource(SubscriptionGroupVersionResource()).Namespace("test").Get(ctx, subscription.Name, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	gotSub, err := toSubscription(unstructuredSub)
	g.Expect(err).To(gomega.BeNil())
	expectedSubStatus := eventingv1alpha1.SubscriptionStatus{}
	g.Expect(expectedSubStatus).To(gomega.Equal(gotSub.Status))

	// the associated APIRule should be deleted
	unstructuredApiRuleAfterCleanup, err := bebCommander.client.Resource(handlers.APIRuleGroupVersionResource()).Namespace("test").Get(ctx, apiRule.Name, metav1.GetOptions{})
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(unstructuredApiRuleAfterCleanup).To(gomega.BeNil())

}

func toSubscription(unstructuredSub *unstructured.Unstructured) (*eventingv1alpha1.Subscription, error) {
	subscription := new(eventingv1alpha1.Subscription)
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredSub.Object, subscription)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

func toUnstructuredApiRule(obj interface{}) (*unstructured.Unstructured, error) {
	unstructured := &unstructured.Unstructured{}
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	unstructured.Object = unstructuredObj
	return unstructured, nil
}

type Client struct {
	Resource dynamic.NamespaceableResourceInterface
}

func NewFakeClient(sub *eventingv1alpha1.Subscription) (dynamic.Interface, error) {
	scheme, err := SetupSchemeOrDie()
	if err != nil {
		return nil, err
	}

	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, sub)
	return dynamicClient, nil
}

func SetupSchemeOrDie() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := eventingv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}

func SubscriptionGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  eventingv1alpha1.GroupVersion.Version,
		Group:    eventingv1alpha1.GroupVersion.Group,
		Resource: "subscriptions",
	}
}

func startBebMock() *controllertesting.BebMock {
	bebConfig := &config.Config{}
	beb := controllertesting.NewBebMock(bebConfig)
	bebURI := beb.Start()
	tokenURL := fmt.Sprintf("%s%s", bebURI, controllertesting.TokenURLPath)
	messagingURL := fmt.Sprintf("%s%s", bebURI, controllertesting.MessagingURLPath)
	beb.TokenURL = tokenURL
	beb.MessagingURL = messagingURL
	bebConfig = config.GetDefaultConfig(messagingURL)
	beb.BebConfig = bebConfig
	return beb
}

func fixtureValidSubscription(name, namespace string) *eventingv1alpha1.Subscription {
	return &eventingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: eventingv1alpha1.SubscriptionSpec{
			ID:       "id",
			Protocol: "BEB",
			ProtocolSettings: &eventingv1alpha1.ProtocolSettings{
				ContentMode: func() *string {
					cm := eventingv1alpha1.ProtocolSettingsContentModeBinary
					return &cm
				}(),
				ExemptHandshake: func() *bool {
					eh := true
					return &eh
				}(),
				Qos: func() *string {
					qos := "AT-LEAST_ONCE"
					return &qos
				}(),
				WebhookAuth: &eventingv1alpha1.WebhookAuth{
					Type:         "oauth2",
					GrantType:    "client_credentials",
					ClientId:     "xxx",
					ClientSecret: "xxx",
					TokenUrl:     "https://oauth2.xxx.com/oauth2/token",
					Scope:        []string{"guid-identifier"},
				},
			},
			Sink: "https://webhook.xxx.com",
			Filter: &eventingv1alpha1.BebFilters{
				Dialect: "beb",
				Filters: []*eventingv1alpha1.BebFilter{
					{
						EventSource: &eventingv1alpha1.Filter{
							Type:     "exact",
							Property: "source",
							Value:    controllertesting.EventSource,
						},
						EventType: &eventingv1alpha1.Filter{
							Type:     "exact",
							Property: "type",
							Value:    controllertesting.EventTypeNotClean,
						},
					},
				},
			},
		},
	}
}
