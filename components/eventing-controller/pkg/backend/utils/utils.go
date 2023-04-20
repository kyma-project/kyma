package utils

import (
	"encoding/json"

	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/nats-io/nats.go"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type EventTypeInfo struct {
	OriginalType  string
	CleanType     string
	ProcessedType string
}

// NameMapper is used to map Kyma-specific resource names to their corresponding name on other
// (external) systems, e.g. on different eventing backends, the same Kyma subscription name
// could map to a different name.
type NameMapper interface {
	MapSubscriptionName(subscriptionName, subscriptionNamespace string) string
}

func APIRuleGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  apigatewayv1beta1.GroupVersion.Version,
		Group:    apigatewayv1beta1.GroupVersion.Group,
		Resource: "apirules",
	}
}

func ConvertMsgToCE(msg *nats.Msg) (*cev2event.Event, error) {
	event := cev2event.New(cev2event.CloudEventsVersionV1)
	err := json.Unmarshal(msg.Data, &event)
	if err != nil {
		return nil, err
	}
	if err := event.Validate(); err != nil {
		return nil, err
	}
	return &event, nil
}
