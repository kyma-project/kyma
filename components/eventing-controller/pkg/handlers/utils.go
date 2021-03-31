package handlers

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"

	"github.com/pkg/errors"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/mitchellh/hashstructure"
)

// MessagingBackendHandler exposes a common handler interface for different messaging backend systems
type MessagingBackendHandler interface {
	Initialize(cfg env.Config) error
	SyncSubscription(subscription *eventingv1alpha1.Subscription, cleaner eventtype.Cleaner, params ...interface{}) (bool, error)
	DeleteSubscription(subscription *eventingv1alpha1.Subscription) error
}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func getHash(subscription *types.Subscription) (int64, error) {
	if hash, err := hashstructure.Hash(subscription, nil); err != nil {
		return 0, err
	} else {
		return int64(hash), nil
	}
}

func getInternalView4Ev2(subscription *eventingv1alpha1.Subscription, apiRule *apigatewayv1alpha1.APIRule, defaultWebhookAuth *types.WebhookAuth) (*types.Subscription, error) {
	emsSubscription := &types.Subscription{}

	// Name
	emsSubscription.Name = subscription.Name
	emsSubscription.ContentMode = subscription.Spec.ProtocolSettings.ContentMode
	emsSubscription.ExemptHandshake = subscription.Spec.ProtocolSettings.ExemptHandshake

	// Qos
	qos := strings.ReplaceAll(subscription.Spec.ProtocolSettings.Qos, "-", "_")
	if qos == string(types.QosAtLeastOnce) {
		emsSubscription.Qos = types.QosAtLeastOnce
	} else if qos == string(types.QosAtMostOnce) {
		emsSubscription.Qos = types.QosAtMostOnce
	} else {
		return nil, fmt.Errorf("invalid Qos: %v", subscription.Spec.ProtocolSettings.Qos)
	}

	// WebhookUrl
	urlTobeRegistered, err := getExposedURLFromAPIRule(apiRule, subscription)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get exposed URL from APIRule")
	}
	emsSubscription.WebhookUrl = urlTobeRegistered

	// Events
	for _, e := range subscription.Spec.Filter.Filters {
		s := e.EventSource.Value
		t := e.EventType.Value
		emsSubscription.Events = append(emsSubscription.Events, types.Event{Source: s, Type: t})
	}

	// Using default webhook auth unless specified in Subscription CR
	auth := defaultWebhookAuth
	if subscription.Spec.ProtocolSettings.WebhookAuth != nil {
		auth.ClientID = subscription.Spec.ProtocolSettings.WebhookAuth.ClientId
		auth.ClientSecret = subscription.Spec.ProtocolSettings.WebhookAuth.ClientSecret
		if subscription.Spec.ProtocolSettings.WebhookAuth.GrantType == string(types.GrantTypeClientCredentials) {
			auth.GrantType = types.GrantTypeClientCredentials
		} else {
			return nil, fmt.Errorf("invalid GrantType: %v", subscription.Spec.ProtocolSettings.WebhookAuth.GrantType)
		}
		if subscription.Spec.ProtocolSettings.WebhookAuth.Type == string(types.AuthTypeClientCredentials) {
			auth.Type = types.AuthTypeClientCredentials
		} else {
			return nil, fmt.Errorf("invalid Type: %v", subscription.Spec.ProtocolSettings.WebhookAuth.Type)
		}
		auth.TokenURL = subscription.Spec.ProtocolSettings.WebhookAuth.TokenUrl
	}
	emsSubscription.WebhookAuth = auth
	return emsSubscription, nil
}

func getExposedURLFromAPIRule(apiRule *apigatewayv1alpha1.APIRule, sub *eventingv1alpha1.Subscription) (string, error) {
	scheme := "https://"
	path := ""

	sURL, err := url.ParseRequestURI(sub.Spec.Sink)
	if err != nil {
		return "", err
	}
	sURLPath := sURL.Path
	if sURL.Path == "" {
		sURLPath = "/"
	}
	for _, rule := range apiRule.Spec.Rules {
		if rule.Path == sURLPath {
			path = rule.Path
			break
		}
	}
	return fmt.Sprintf("%s%s%s", scheme, *apiRule.Spec.Service.Host, path), nil
}

func getInternalView4Ems(subscription *types.Subscription) (*types.Subscription, error) {
	emsSubscription := &types.Subscription{}

	// Name
	emsSubscription.Name = subscription.Name
	emsSubscription.ContentMode = subscription.ContentMode
	emsSubscription.ExemptHandshake = subscription.ExemptHandshake

	// Qos
	emsSubscription.Qos = subscription.Qos

	// WebhookUrl
	emsSubscription.WebhookUrl = subscription.WebhookUrl

	// Events
	for _, e := range subscription.Events {
		s := e.Source
		t := e.Type
		emsSubscription.Events = append(emsSubscription.Events, types.Event{Source: s, Type: t})
	}

	return emsSubscription, nil
}

// GetRandString returns a random string of the given length
func GetRandString(l int) string {
	b := make([]byte, l)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
