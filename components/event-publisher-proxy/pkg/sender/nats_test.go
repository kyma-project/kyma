package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	pkgnats "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/nats"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

func TestNatsMessageSender(t *testing.T) {
	testCases := []struct {
		name                    string
		givenNatsServerShutdown bool
		wantError               bool
		wantStatusCode          int
	}{
		{
			name:                    "Send should succeed if NATS connection is open",
			givenNatsServerShutdown: false,
			wantError:               false,
			wantStatusCode:          http.StatusNoContent,
		},
		{
			name:                    "Send should fail if NATS connection is not open",
			givenNatsServerShutdown: true,
			wantError:               true,
			wantStatusCode:          http.StatusBadGateway,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			natsServer := testingutils.StartNatsServer()
			assert.NotNil(t, natsServer)
			defer natsServer.Shutdown()

			bc := pkgnats.NewBackendConnection(natsServer.ClientURL(), true, 1, time.Second)
			err := bc.Connect()
			assert.NoError(t, err)
			assert.NotNil(t, bc.Connection)

			if tc.givenNatsServerShutdown {
				natsServer.Shutdown()
			}

			receive := make(chan bool, 1)
			validator := testingutils.ValidateNatsMessageDataOrFail(t, fmt.Sprintf(`"%s"`, testingutils.CloudEventData), receive)
			testingutils.SubscribeToEventOrFail(t, bc.Connection, testingutils.CloudEventType, validator)

			ce := testingutils.StructuredCloudEventPayloadWithCleanEventType
			event := cloudevents.NewEvent()
			event.SetType(testingutils.CloudEventType)
			err = json.Unmarshal([]byte(ce), &event)
			assert.NoError(t, err)

			ctx := context.Background()
			sender := NewNatsMessageSender(context.Background(), bc, logrus.New())

			status, err := sender.Send(ctx, &event)
			assert.Equal(t, tc.wantError, err != nil)
			assert.Equal(t, status, tc.wantStatusCode)

			err = testingutils.WaitForChannelOrTimeout(receive, time.Second*3)
			assert.Equal(t, tc.wantError, err != nil)
		})
	}
}
