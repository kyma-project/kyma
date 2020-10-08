package ems2


import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	client2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/client"
	config2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/config"
	types2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/types"
	auth2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/auth"

	"github.com/mitchellh/hashstructure"
)

const (
	// timeouts
	activeTimeout = time.Second * 10 // timeout for the subscription to be active

	// subscription config
	source1           = "/default/sap.kyma/kt1"  // from env
	source2           = "/default/sap.kyma/kt1" // TODO use another source when available
	subscriptionName1 = "testSubscription1"
	subscriptionName2 = "testSubscription2"
	eventType1        = "tunas.ev2.poc.event1.v1"
	eventType2        = "tunas.ev2.poc.event2.v1"
	endpoint1         = "https://httpbin.radu-1.tunas.nachtmaar.de/post"  // from env
	endpoint2         = "https://httpbin.radu-1.tunas.nachtmaar.de/post"
	// calling the subscriber
	subscriptionClientId = "b18bb6c4-2ba1-403b-af7e-a03a89af0343"  // from env
	subscriptionClientSecret = "Dtrs5-dwPvmcibLu53zm8MVNNA"        // from env
	subscriptionTokenUrl = "https://oauth2.radu-1.tunas.nachtmaar.de/oauth2/token" //from env
)

func TestEmsE2E(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		eventType    string
		endpoint     string
		publishCount int
		cleanup      bool
	}{
		{
			name:         subscriptionName1,
			source:       source1,
			eventType:    eventType1,
			endpoint:     endpoint1,
			publishCount: 2,
			cleanup:      true,
		},
		{
			name:         subscriptionName2,
			source:       source2,
			eventType:    eventType2,
			endpoint:     endpoint2,
			publishCount: 2,
			cleanup:      true,
		},
	}

	// authenticate
	authenticator := auth2.NewAuthenticator(auth2.GetDefaultConfig())
	token, err := authenticator.Authenticate()
	if err != nil {
		t.Fatalf("Failed to authenticate with error: %v", err)
	}

	evtClient := client2.NewClient(config2.GetDefaultConfig())

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create subscription
			//subscription := newSubscription(test.name, test.source, test.eventType, test.endpoint)
			subscription := newSubscriptionOauth2(test.name, test.source, test.eventType, test.endpoint)
			createResponse, err := evtClient.Create(token, subscription)
			if err != nil {
				t.Logf("Failed to create subscription with error: %#v", err)
			}
			if createResponse.StatusCode > http.StatusAccepted && createResponse.StatusCode != http.StatusConflict {
				t.Logf("Failed to create subscription with error: %#v", createResponse)
			}

			// get the subscription
			emsSubscription, err :=  evtClient.Get(token, test.name)
			if err != nil {
				t.Logf("Failed to get subscription with error: %#v", err)
			}
			t.Logf("EMS Subscription: %#v", emsSubscription)
			// calculate the hash value for EMS Subscription
			hash, err := hashstructure.Hash(emsSubscription, nil)
			if err != nil {
				t.Logf("Failed to calculate the hash value with error: %#v", err)
			} else {
				t.Logf("EMS Subscription hash value: %v", hash)
			}

			// publish if subscription is active
			if waitForSubscriptionActive(evtClient, token, subscription.Name, activeTimeout) {
				for i := 0; i < test.publishCount; i++ {
					id := fmt.Sprintf("%s-%d", test.name, i)
					response, err := evtClient.Publish(token, newCloudevent(id, test.source, test.eventType), types2.QosAtLeastOnce)
					if err != nil {
						t.Logf("Cannot publish event: %v", err)
					}
					if response.StatusCode > http.StatusNoContent {
						t.Logf("Cannot publish event: %v", response)
					}
				}
			}

			// TODO test event delivery

			if !test.cleanup {
				return
			}

			// delete subscription
			if _, err := evtClient.Delete(token, subscription.Name); err != nil {
				t.Logf("Failed to delete subscription\n")
			}
		})
	}
}

func newSubscription(name, source, eventType, endpoint string) types2.Subscription {
	return types2.Subscription{
		Name: name,
		Events: []types2.Event{
			{
				Source: source,
				Type:   eventType,
			},
		},
		WebhookUrl: endpoint,
		WebhookAuth: &types2.WebhookAuth{
			Type:     types2.AuthTypeBasic,
			User:     fmt.Sprintf("%s-usr", name),
			Password: fmt.Sprintf("%s-pwd", name),
		},
		ExemptHandshake: true,
		Qos:             types2.QosAtLeastOnce,
		//ContentMode:     types2.ContentModeBinary,
	}
}

func newSubscriptionOauth2(name, source, eventType, endpoint string) types2.Subscription {
	return types2.Subscription{
		Name: name,
		Events: []types2.Event{
			{
				Source: source,
				Type:   eventType,
			},
		},
		WebhookUrl: endpoint,
		WebhookAuth: &types2.WebhookAuth{
			Type:     types2.AuthTypeClientCredentials,
			GrantType: types2.GrantTypeClientCredentials,
			ClientID: subscriptionClientId,
			ClientSecret: subscriptionClientSecret,
			TokenURL: subscriptionTokenUrl,
		},
		ExemptHandshake: true,
		Qos:             types2.QosAtLeastOnce,
		//ContentMode:     types2.ContentModeBinary,
	}
}

func newCloudevent(id, source, eventType string) cloudevents.Event {
	payload := struct {
		ID          string `json:"ID,omitempty"`
		Source      string `json:"Source,omitempty"`
		EventType   string `json:"EventType,omitempty"`
		PublishTime string `json:"PublishTime,omitempty"`
	}{
		ID:          id,
		Source:      source,
		EventType:   eventType,
		PublishTime: fmt.Sprintf("%d", time.Now().Unix()),
	}

	event := cloudevents.NewEvent()
	event.SetID(id)
	event.SetSpecVersion("1.0")
	event.SetSource(source)
	event.SetType(eventType)
	if err := event.SetData(cloudevents.ApplicationJSON, payload); err != nil {
		log.Fatalf("Failed to set cloudevent data: %v", err)
	}

	return event
}

func waitForSubscriptionActive(evtClient *client2.Client, token *auth2.AccessToken, name string, timeLimit time.Duration) bool {
	timeout := time.After(timeLimit)
	tick := time.Tick(time.Millisecond * 100)

	for {
		select {
		case <-timeout:
			{
				return false
			}
		case <-tick:
			{
				sub, err := evtClient.Get(token, name)
				if err != nil {
					return false
				}
				if sub != nil && sub.SubscriptionStatus == types2.SubscriptionStatusActive {
					return true
				}
			}
		}
	}
}
