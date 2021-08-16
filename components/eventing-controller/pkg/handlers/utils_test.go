package handlers

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"testing"

	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/gomega"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"

	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func TestGetHash(t *testing.T) {
	g := NewGomegaWithT(t)

	bebSubscription := types.Subscription{}
	hash, err := getHash(&bebSubscription)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(hash).To(BeNumerically(">", 0))
}

func TestGetInternalView4Ev2(t *testing.T) {
	defaultProtocolSettings := &eventingv1alpha1.ProtocolSettings{
		ContentMode: func() *string {
			cm := "test"
			return &cm
		}(),
		ExemptHandshake: func() *bool {
			eh := true
			return &eh
		}(),
		Qos: func() *string {
			qos := "AT_LEAST_ONCE"
			return &qos
		}(),
	}

	defaultWebhookAuth := &types.WebhookAuth{
		Type:         types.AuthTypeClientCredentials,
		GrantType:    types.GrantTypeClientCredentials,
		ClientID:     "clientID",
		ClientSecret: "clientSecret",
		TokenURL:     "tokenURL",
	}

	defaultNameMapper := NewBebSubscriptionNameMapper("my-shoot", 50)

	defaultNamespace := "defaultNS"
	svcName := "foo-svc"
	host := "foo-host"
	scheme := "https"
	expectedWebhookURL := fmt.Sprintf("%s://%s", scheme, host)
	g := NewGomegaWithT(t)

	t.Run("subscription with protocolsettings where defaults are overriden", func(t *testing.T) {
		// given
		subscription := reconcilertesting.NewSubscription("name", "namespace", eventingtesting.WithEventTypeFilter)
		eventingtesting.WithValidSink("ns", svcName, subscription)

		givenProtocolSettings := &eventingv1alpha1.ProtocolSettings{
			ContentMode: func() *string {
				contentMode := eventingv1alpha1.ProtocolSettingsContentModeBinary
				return &contentMode
			}(),
			ExemptHandshake: func() *bool {
				exemptHandshake := true
				return &exemptHandshake
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
		}
		subscription.Spec.ProtocolSettings = givenProtocolSettings

		// Values should be overriden by the given values in subscription
		expectedBEBSubscription := types.Subscription{
			Name:            defaultNameMapper.MapSubscriptionName(subscription),
			ContentMode:     *givenProtocolSettings.ContentMode,
			Qos:             types.QosAtLeastOnce,
			ExemptHandshake: *givenProtocolSettings.ExemptHandshake,
			Events: types.Events{
				{
					Source: reconcilertesting.EventSource,
					Type:   reconcilertesting.EventType,
				},
			},
			WebhookUrl: expectedWebhookURL,
			WebhookAuth: &types.WebhookAuth{
				Type:         types.AuthTypeClientCredentials,
				GrantType:    types.GrantTypeClientCredentials,
				ClientID:     subscription.Spec.ProtocolSettings.WebhookAuth.ClientId,
				ClientSecret: subscription.Spec.ProtocolSettings.WebhookAuth.ClientSecret,
				TokenURL:     subscription.Spec.ProtocolSettings.WebhookAuth.TokenUrl,
			},
		}

		apiRule := reconcilertesting.NewAPIRule(subscription, reconcilertesting.WithPath)
		reconcilertesting.WithService(host, svcName, apiRule)

		// then
		gotBEBSubscription, err := getInternalView4Ev2(subscription, apiRule, defaultWebhookAuth, defaultProtocolSettings, "", defaultNameMapper)

		// when
		g.Expect(err).To(BeNil())
		g.Expect(expectedBEBSubscription).To(Equal(*gotBEBSubscription))
	})

	t.Run("subscription with default setting", func(t *testing.T) {
		// given
		subscription := reconcilertesting.NewSubscription("name", "namespace", eventingtesting.WithEmptySourceEventType)
		eventingtesting.WithValidSink("ns", svcName, subscription)

		// Values should retain defaults
		expectedBEBSubscription := types.Subscription{
			Name: defaultNameMapper.MapSubscriptionName(subscription),
			Events: types.Events{
				{
					Source: defaultNamespace,
					Type:   reconcilertesting.EventType,
				},
			},
			WebhookUrl:      expectedWebhookURL,
			WebhookAuth:     defaultWebhookAuth,
			Qos:             types.QosAtLeastOnce,
			ExemptHandshake: *defaultProtocolSettings.ExemptHandshake,
			ContentMode:     *defaultProtocolSettings.ContentMode,
		}

		apiRule := reconcilertesting.NewAPIRule(subscription, reconcilertesting.WithPath)
		reconcilertesting.WithService(host, svcName, apiRule)

		// then
		gotBEBSubscription, err := getInternalView4Ev2(subscription, apiRule, defaultWebhookAuth, defaultProtocolSettings, defaultNamespace, defaultNameMapper)

		// when
		g.Expect(err).To(BeNil())
		g.Expect(expectedBEBSubscription).To(Equal(*gotBEBSubscription))
	})
}

func TestGetInternalView4Ems(t *testing.T) {
	g := NewGomegaWithT(t)

	// given
	emsSubscription := &types.Subscription{
		Name:            "ev2subs1",
		ContentMode:     types.ContentModeStructured,
		ExemptHandshake: true,
		Qos:             types.QosAtLeastOnce,
		WebhookUrl:      "https://webhook.xxx.com",

		Events: []types.Event{
			{
				Source: reconcilertesting.EventSource,
				Type:   reconcilertesting.EventTypeNotClean,
			},
		},
	}

	// then
	bebSubscription, err := getInternalView4Ems(emsSubscription)

	// when
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(bebSubscription.Name).To(BeEquivalentTo(emsSubscription.Name))
	g.Expect(bebSubscription.ContentMode).To(BeEquivalentTo(emsSubscription.ContentMode))
	g.Expect(bebSubscription.ExemptHandshake).To(BeEquivalentTo(emsSubscription.ExemptHandshake))
	g.Expect(bebSubscription.Qos).To(BeEquivalentTo(types.QosAtLeastOnce))
	g.Expect(bebSubscription.WebhookUrl).To(BeEquivalentTo(emsSubscription.WebhookUrl))

	g.Expect(bebSubscription.Events).To(BeEquivalentTo(types.Events{
		{
			Source: reconcilertesting.EventSource,
			Type:   reconcilertesting.EventTypeNotClean,
		},
	}))
	g.Expect(bebSubscription)
}

func TestGetRandSuffix(t *testing.T) {
	totalExecutions := 10
	lengthOfRandomSuffix := 6
	results := make(map[string]bool)
	for i := 0; i < totalExecutions; i++ {
		result := GetRandString(lengthOfRandomSuffix)
		if _, ok := results[result]; ok {
			t.Fatalf("generated string already exists: %s", result)
		}
		results[result] = true
	}
}

func TestBebSubscriptionNameMapper(t *testing.T) {
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
	for _, test := range tests {
		mapper := NewBebSubscriptionNameMapper(test.domainName, test.maxLen)
		s := mapper.MapSubscriptionName(test.inputSub)
		g.Expect(len(s)).To(BeNumerically("<=", test.maxLen))
		// the mapped name should always end with the SHA1
		g.Expect(strings.HasSuffix(s, test.outputHash)).To(BeTrue())
		// and have the first 10 char of the name
		prefixLen := min(len(test.inputSub.Name), test.maxLen-hashLength)
		g.Expect(strings.HasPrefix(s, test.inputSub.Name[:prefixLen]))
	}

	// Same domain and subscription name/namespace should map to the same name
	mapper := NewBebSubscriptionNameMapper(domain1, 50)
	g.Expect(mapper.MapSubscriptionName(s3)).To(Equal(mapper.MapSubscriptionName(s4)))

	// If the same names are used in different order, they get mapped to different names
	g.Expect(mapper.MapSubscriptionName(s4)).ToNot(Equal(mapper.MapSubscriptionName(s5)))
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
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}
