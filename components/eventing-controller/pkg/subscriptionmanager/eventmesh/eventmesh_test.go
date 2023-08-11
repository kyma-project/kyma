package eventmesh

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"

	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	backendeventmesh "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventmesh"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/client"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

const (
	domain = "domain.com"
)

type bebSubMgrMock struct {
	Client           dynamic.Interface
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

func Test_cleanupEventMesh(t *testing.T) {
	// given
	bebSubMgr := bebSubMgrMock{}
	ctx := context.Background()

	// create a Kyma subscription
	subscription := controllertesting.NewSubscription("test", "test",
		controllertesting.WithWebhookAuthForEventMesh(),
		controllertesting.WithFakeSubscriptionStatus(),
		controllertesting.WithOrderCreatedFilter(),
	)
	subscription.Spec.Sink = "https://bla.test.svc.cluster.local"

	// create an APIRule
	apiRule := controllertesting.NewAPIRule(subscription,
		controllertesting.WithPath(),
		controllertesting.WithService("svc-test", "host-test"),
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
	}
	credentials := &backendeventmesh.OAuth2ClientCredentials{
		ClientID:     "webhook_client_id",
		ClientSecret: "webhook_client_secret",
	}

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	// create a EventMesh handler to connect to BEB Mock
	nameMapper := utils.NewBEBSubscriptionNameMapper("mydomain.com",
		backendeventmesh.MaxSubscriptionNameLength)
	eventMeshHandler := backendeventmesh.NewEventMesh(credentials, nameMapper, defaultLogger)
	err = eventMeshHandler.Initialize(envConf)
	require.NoError(t, err)
	bebSubMgr.eventMeshBackend = eventMeshHandler

	// create fake Dynamic clients
	fakeClient, err := controllertesting.NewFakeSubscriptionClient(subscription)
	require.NoError(t, err)
	bebSubMgr.Client = fakeClient

	// Create APIRule
	unstructuredAPIRule, err := controllertesting.ToUnstructuredAPIRule(apiRule)
	require.NoError(t, err)
	unstructuredAPIRuleBeforeCleanup, err := bebSubMgr.Client.Resource(utils.APIRuleGroupVersionResource()).Namespace(
		"test").Create(ctx, unstructuredAPIRule, metav1.CreateOptions{})
	require.NoError(t, err)
	require.NotNil(t, unstructuredAPIRuleBeforeCleanup)

	// create an EventMesh subscription from Kyma subscription
	eventMeshCleaner := cleaner.NewEventMeshCleaner(defaultLogger)
	_, err = bebSubMgr.eventMeshBackend.SyncSubscription(subscription, eventMeshCleaner, apiRule)
	require.NoError(t, err)

	// check that the subscription exist in bebMock
	getSubscriptionURL := fmt.Sprintf(client.GetURLFormat, nameMapper.MapSubscriptionName(subscription.Name,
		subscription.Namespace))
	getSubscriptionURL = bebMock.MessagingURL + getSubscriptionURL
	resp, err := http.Get(getSubscriptionURL)
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, http.StatusOK)

	// check that the Kyma subscription exists
	unstructuredSub, err := bebSubMgr.Client.Resource(controllertesting.SubscriptionGroupVersionResource()).Namespace(
		"test").Get(ctx, subscription.Name, metav1.GetOptions{})
	require.NoError(t, err)
	_, err = controllertesting.ToSubscription(unstructuredSub)
	require.NoError(t, err)

	// check that the APIRule exists
	unstructuredAPIRuleBeforeCleanup, err = bebSubMgr.Client.Resource(utils.APIRuleGroupVersionResource()).Namespace(
		"test").Get(ctx, apiRule.Name, metav1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, unstructuredAPIRuleBeforeCleanup)

	// when
	err = cleanupEventMesh(bebSubMgr.eventMeshBackend, bebSubMgr.Client, defaultLogger.WithContext())
	require.NoError(t, err)

	// then
	// the BEB subscription should be deleted from BEB Mock
	resp, err = http.Get(getSubscriptionURL)
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, http.StatusNotFound)

	// the Kyma subscription status should be empty
	unstructuredSub, err = bebSubMgr.Client.Resource(controllertesting.SubscriptionGroupVersionResource()).Namespace(
		"test").Get(ctx, subscription.Name, metav1.GetOptions{})
	require.NoError(t, err)
	gotSub, err := controllertesting.ToSubscription(unstructuredSub)
	require.NoError(t, err)
	expectedSubStatus := eventingv1alpha2.SubscriptionStatus{Types: []eventingv1alpha2.EventType{}}
	require.Equal(t, expectedSubStatus, gotSub.Status)

	// the associated APIRule should be deleted
	unstructuredAPIRuleAfterCleanup, err := bebSubMgr.Client.Resource(utils.APIRuleGroupVersionResource()).Namespace(
		"test").Get(ctx, apiRule.Name, metav1.GetOptions{})
	require.Error(t, err)
	require.Nil(t, unstructuredAPIRuleAfterCleanup)
	bebMock.Stop()
}

func Test_markAllV1Alpha2SubscriptionsAsNotReady(t *testing.T) {
	// given
	ctx := context.Background()

	// create a Kyma subscription
	subscription := controllertesting.NewSubscription("test", "test",
		controllertesting.WithDefaultSource(),
		controllertesting.WithOrderCreatedFilter(),
		controllertesting.WithStatus(true),
	)

	// set hashes
	const (
		ev2Hash            = int64(6118518533334734626)
		eventMeshHash      = int64(6748405436686967274)
		webhookAuthHash    = int64(6118518533334734627)
		eventMeshLocalHash = int64(6883494500014499539)
	)
	subscription.Status.Backend.Ev2hash = ev2Hash
	subscription.Status.Backend.EventMeshHash = eventMeshHash
	subscription.Status.Backend.WebhookAuthHash = webhookAuthHash
	subscription.Status.Backend.EventMeshLocalHash = eventMeshLocalHash

	// create logger
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(t, err)

	// create fake k8s dynamic client
	fakeClient, err := controllertesting.NewFakeSubscriptionClient(subscription)
	require.NoError(t, err)

	// verify that the subscription status is ready
	unstructuredSub, err := fakeClient.Resource(controllertesting.SubscriptionGroupVersionResource()).Namespace(
		"test").Get(ctx, subscription.Name, metav1.GetOptions{})
	require.NoError(t, err)
	gotSub, err := controllertesting.ToSubscription(unstructuredSub)
	require.NoError(t, err)
	require.Equal(t, true, gotSub.Status.Ready)

	// when
	err = markAllV1Alpha2SubscriptionsAsNotReady(fakeClient, defaultLogger.WithContext())
	require.NoError(t, err)

	// then
	unstructuredSub, err = fakeClient.Resource(controllertesting.SubscriptionGroupVersionResource()).Namespace(
		"test").Get(ctx, subscription.Name, metav1.GetOptions{})
	require.NoError(t, err)
	gotSub, err = controllertesting.ToSubscription(unstructuredSub)
	require.NoError(t, err)
	require.Equal(t, false, gotSub.Status.Ready)

	// ensure hashes are preserved
	require.Equal(t, ev2Hash, gotSub.Status.Backend.Ev2hash)
	require.Equal(t, eventMeshHash, gotSub.Status.Backend.EventMeshHash)
	require.Equal(t, webhookAuthHash, gotSub.Status.Backend.WebhookAuthHash)
	require.Equal(t, eventMeshLocalHash, gotSub.Status.Backend.EventMeshLocalHash)
}

func startBEBMock() *controllertesting.EventMeshMock {
	// TODO(k15r): FIX THIS HACK
	// this is a very evil hack for the time being, until we refactored the config properly
	// it sets the URLs to relative paths, that can easily be used in the mux.
	b := controllertesting.NewEventMeshMock()
	b.Start()
	return b
}
