# Eventing Backend Stopped Receiving Events Due To Full Storage

## Symptom

NATS JetStream backend stopped receiving events due to full storage.

You observe the following behavior in the Eventing Publisher Proxy (EPP):

- `507 Insufficient Storage` HTTP Status from EPP on the publish request.
- `cannot send to stream: nats: maximum bytes exceeded` in the EPP logs.

## Cause

In Kyma, the default retention policy for NATS JetStream is [Interest](https://docs.nats.io/using-nats/developer/develop_jetstream/model_deep_dive).
This retention policy keeps messages in the stream if they can't be delivered to the sink, as long as there are consumers in the stream that match the published event's subject.

If there are too many undelivered events, the NATS JetStream storage may get full.
To prevent event loss, the backend stops receiving events, and no further events can be persisted to the stream.

> [!NOTE]
> If you delete a Subscriber (sink) while there is still a Kyma Subscription pointing to that sink, the events published to that Subscription pile up in the stream and possibly delay the event delivery to other Subscribers.

## Solution

There are several ways to free the space on NATS JetStream backend:

- If the published events are too large, the consumer cannot deliver them fast enough before the storage is full.
  In that case, either slow down the events' publish rate until the events are delivered, or scale the NATS backend with additional replicas.

- Check the [NATS JetStream backend status](evnt-01-eventing-troubleshooting.md#step-6-check-nats-jetstream-status) and if [the sink is reachable and can accept the events](evnt-01-eventing-troubleshooting.md#step-5-check-if-the-subscription-sink-is-healthy).

- The `Interest` retention policy specifies that events published to the subject are not kept in the stream if they don't match any consumer filter.
  You can delete a Kyma Subscription, which automatically removes all the pending messages in the stream that were published to that Subscription's subject.

- If the events' publish rate is very high (more than 1.5k events per second), speed up the event dispatching by increasing the `maxInFlightMessages` configuration of the Subscription (default is set to 10) accordingly. Due to low `maxInFlightMessages`, the dispatcher will not be able to keep up with the publisher, and as a result, the stream size will keep growing.  
