package sender

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"

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

			natsServer := testingutils.StartNATSServer(false)
			assert.NotNil(t, natsServer)
			defer natsServer.Shutdown()

			connection, err := pkgnats.Connect(natsServer.ClientURL(),
				pkgnats.WithRetryOnFailedConnect(true),
				pkgnats.WithMaxReconnects(1),
				pkgnats.WithReconnectWait(time.Second),
			)
			assert.NoError(t, err)
			assert.NotNil(t, connection)
			defer func() { connection.Close() }()

			receive := make(chan bool, 1)
			validator := testingutils.ValidateNatsMessageDataOrFail(t, fmt.Sprintf(`"%s"`, testingutils.EventData), receive)
			testingutils.SubscribeToEventOrFail(t, connection, testingutils.CloudEventType, validator)

			ce := createCloudEvent(t)

			mockedLogger, _ := logger.New("json", "info")

			ctx := context.Background()
			sender := NewNatsMessageSender(context.Background(), connection, mockedLogger)

			if tc.givenNatsConnectionClosed {
				connection.Close()
			}

			status, err := sender.Send(ctx, ce)
			assert.Equal(t, tc.wantError, err != nil)
			assert.Equal(t, tc.wantStatusCode, status)

			err = testingutils.WaitForChannelOrTimeout(receive, time.Millisecond*5)
			assert.Equal(t, tc.wantError, err != nil)
		})
	}
}
