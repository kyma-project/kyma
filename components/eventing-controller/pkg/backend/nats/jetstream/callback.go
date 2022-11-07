//go:build !noopcallback

package jetstream

import (
	"context"
	"net/http"

	cev2 "github.com/cloudevents/sdk-go/v2"
	cev2protocol "github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/nats-io/nats.go"

	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/tracing"
)

func (js *JetStream) getCallback(subKeyPrefix, subscriptionName string) nats.MsgHandler {
	return func(msg *nats.Msg) {
		// fetch sink info from storage
		sinkValue, ok := js.sinks.Load(subKeyPrefix)
		if !ok {
			js.namedLogger().Errorw("Failed to find sink URL in storage", "keyPrefix", subKeyPrefix)
			return
		}
		// convert interface type to string
		sink, ok := sinkValue.(string)
		if !ok {
			js.namedLogger().Errorw("Failed to convert sink value to string", "sinkValue", sinkValue)
			return
		}
		ce, err := backendutils.ConvertMsgToCE(msg)
		if err != nil {
			js.namedLogger().Errorw("Failed to convert JetStream message to CloudEvent", "error", err)
			return
		}

		// setup context for dispatching
		ctxWithCancel, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctxWithCE := cev2.ContextWithTarget(ctxWithCancel, sink)
		traceCtxWithCE := tracing.AddTracingHeadersToContext(ctxWithCE, ce)

		// decorate the logger with CloudEvent context
		ceLogger := js.namedLogger().With("id", ce.ID(), "source", ce.Source(), "type", ce.Type(), "sink", sink)

		ceLogger.Debugw("Sending the CloudEvent")

		// dispatch the event to sink
		result := js.client.Send(traceCtxWithCE, *ce)
		if !cev2protocol.IsACK(result) {
			js.metricsCollector.RecordDeliveryPerSubscription(subscriptionName, ce.Type(), sink, http.StatusInternalServerError)
			ceLogger.Errorw("Failed to dispatch the CloudEvent")
			// Do not NAK the msg so that the server waits for AckWait and then redeliver the msg.
			return
		}

		// event was successfully dispatched, check if acknowledged by the NATS server
		// if not, the message is redelivered.
		if err := msg.Ack(); err != nil {
			ceLogger.Errorw("Failed to ACK an event on JetStream")
		}

		js.metricsCollector.RecordDeliveryPerSubscription(subscriptionName, ce.Type(), sink, http.StatusOK)
		ceLogger.Infow("CloudEvent was dispatched")
	}
}
