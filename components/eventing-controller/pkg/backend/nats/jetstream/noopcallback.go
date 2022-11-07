//go:build noopcallback

package jetstream

import (
	"net/http"
	"strings"

	"github.com/nats-io/nats.go"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

const streamPrefix = env.JetStreamSubjectPrefix + "."

func (js *JetStream) getCallback(subKeyPrefix, subscriptionName string) nats.MsgHandler {
	return func(msg *nats.Msg) {
		et := strings.TrimPrefix(msg.Subject, streamPrefix)

		if err := msg.Ack(); err != nil {
			ceLogger.Errorw("Failed to ACK an event on JetStream")
		}

		js.metricsCollector.RecordDeliveryPerSubscription(subscriptionName, et, "noop-sink", http.StatusOK)
	}
}
