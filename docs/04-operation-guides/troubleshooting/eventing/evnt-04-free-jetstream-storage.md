---
title: Free NATS JetStream file storage when it gets full
---

## Symptom

Free NATS JetStream file storage when it gets full.

## Cause

NATS JetStream uses [Interest retention policy](https://docs.nats.io/using-nats/developer/develop_jetstream/model_deep_dive) in Kyma per default.
It means, that as long as there are consumers on the stream, which the published event's subject, the messages will be kept in the stream if they cannot be delivered to the sink.

In some cases, it might happen, that the NATS JetStream storage gets full due to too many undelivered events.
In order to not lose events, JetStream backend will just stop receiving and the user will get `507 Insufficient Storage` Error from Publisher Proxy or `no space left on device` from the Backend directly.
This means, that the Backend's storage is full and no further events can be persisted to the stream.

## Remedy

There are several ways of how to free the space on NATS JetStream Backend:

- The published events might be too large, so the consumer isn't fast enough to deliver them, before the storage gets full. In that case eiter wait until the events get delivered or extend the NATS Backend by additional replicas.


- Check if the sink is reachable and can accept the events.


- Due to `Interest` Policy, the events published to the subject, which doesn't match any consumer filter, will not be kept in the stream.
  You can delete a Kyma Subscription, which will automatically remove all the pending messages in the stream, which were published the Subscription's subject. 
