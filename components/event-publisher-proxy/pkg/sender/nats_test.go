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
	t.Parallel()

	testCases := []struct {
		name                      string
		givenNatsConnectionClosed bool
		wantError                 bool
		wantStatusCode            int
	}{
		{
			name:                      "send should succeed if NATS connection is open",
			givenNatsConnectionClosed: false,
			wantError:                 false,
			wantStatusCode:            http.StatusNoContent,
		},
		{
			name:                      "send should fail if NATS connection is not open",
			givenNatsConnectionClosed: true,
			wantError:                 true,
			wantStatusCode:            http.StatusBadGateway,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			natsServer := testingutils.StartNatsServer()
			assert.NotNil(t, natsServer)
			defer natsServer.Shutdown()

			connection, err := pkgnats.Connect(natsServer.ClientURL(),
				pkgnats.WithRetryOnFailedConnect(true),
				pkgnats.WithMaxReconnects(1),
				pkgnats.WithReconnectWait(time.Second),
			)
			assert.NoError(t, err)
			assert.NotNil(t, connection)

			receive := make(chan bool, 1)
			validator := testingutils.ValidateNatsMessageDataOrFail(t, fmt.Sprintf(`"%s"`, testingutils.CloudEventData), receive)
			testingutils.SubscribeToEventOrFail(t, connection, testingutils.CloudEventType, validator)

			ce := testingutils.StructuredCloudEventPayloadWithCleanEventType
			event := cloudevents.NewEvent()
			event.SetType(testingutils.CloudEventType)
			err = json.Unmarshal([]byte(ce), &event)
			assert.NoError(t, err)

			ctx := context.Background()
			sender := NewNatsMessageSender(context.Background(), connection, logrus.New())

			if tc.givenNatsConnectionClosed {
				connection.Close()
			}

			status, err := sender.Send(ctx, &event)
			assert.Equal(t, tc.wantError, err != nil)
			assert.Equal(t, tc.wantStatusCode, status)

			err = testingutils.WaitForChannelOrTimeout(receive, time.Millisecond*5)
			assert.Equal(t, tc.wantError, err != nil)
		})
	}
}
