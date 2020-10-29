package handlers

import (
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/mitchellh/hashstructure"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func getHash(subscription *types.Subscription) (int64, error) {
	if hash, err := hashstructure.Hash(subscription, nil); err != nil {
		return 0, err
	} else {
		return int64(hash), nil
	}
}

func getInternalView4Ev2(subscription *eventingv1alpha1.Subscription, apiRule *apigatewayv1alpha1.APIRule) (*types.Subscription, error) {
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

	// WebhookAuth
	auth := &types.WebhookAuth{}
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
		emsSubscription.WebhookAuth = auth
	}

	return emsSubscription, nil
}

// ConvertURLPortForApiRulePort converts string port from url.URL to uint32 port
func ConvertStringPortUInt32Port(u url.URL) (uint32, error) {
	port := uint32(0)
	sinkPort := u.Port()
	if sinkPort != "" {
		u64, err := strconv.ParseUint(u.Port(), 10, 32)
		if err != nil {
			return port, errors.Wrapf(err, "failed to convert port: %s", u.Port())
		}
		port = uint32(u64)
	}
	if port == uint32(0) {
		switch strings.ToLower(u.Scheme) {
		case "http":
			port = uint32(80)
		case "https":
			port = uint32(443)
		}
	}
	return port, nil
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

// GetRandSuffix returns a random suffix of length l or -1 for the whole random string
func GetRandSuffix(l int) string {
	b := make([]byte, l)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
