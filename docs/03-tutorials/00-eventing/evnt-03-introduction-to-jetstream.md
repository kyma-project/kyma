---
title: Introduction to JetStream
---

JetStream is the 'built-in distributed persistence system' of NATS. Read the [Eventing architecture](../../05-technical-reference/00-architecture/evnt-01-architecture.md#jetstream) page for more information.

## Streams and Consumers

A Stream stores messages for the published events. In Kyma, we use only one stream for all the events. You can configure the retention and delivery policies for the stream, depending on the use case.

A Consumer reads or consumes the messages from the stream. Kyma Subscription creates one consumer for each filter specified. In Kyma we use push-based consumers.

## Steps to verify at least once delivery with JetStream backend

This tutorial shows how JetStream persists events, even when the sink is not reachable, and redelivers the event when the sink is available again.

1. Create a [Function](../../02-get-started/04-trigger-workload-with-event.md#create-a-function), [Subscription](../../02-get-started/04-trigger-workload-with-event.md#create-a-subscription) and [trigger the workload with an event](../../02-get-started/04-trigger-workload-with-event.md#trigger-the-workload-with-an-event).

2. Once you have JetStream enabled, delete your Function (the sink of Subscription).

```bash
kubectl -n default delete function lastorder
```

3. Follow the [Trigger the workload with an event](../../02-get-started/04-trigger-workload-with-event.md#trigger-the-workload-with-an-event) tutorial to trigger the event once again. The message is stored and is visible in the stream.
    1. Port forward the Kyma Eventing NATS Service to localhost. Use port `4222`. Run:
       ```bash
       kubectl -n kyma-system port-forward svc/eventing-nats 4222
       ```
    2. Use the [nats CLI](https://github.com/nats-io/natscli) to list the streams:
       ```bash
       nats stream ls
       ```

       Notice that the stream contains undelivered messages:
       ```bash
        ╭────────────────────────────────────────────────────────────────────────────╮
        │                                  Streams                                   │
        ├──────┬─────────────┬─────────────────────┬──────────┬───────┬──────────────┤
        │ Name │ Description │ Created             │ Messages │ Size  │ Last Message │
        ├──────┼─────────────┼─────────────────────┼──────────┼───────┼──────────────┤
        │ sap  │             │ 2022-04-26 00:00:00 │ 1        │ 318 B │ 5.80s        │
        ╰──────┴─────────────┴─────────────────────┴──────────┴───────┴──────────────╯
       ```

5. Recreate your [Function](../../02-get-started/04-trigger-workload-with-event.md#create-a-function). The stream no longer contains any undelivered messages and the event is delivered. Follow the [tutorial](../../02-get-started/04-trigger-workload-with-event.md#verify-the-event-delivery) to verify the event delivery.
