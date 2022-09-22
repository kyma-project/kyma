//nolint:gosec
package utils

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func TestGetHash(t *testing.T) {
	g := NewGomegaWithT(t)

	bebSubscription := types.Subscription{}
	hash, err := GetHash(&bebSubscription)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(hash).To(BeNumerically(">", 0))
}

//func TestConvertKymaSubToEventMeshSub(t *testing.T) {
//	defaultProtocolSettings := &eventingv1alpha2.ProtocolSettings{
//		ContentMode: func() *string {
//			cm := types.ContentModeBinary
//			return &cm
//		}(),
//		ExemptHandshake: func() *bool {
//			eh := true
//			return &eh
//		}(),
//		Qos: utils.StringPtr(string(types.QosAtLeastOnce)),
//	}
//
//	defaultWebhookAuth := &types.WebhookAuth{
//		Type:         types.AuthTypeClientCredentials,
//		GrantType:    types.GrantTypeClientCredentials,
//		ClientID:     "clientID",
//		ClientSecret: "clientSecret",
//		TokenURL:     "tokenURL",
//	}
//	defaultNameMapper := NewBEBSubscriptionNameMapper("my-shoot", 50)
//
//	bebSubEvents := types.Events{types.Event{
//		Source: eventingtesting.EventMeshNamespace,
//		//Type:   eventingtesting.OrderCreatedEventType,
//		Type: "prefix.testapp1023.order.created.v1", // @TODO: update it once cleaner logic is finalized
//	}}
//
//	// getProcessedEventTypes returns the processed types after cleaning and prefixing.
//	getTypeInfos := func(types []string) []EventTypeInfo {
//		result := make([]EventTypeInfo, 0, len(types))
//		for _, t := range types {
//			result = append(result, EventTypeInfo{OriginalType: t, CleanType: t, ProcessedType: t})
//		}
//		return result
//	}
//
//	defaultNamespace := eventingtesting.EventMeshNamespace
//	svcName := "foo-svc"
//	host := "foo-host"
//	scheme := "https"
//	expectedWebhookURL := fmt.Sprintf("%s://%s", scheme, host)
//	g := NewGomegaWithT(t)
//
//	t.Run("subscription with protocol settings where defaults are overridden", func(t *testing.T) {
//		// given
//		subscription := eventingtesting.NewSubscription("name", "namespace",
//			eventingtesting.WithDefaultSource(),
//			eventingtesting.WithOrderCreatedFilter(),
//			eventingtesting.WithValidSink("ns", svcName),
//		)
//
//		subscription.Spec.ProtocolSettings = eventingtesting.NewProtocolSettings(
//			eventingtesting.WithBinaryContentMode(),
//			eventingtesting.WithExemptHandshake(),
//			eventingtesting.WithAtLeastOnceQOS(),
//			eventingtesting.WithDefaultWebhookAuth(),
//		)
//
//		eventTypeInfos := getTypeInfos(subscription.Spec.Types)
//
//		// Values should be overridden by the given values in subscription
//		expectedWebhookAuth := &types.WebhookAuth{
//			Type:         types.AuthTypeClientCredentials,
//			GrantType:    types.GrantTypeClientCredentials,
//			ClientID:     subscription.Spec.ProtocolSettings.WebhookAuth.ClientID,
//			ClientSecret: subscription.Spec.ProtocolSettings.WebhookAuth.ClientSecret,
//			TokenURL:     subscription.Spec.ProtocolSettings.WebhookAuth.TokenURL,
//		}
//		expectedBEBSubscription := eventingtesting.NewBEBSubscription(
//			defaultNameMapper.MapSubscriptionName(subscription),
//			*subscription.Spec.ProtocolSettings.ContentMode,
//			expectedWebhookURL,
//			bebSubEvents,
//			expectedWebhookAuth,
//		)
//
//		apiRule := eventingtesting.NewAPIRule(subscription,
//			eventingtesting.WithPath(),
//			eventingtesting.WithService(svcName, host),
//		)
//
//		// then
//		gotBEBSubscription, err := ConvertKymaSubToEventMeshSub(subscription, eventTypeInfos, apiRule, defaultWebhookAuth, defaultProtocolSettings, defaultNamespace, defaultNameMapper)
//
//		// when
//		g.Expect(err).To(BeNil())
//		g.Expect(*expectedBEBSubscription).To(Equal(*gotBEBSubscription))
//	})
//
//	t.Run("subscription with default setting", func(t *testing.T) {
//		// given
//		subscription := eventingtesting.NewSubscription("name", "namespace",
//			eventingtesting.WithOrderCreatedFilter(),
//			eventingtesting.WithValidSink("ns", svcName),
//		)
//
//		eventTypeInfos := getTypeInfos(subscription.Spec.Types)
//
//		expectedBEBSubWithDefault := eventingtesting.NewBEBSubscription(
//			defaultNameMapper.MapSubscriptionName(subscription),
//			*defaultProtocolSettings.ContentMode,
//			expectedWebhookURL,
//			bebSubEvents,
//			defaultWebhookAuth, // WebhookAuth should retain defaults
//		)
//
//		apiRule := eventingtesting.NewAPIRule(subscription,
//			eventingtesting.WithPath(),
//			eventingtesting.WithService(svcName, host),
//		)
//
//		// then
//		gotBEBSubscription, err := ConvertKymaSubToEventMeshSub(subscription, eventTypeInfos, apiRule, defaultWebhookAuth, defaultProtocolSettings, defaultNamespace, defaultNameMapper)
//
//		// when
//		g.Expect(err).To(BeNil())
//		g.Expect(*expectedBEBSubWithDefault).To(Equal(*gotBEBSubscription))
//	})
//
//	t.Run("subscription with custom webhookauth config followed by a subscription with default webhookauth config should not alter the default config", func(t *testing.T) {
//		// given
//		subWithGivenWebhookAuth := eventingtesting.NewSubscription("name", "namespace",
//			eventingtesting.WithOrderCreatedFilter(),
//			eventingtesting.WithValidSink("ns", svcName),
//		)
//
//		subWithGivenWebhookAuth.Spec.ProtocolSettings = eventingtesting.NewProtocolSettings(
//			eventingtesting.WithBinaryContentMode(),
//			eventingtesting.WithExemptHandshake(),
//			eventingtesting.WithAtLeastOnceQOS(),
//			eventingtesting.WithDefaultWebhookAuth(),
//		)
//		expectedWebhookAuth := types.WebhookAuth{
//			Type:         types.AuthTypeClientCredentials,
//			GrantType:    types.GrantTypeClientCredentials,
//			ClientID:     subWithGivenWebhookAuth.Spec.ProtocolSettings.WebhookAuth.ClientID,
//			ClientSecret: subWithGivenWebhookAuth.Spec.ProtocolSettings.WebhookAuth.ClientSecret,
//			TokenURL:     subWithGivenWebhookAuth.Spec.ProtocolSettings.WebhookAuth.TokenURL,
//		}
//
//		expectedBEBSubWithWebhookAuth := eventingtesting.NewBEBSubscription(
//			defaultNameMapper.MapSubscriptionName(subWithGivenWebhookAuth),
//			*subWithGivenWebhookAuth.Spec.ProtocolSettings.ContentMode,
//			expectedWebhookURL,
//			bebSubEvents,
//			&expectedWebhookAuth, // WebhookAuth should retain the supplied config
//		)
//
//		apiRule := eventingtesting.NewAPIRule(subWithGivenWebhookAuth,
//			eventingtesting.WithPath(),
//			eventingtesting.WithService(svcName, host),
//		)
//
//		// then
//		eventTypeInfos := getTypeInfos(subWithGivenWebhookAuth.Spec.Types)
//		gotBEBSubscription, err := ConvertKymaSubToEventMeshSub(subWithGivenWebhookAuth, eventTypeInfos, apiRule, defaultWebhookAuth, defaultProtocolSettings, defaultNamespace, defaultNameMapper)
//
//		// when
//		g.Expect(err).To(BeNil())
//		g.Expect(*expectedBEBSubWithWebhookAuth).To(Equal(*gotBEBSubscription))
//
//		// Use another subscription without webhookAuthConfig
//		// given
//		subscriptionWithoutWebhookAuth := eventingtesting.NewSubscription("name", "namespace",
//			eventingtesting.WithOrderCreatedFilter(),
//			eventingtesting.WithValidSink("ns", svcName),
//		)
//
//		expectedBEBSubWithDefault := eventingtesting.NewBEBSubscription(
//			defaultNameMapper.MapSubscriptionName(subscriptionWithoutWebhookAuth),
//			*subWithGivenWebhookAuth.Spec.ProtocolSettings.ContentMode,
//			expectedWebhookURL,
//			bebSubEvents,
//			defaultWebhookAuth, // WebhookAuth should retain defaults
//		)
//
//		apiRule = eventingtesting.NewAPIRule(subscriptionWithoutWebhookAuth,
//			eventingtesting.WithPath(),
//			eventingtesting.WithService(svcName, host),
//		)
//
//		// then
//		gotBEBSubWithDefaultCfg, err := ConvertKymaSubToEventMeshSub(subscriptionWithoutWebhookAuth, eventTypeInfos, apiRule, defaultWebhookAuth, defaultProtocolSettings, defaultNamespace, defaultNameMapper)
//
//		// when
//		g.Expect(err).To(BeNil())
//		g.Expect(*expectedBEBSubWithDefault).To(Equal(*gotBEBSubWithDefaultCfg))
//		g.Expect(*expectedBEBSubWithDefault.WebhookAuth).To(Equal(*gotBEBSubWithDefaultCfg.WebhookAuth))
//
//	})
//}

func TestGetCleanedEventMeshSubscription(t *testing.T) {
	g := NewGomegaWithT(t)

	// given
	emsSubscription := &types.Subscription{
		Name:            "ev2subs1",
		ContentMode:     types.ContentModeStructured,
		ExemptHandshake: true,
		Qos:             types.QosAtLeastOnce,
		WebhookURL:      "https://webhook.xxx.com",

		Events: []types.Event{
			{
				Source: eventingtesting.EventSource,
				Type:   eventingtesting.OrderCreatedEventTypeNotClean,
			},
		},
	}

	// then
	eventMeshSubscription := GetCleanedEventMeshSubscription(emsSubscription)

	// when
	g.Expect(eventMeshSubscription.Name).To(BeEquivalentTo(emsSubscription.Name))
	g.Expect(eventMeshSubscription.ContentMode).To(BeEquivalentTo(emsSubscription.ContentMode))
	g.Expect(eventMeshSubscription.ExemptHandshake).To(BeEquivalentTo(emsSubscription.ExemptHandshake))
	g.Expect(eventMeshSubscription.Qos).To(BeEquivalentTo(types.QosAtLeastOnce))
	g.Expect(eventMeshSubscription.WebhookURL).To(BeEquivalentTo(emsSubscription.WebhookURL))

	g.Expect(eventMeshSubscription.Events).To(BeEquivalentTo(types.Events{
		{
			Source: eventingtesting.EventSource,
			Type:   eventingtesting.OrderCreatedEventTypeNotClean,
		},
	}))
	g.Expect(eventMeshSubscription)
}

func TestBEBSubscriptionNameMapper(t *testing.T) {
	g := NewGomegaWithT(t)

	s1 := &eventingv1alpha1.Subscription{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      "subscription1",
			Namespace: "my-namespace",
		},
	}
	s2 := &eventingv1alpha1.Subscription{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      "mysub",
			Namespace: "another-namespace",
		},
	}

	s3 := &eventingv1alpha1.Subscription{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      "name1",
			Namespace: "name2",
		},
		Spec: eventingv1alpha1.SubscriptionSpec{
			Sink: "sub3-sink",
		},
	}
	s4 := &eventingv1alpha1.Subscription{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      "name1",
			Namespace: "name2",
		},
		Spec: eventingv1alpha1.SubscriptionSpec{
			Sink: "sub4-sink",
		},
	}
	s5 := &eventingv1alpha1.Subscription{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      "name2",
			Namespace: "name1",
		},
	}

	domain1 := "my-domain-name.com"
	domain2 := "another.domain.com"

	hashLength := 40

	tests := []struct {
		domainName string
		maxLen     int
		inputSub   *eventingv1alpha1.Subscription
		outputHash string
	}{
		{
			domainName: domain1,
			maxLen:     50,
			inputSub:   s1,
			outputHash: hashSubscriptionFullName(domain1, s1.Namespace, s1.Name),
		},
		{
			domainName: domain2,
			maxLen:     50,
			inputSub:   s1,
			outputHash: hashSubscriptionFullName(domain2, s1.Namespace, s1.Name),
		},
		{
			domainName: "",
			maxLen:     50,
			inputSub:   s2,
			outputHash: hashSubscriptionFullName("", s2.Namespace, s2.Name),
		},
	}

	minFunc := func(i, j int) int {
		if i < j {
			return i
		}
		return j
	}

	for _, test := range tests {
		mapper := NewBEBSubscriptionNameMapper(test.domainName, test.maxLen)
		s := mapper.MapSubscriptionName(test.inputSub.Name, test.inputSub.Namespace)
		g.Expect(len(s)).To(BeNumerically("<=", test.maxLen))
		// the mapped name should always end with the SHA1
		g.Expect(strings.HasSuffix(s, test.outputHash)).To(BeTrue())
		// and have the first 10 char of the name
		prefixLen := minFunc(len(test.inputSub.Name), test.maxLen-hashLength)
		g.Expect(strings.HasPrefix(s, test.inputSub.Name[:prefixLen]))
	}

	// Same domain and subscription name/namespace should map to the same name
	mapper := NewBEBSubscriptionNameMapper(domain1, 50)
	g.Expect(mapper.MapSubscriptionName(s3.Name, s3.Namespace)).To(Equal(mapper.MapSubscriptionName(s4.Name, s4.Namespace)))

	// If the same names are used in different order, they get mapped to different names
	g.Expect(mapper.MapSubscriptionName(s4.Name, s4.Namespace)).ToNot(Equal(mapper.MapSubscriptionName(s5.Name, s5.Namespace)))
}

func TestShortenNameAndAppendHash(t *testing.T) {
	g := NewGomegaWithT(t)
	fakeHash := fmt.Sprintf("%x", sha1.Sum([]byte("myshootmynamespacemyname")))

	tests := []struct {
		name   string
		hash   string
		maxLen int
		output string
	}{
		{
			name:   "mylongsubscription",
			hash:   fakeHash,
			maxLen: 50,
			output: "mylongsubs" + fakeHash,
		},
		{
			name:   "mysub",
			hash:   fakeHash,
			maxLen: 50,
			output: "mysub" + fakeHash,
		},
		{
			name:   "mysub",
			hash:   fakeHash,
			maxLen: 40,
			output: fakeHash, // no room for name!
		},
	}
	for _, test := range tests {
		nameWithHash := shortenNameAndAppendHash(test.name, test.hash, test.maxLen)
		g.Expect(nameWithHash).To(Equal(test.output))
	}

	// shortenNameAndAppendHash should panic if it cannot fit the hash
	defer func() {
		g.Expect(recover()).ToNot(BeNil())
	}()
	shortenNameAndAppendHash("panic-much", fakeHash, len(fakeHash)-1)
}
