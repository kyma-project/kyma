package beb

import (
	"testing"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	. "github.com/onsi/gomega"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func Test_SyncBEBSubscription(t *testing.T) {
	g := NewWithT(t)

	credentials := &OAuth2ClientCredentials{
		ClientID:     "foo-client-id",
		ClientSecret: "foo-client-secret",
	}
	// given
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	g.Expect(err).To(BeNil())

	nameMapper := utils.NewBEBSubscriptionNameMapper("mydomain.com", MaxBEBSubscriptionNameLength)
	beb := NewBEB(credentials, nameMapper, defaultLogger)

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
		EventTypePrefix:          controllertesting.EventTypePrefix,
		BEBNamespace:             "/default/ns",
		Qos:                      string(types.QosAtLeastOnce),
	}

	err = beb.Initialize(envConf)
	g.Expect(err).To(BeNil())

	// when
	subscription := fixtureValidSubscription("some-name", "some-namespace")
	subscription.Status.Emshash = 0
	subscription.Status.Ev2hash = 0

	apiRule := controllertesting.NewAPIRule(subscription,
		controllertesting.WithPath(),
		controllertesting.WithService("foo-svc", "foo-host"),
	)

	// then
	changed, err := beb.SyncSubscription(subscription, &Cleaner{}, apiRule)
	g.Expect(err).To(BeNil())
	g.Expect(changed).To(BeTrue())
	bebMock.Stop()
}

// fixtureValidSubscription returns a valid subscription
func fixtureValidSubscription(name, namespace string) *eventingv1alpha1.Subscription {
	return controllertesting.NewSubscription(
		name, namespace,
		controllertesting.WithSinkURL("https://webhook.xxx.com"),
		controllertesting.WithFilter(controllertesting.EventSource, controllertesting.OrderCreatedEventTypeNotClean),
		controllertesting.WithWebhookAuthForBEB(),
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
