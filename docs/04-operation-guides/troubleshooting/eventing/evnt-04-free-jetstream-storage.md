---
title: Eventing backend stopped receiving events due to full storage
---

## Symptom

NATS JetStream backend stopped receiving events due to full storage.

You see one of the following messages:
- `507 Insufficient Storage` error from Event Publisher Proxy.
- `no space left on device` from the eventing backend.

## Cause

In Kyma, the default retention policy for NATS JetStream is [Interest](https://docs.nats.io/using-nats/developer/develop_jetstream/model_deep_dive).
This retention policy keeps messages in the stream if they can't be delivered to the sink, as long as there are consumers in the stream that match the published event's subject.

If there are too many undelivered events, the NATS JetStream storage may get full.
To prevent event loss, the backend stops receiving events, and no further events can be persisted to the stream.

## Remedy

There are several ways to free the space on NATS JetStream backend:

- If the published events are too large, the consumer cannot deliver them fast enough before the storage is full.
  In that case, either wait until the events are delivered, or scale the NATS backend with additional replicas.


- Check the [NATS JetStream backend status](evnt-01-eventing-troubleshooting.md#step-6-check-nats-jetstream-status) and if [the sink is reachable and can accept the events](evnt-01-eventing-troubleshooting.md#step-5-check-if-the-subscription-sink-is-healthy).


- The `Interest` retention policy specifies that events published to the subject are not kept in the stream if they don't match any consumer filter.
  You can delete a Kyma Subscription, which automatically removes all the pending messages in the stream that were published to that Subscription's subject. 
