package beb

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/stretchr/testify/require"

	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	backendbeb "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/beb"
	backendeventmesh "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventmesh"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/client"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	controllertestingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
)

const (
	domain = "domain.com"
)

type bebSubMgrMock struct {
	Client           dynamic.Interface
	bebBackend       backendbeb.Backend
	eventMeshBackend backendeventmesh.Backend
}

func (c *bebSubMgrMock) Init(_ manager.Manager) error {
	return nil
}

func (c *bebSubMgrMock) Start(_ env.DefaultSubscriptionConfig, _ subscriptionmanager.Params) error {
	return nil
}

func (c *bebSubMgrMock) Stop(_ bool) error {
	return nil
}

func TestCleanup(t *testing.T) {
	bebSubMgr := bebSubMgrMock{}
	g := gomega.NewWithT(t)

	// When
	ctx := context.Background()

	// create a Kyma subscription
	subscription := controllertesting.NewSubscription("test", "test",
		controllertesting.WithWebhookAuthForBEB(),
		controllertesting.WithFakeSubscriptionStatus(),
		controllertesting.WithOrderCreatedFilter(),
	)
	subscription.Spec.Sink = "https://bla.test.svc.cluster.local"

	// create an APIRule
	apiRule := controllertesting.NewAPIRule(subscription,
		controllertesting.WithPath(),
		controllertesting.WithService("svc-test", "host-test"),
	)
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
		Qos:                      string(types.QosAtLeastOnce),
	}
	credentials := &backendbeb.OAuth2ClientCredentials{
		ClientID:     "webhook_client_id",
		ClientSecret: "webhook_client_secret",
	}

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	g.Expect(err).To(gomega.BeNil())

	// create a BEB handler to connect to BEB Mock
	nameMapper := utils.NewBEBSubscriptionNameMapper("mydomain.com", backendbeb.MaxBEBSubscriptionNameLength)
	bebHandler := backendbeb.NewBEB(credentials, nameMapper, defaultLogger)
	err = bebHandler.Initialize(envConf)
	g.Expect(err).To(gomega.BeNil())
	bebSubMgr.bebBackend = bebHandler

	// create fake Dynamic clients
	fakeClient, err := controllertesting.NewFakeSubscriptionClient(subscription)
	g.Expect(err).To(gomega.BeNil())
	bebSubMgr.Client = fakeClient

	// Create APIRule
	unstructuredAPIRule, err := controllertesting.ToUnstructuredAPIRule(apiRule)
	g.Expect(err).To(gomega.BeNil())
	unstructuredAPIRuleBeforeCleanup, err := bebSubMgr.Client.Resource(utils.APIRuleGroupVersionResource()).Namespace("test").Create(ctx, unstructuredAPIRule, metav1.CreateOptions{})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(unstructuredAPIRuleBeforeCleanup).ToNot(gomega.BeNil())

	// create a BEB subscription from Kyma subscription
	bebCleaner := func(et string) (string, error) {
		return et, nil
	}
	_, err = bebSubMgr.bebBackend.SyncSubscription(subscription, eventtype.CleanerFunc(bebCleaner), apiRule)
	g.Expect(err).To(gomega.BeNil())

	//  check that the susbcription exist in bebMock
	getSubscriptionURL := fmt.Sprintf(client.GetURLFormat, nameMapper.MapSubscriptionName(subscription.Name, subscription.Namespace))
	getSubscriptionURL = bebMock.MessagingURL + getSubscriptionURL
	resp, err := http.Get(getSubscriptionURL) //nolint:gosec
	g.Expect(err).To(gomega.BeNil())
	g.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusOK))

	// check that the Kyma subscription exists
	unstructuredSub, err := bebSubMgr.Client.Resource(controllertesting.SubscriptionGroupVersionResource()).Namespace("test").Get(ctx, subscription.Name, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	_, err = controllertesting.ToSubscription(unstructuredSub)
	g.Expect(err).To(gomega.BeNil())

	// check that the APIRule exists
	unstructuredAPIRuleBeforeCleanup, err = bebSubMgr.Client.Resource(utils.APIRuleGroupVersionResource()).Namespace("test").Get(ctx, apiRule.Name, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(unstructuredAPIRuleBeforeCleanup).ToNot(gomega.BeNil())

	// Then
	err = cleanup(bebSubMgr.bebBackend, bebSubMgr.Client, defaultLogger.WithContext())
	g.Expect(err).To(gomega.BeNil())

	// Expect
	// the BEB subscription should be deleted from BEB Mock
	resp, err = http.Get(getSubscriptionURL) //nolint:gosec
	g.Expect(err).To(gomega.BeNil())
	g.Expect(resp.StatusCode).Should(gomega.Equal(http.StatusNotFound))

	// the Kyma subscription status should be empty
	unstructuredSub, err = bebSubMgr.Client.Resource(controllertesting.SubscriptionGroupVersionResource()).Namespace("test").Get(ctx, subscription.Name, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	gotSub, err := controllertesting.ToSubscription(unstructuredSub)
	g.Expect(err).To(gomega.BeNil())
	expectedSubStatus := eventingv1alpha1.SubscriptionStatus{CleanEventTypes: []string{}}
	g.Expect(expectedSubStatus).To(gomega.Equal(gotSub.Status))

	// the associated APIRule should be deleted
	unstructuredAPIRuleAfterCleanup, err := bebSubMgr.Client.Resource(utils.APIRuleGroupVersionResource()).Namespace("test").Get(ctx, apiRule.Name, metav1.GetOptions{})
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(unstructuredAPIRuleAfterCleanup).To(gomega.BeNil())
	bebMock.Stop()
}

func Test_cleanupEventMesh(t *testing.T) {
	// given
	bebSubMgr := bebSubMgrMock{}
	ctx := context.Background()

	// create a Kyma subscription
	subscription := controllertestingv2.NewSubscription("test", "test",
		controllertestingv2.WithWebhookAuthForBEB(),
		controllertestingv2.WithFakeSubscriptionStatus(),
		controllertestingv2.WithOrderCreatedFilter(),
	)
	subscription.Spec.Sink = "https://bla.test.svc.cluster.local"

	// create an APIRule
	apiRule := controllertestingv2.NewAPIRule(subscription,
		controllertestingv2.WithPath(),
		controllertestingv2.WithService("svc-test", "host-test"),
	)
	subscription.Status.Backend.APIRuleName = apiRule.Name

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
		Qos:                      string(types.QosAtLeastOnce),
		EnableNewCRDVersion:      true,
	}
	credentials := &backendbeb.OAuth2ClientCredentials{
		ClientID:     "webhook_client_id",
		ClientSecret: "webhook_client_secret",
	}

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	// create a EventMesh handler to connect to BEB Mock
	nameMapper := utils.NewBEBSubscriptionNameMapper("mydomain.com", backendbeb.MaxBEBSubscriptionNameLength)
	eventMeshHandler := backendeventmesh.NewEventMesh(credentials, nameMapper, defaultLogger)
	err = eventMeshHandler.Initialize(envConf)
	require.NoError(t, err)
	bebSubMgr.eventMeshBackend = eventMeshHandler

	// create fake Dynamic clients
	fakeClient, err := controllertestingv2.NewFakeSubscriptionClient(subscription)
	require.NoError(t, err)
	bebSubMgr.Client = fakeClient

	// Create APIRule
	unstructuredAPIRule, err := controllertestingv2.ToUnstructuredAPIRule(apiRule)
	require.NoError(t, err)
	unstructuredAPIRuleBeforeCleanup, err := bebSubMgr.Client.Resource(utils.APIRuleGroupVersionResource()).Namespace("test").Create(ctx, unstructuredAPIRule, metav1.CreateOptions{})
	require.NoError(t, err)
	require.NotNil(t, unstructuredAPIRuleBeforeCleanup)

	// create an EventMesh subscription from Kyma subscription
	eventMeshCleaner := cleaner.NewEventMeshCleaner(defaultLogger)
	_, err = bebSubMgr.eventMeshBackend.SyncSubscription(subscription, eventMeshCleaner, apiRule)
	require.NoError(t, err)

	// check that the subscription exist in bebMock
	getSubscriptionURL := fmt.Sprintf(client.GetURLFormat, nameMapper.MapSubscriptionName(subscription.Name, subscription.Namespace))
	getSubscriptionURL = bebMock.MessagingURL + getSubscriptionURL
	resp, err := http.Get(getSubscriptionURL) //nolint:gosec
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, http.StatusOK)

	// check that the Kyma subscription exists
	unstructuredSub, err := bebSubMgr.Client.Resource(controllertestingv2.SubscriptionGroupVersionResource()).Namespace("test").Get(ctx, subscription.Name, metav1.GetOptions{})
	require.NoError(t, err)
	_, err = controllertestingv2.ToSubscription(unstructuredSub)
	require.NoError(t, err)

	// check that the APIRule exists
	unstructuredAPIRuleBeforeCleanup, err = bebSubMgr.Client.Resource(utils.APIRuleGroupVersionResource()).Namespace("test").Get(ctx, apiRule.Name, metav1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, unstructuredAPIRuleBeforeCleanup)

	// when
	err = cleanupEventMesh(bebSubMgr.eventMeshBackend, bebSubMgr.Client, defaultLogger.WithContext())
	require.NoError(t, err)

	// then
	// the BEB subscription should be deleted from BEB Mock
	resp, err = http.Get(getSubscriptionURL) //nolint:gosec
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, http.StatusNotFound)

	// the Kyma subscription status should be empty
	unstructuredSub, err = bebSubMgr.Client.Resource(controllertestingv2.SubscriptionGroupVersionResource()).Namespace("test").Get(ctx, subscription.Name, metav1.GetOptions{})
	require.NoError(t, err)
	gotSub, err := controllertestingv2.ToSubscription(unstructuredSub)
	require.NoError(t, err)
	expectedSubStatus := eventingv1alpha2.SubscriptionStatus{Types: []eventingv1alpha2.EventType{}}
	require.Equal(t, expectedSubStatus, gotSub.Status)

	// the associated APIRule should be deleted
	unstructuredAPIRuleAfterCleanup, err := bebSubMgr.Client.Resource(utils.APIRuleGroupVersionResource()).Namespace("test").Get(ctx, apiRule.Name, metav1.GetOptions{})
	require.Error(t, err)
	require.Nil(t, unstructuredAPIRuleAfterCleanup)
	bebMock.Stop()
}

func Test_markAllV1Alpha2SubscriptionsAsNotReady(t *testing.T) {
	// given
	ctx := context.Background()

	// create a Kyma subscription
	subscription := controllertestingv2.NewSubscription("test", "test",
		controllertestingv2.WithDefaultSource(),
		controllertestingv2.WithOrderCreatedFilter(),
		controllertestingv2.WithStatus(true),
	)

	// create logger
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	// create fake k8s dynamic client
	fakeClient, err := controllertestingv2.NewFakeSubscriptionClient(subscription)
	require.NoError(t, err)

	// verify that the subscription status is ready
	unstructuredSub, err := fakeClient.Resource(controllertestingv2.SubscriptionGroupVersionResource()).Namespace("test").Get(ctx, subscription.Name, metav1.GetOptions{})
	require.NoError(t, err)
	gotSub, err := controllertestingv2.ToSubscription(unstructuredSub)
	require.NoError(t, err)
	require.Equal(t, true, gotSub.Status.Ready)

	// when
	err = markAllV1Alpha2SubscriptionsAsNotReady(fakeClient, defaultLogger.WithContext())
	require.NoError(t, err)

	// then
	unstructuredSub, err = fakeClient.Resource(controllertestingv2.SubscriptionGroupVersionResource()).Namespace("test").Get(ctx, subscription.Name, metav1.GetOptions{})
	require.NoError(t, err)
	gotSub, err = controllertestingv2.ToSubscription(unstructuredSub)
	require.NoError(t, err)
	require.Equal(t, false, gotSub.Status.Ready)
}

func startBEBMock() *controllertesting.BEBMock {
	// TODO(k15r): FIX THIS HACK
	// this is a very evil hack for the time being, until we refactored the config properly
	// it sets the URLs to relative paths, that can easily be used in the mux.
	b := controllertesting.NewBEBMock()
	b.Start()
	return b
}
