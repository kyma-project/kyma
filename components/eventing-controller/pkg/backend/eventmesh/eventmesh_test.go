package eventmesh

import (
	"testing"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	. "github.com/onsi/gomega"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	backendbebv1 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/beb"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	controllertestingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
)

func Test_SyncBEBSubscription(t *testing.T) {
	g := NewWithT(t)

	credentials := &backendbebv1.OAuth2ClientCredentials{
		ClientID:     "foo-client-id",
		ClientSecret: "foo-client-secret",
	}
	// given
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	g.Expect(err).To(BeNil())

	nameMapper := utils.NewBEBSubscriptionNameMapper("mydomain.com", MaxEventMeshSubscriptionNameLength)
	beb := NewEventMesh(credentials, nameMapper, defaultLogger)

	// start BEB Mock
	bebMock := startBEBMock()
	envConf := env.Config{
		BEBAPIURL:                bebMock.MessagingURL,
		ClientID:                 "client-id",
		ClientSecret:             "client-secret",
		TokenEndpoint:            bebMock.TokenURL,
		WebhookActivationTimeout: 0,
		WebhookTokenEndpoint:     "webhook-token-endpoint",
		Domain:                   "domain.com",
		EventTypePrefix:          controllertestingv2.EventTypePrefix,
		BEBNamespace:             "/default/ns",
		Qos:                      string(types.QosAtLeastOnce),
	}

	err = beb.Initialize(envConf)
	g.Expect(err).To(BeNil())

	// when
	subscription := fixtureValidSubscription("some-name", "some-namespace")
	subscription.Status.Backend.Emshash = 0
	subscription.Status.Backend.Ev2hash = 0

	apiRule := controllertestingv2.NewAPIRule(subscription,
		controllertestingv2.WithPath(),
		controllertestingv2.WithService("foo-svc", "foo-host"),
	)

	// then
	changed, err := beb.SyncSubscription(subscription, &Cleaner{}, apiRule)
	g.Expect(err).To(BeNil())
	g.Expect(changed).To(BeTrue())
	bebMock.Stop()
}

// fixtureValidSubscription returns a valid subscription.
func fixtureValidSubscription(name, namespace string) *eventingv1alpha2.Subscription {
	return controllertestingv2.NewSubscription(
		name, namespace,
		controllertestingv2.WithSinkURL("https://webhook.xxx.com"),
		controllertestingv2.WithDefaultSource(),
		controllertestingv2.WithEventType(controllertestingv2.OrderCreatedEventTypeNotClean),
		controllertestingv2.WithWebhookAuthForBEB(),
	)
}

func startBEBMock() *controllertesting.BEBMock {
	beb := controllertesting.NewBEBMock()
	beb.Start()
	return beb
}

type Cleaner struct {
}

func (c *Cleaner) Clean(eventType string) (string, error) {
	// Cleaning is not needed in this test
	return eventType, nil
}
