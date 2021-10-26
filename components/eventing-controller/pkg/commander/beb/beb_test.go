package beb

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/commander/fake"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/config"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

const (
	domain = "domain.com"
)

func TestCleanup(t *testing.T) {
	bebCommander := fake.Commander{}
	g := gomega.NewWithT(t)

	// When
	ctx := context.Background()

	// create a Kyma subscription
	subscription := controllertesting.NewSubscription("test", "test",
		controllertesting.WithWebhookAuthForBEB, controllertesting.WithFakeSubscriptionStatus, controllertesting.WithEventTypeFilter)
	subscription.Spec.Sink = "https://bla.test.svc.cluster.local"

	// create an APIRule
	apiRule := controllertesting.NewAPIRule(subscription, controllertesting.WithPath)
	controllertesting.WithService("host-test", "svc-test", apiRule)
	subscription.Status.APIRuleName = apiRule.Name

	// start BEB Mock
	bebMock := startBEBMock()
	envConf := env.Config{

		BEBAPIURL:                bebMock.MessagingURL,
		ClientID:                 "client-id",
		ClientSecret:             "client-secret",
		TokenEndpoint:            bebMock.TokenURL,
		WebhookActivationTimeout: 0,
		WebhookTokenEndpoint:     "webhook-token-endpoint",
		Domain:                   domain,
		EventTypePrefix:          controllertesting.EventTypePrefix,
		BEBNamespace:             "/default/ns",
		Qos:                      "AT_LEAST_ONCE",
	}
	credentials := &handlers.OAuth2ClientCredentials{
		ClientID:     "webhook_client_id",
		ClientSecret: "webhook_client_secret",
	}

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	g.Expect(err).To(gomega.BeNil())

	// create a BEB handler to connect to BEB Mock
	nameMapper := handlers.NewBEBSubscriptionNameMapper("mydomain.com", handlers.MaxBEBSubscriptionNameLength)
	bebHandler := handlers.NewBEB(credentials, nameMapper, defaultLogger)
	err = bebHandler.Initialize(envConf)
	g.Expect(err).To(gomega.BeNil())
	bebCommander.Backend = bebHandler

	// create fake Dynamic clients
	fakeClient, err := controllertesting.NewFakeSubscriptionClient(subscription)
	g.Expect(err).To(gomega.BeNil())
	bebCommander.Client = fakeClient

	// Create APIRule
	unstructuredAPIRule, err := controllertesting.ToUnstructuredAPIRule(apiRule)
	g.Expect(err).To(gomega.BeNil())
	unstructuredAPIRuleBeforeCleanup, err := bebCommander.Client.Resource(handlers.APIRuleGroupVersionResource()).Namespace("test").Create(ctx, unstructuredAPIRule, metav1.CreateOptions{})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(unstructuredAPIRuleBeforeCleanup).ToNot(gomega.BeNil())

	// create a BEB subscription from Kyma subscription
	fakeCleaner := fake.Cleaner{}
	_, err = bebCommander.Backend.SyncSubscription(subscription, &fakeCleaner, apiRule)
	g.Expect(err).To(gomega.BeNil())

	//  check that the susbcription exist in bebMock
	getSubscriptionURL := fmt.Sprintf(bebMock.BEBConfig.GetURLFormat, nameMapper.MapSubscriptionName(subscription))
	resp, err := http.Get(getSubscriptionURL) //nolint:gosec
	g.Expect(err).To(gomega.BeNil())
	g.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))

	// check that the Kyma subscription exists
	unstructuredSub, err := bebCommander.Client.Resource(controllertesting.SubscriptionGroupVersionResource()).Namespace("test").Get(ctx, subscription.Name, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	_, err = controllertesting.ToSubscription(unstructuredSub)
	g.Expect(err).To(gomega.BeNil())

	// check that the APIRule exists
	unstructuredAPIRuleBeforeCleanup, err = bebCommander.Client.Resource(handlers.APIRuleGroupVersionResource()).Namespace("test").Get(ctx, apiRule.Name, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(unstructuredAPIRuleBeforeCleanup).ToNot(gomega.BeNil())

	// Then
	err = cleanup(bebCommander.Backend, bebCommander.Client, defaultLogger.WithContext())
	g.Expect(err).To(gomega.BeNil())

	// Expect
	// the BEB subscription should be deleted from BEB Mock
	resp, err = http.Get(getSubscriptionURL) //nolint:gosec
	g.Expect(err).To(gomega.BeNil())
	g.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusNotFound))

	// the Kyma subscription status should be empty
	unstructuredSub, err = bebCommander.Client.Resource(controllertesting.SubscriptionGroupVersionResource()).Namespace("test").Get(ctx, subscription.Name, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	gotSub, err := controllertesting.ToSubscription(unstructuredSub)
	g.Expect(err).To(gomega.BeNil())
	expectedSubStatus := eventingv1alpha1.SubscriptionStatus{}
	g.Expect(expectedSubStatus).To(gomega.Equal(gotSub.Status))

	// the associated APIRule should be deleted
	unstructuredAPIRuleAfterCleanup, err := bebCommander.Client.Resource(handlers.APIRuleGroupVersionResource()).Namespace("test").Get(ctx, apiRule.Name, metav1.GetOptions{})
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(unstructuredAPIRuleAfterCleanup).To(gomega.BeNil())

}

func startBEBMock() *controllertesting.BEBMock {
	bebConfig := &config.Config{}
	beb := controllertesting.NewBEBMock(bebConfig)
	bebURI := beb.Start()
	tokenURL := fmt.Sprintf("%s%s", bebURI, controllertesting.TokenURLPath)
	messagingURL := fmt.Sprintf("%s%s", bebURI, controllertesting.MessagingURLPath)
	beb.TokenURL = tokenURL
	beb.MessagingURL = messagingURL
	bebConfig = config.GetDefaultConfig(messagingURL)
	beb.BEBConfig = bebConfig
	return beb
}
