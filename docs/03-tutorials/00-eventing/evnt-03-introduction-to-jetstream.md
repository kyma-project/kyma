---
title: Introduction to JetStream
---

JetStream is the 'built-in distributed persistence system' of NATS. You can read more about it [here](../../05-technical-reference/00-architecture/evnt-01-architecture.md#jetstream).

## Streams and Consumers

A `Stream` is a 'message store' for the events published. In Kyma, we use only one stream for all the events. You can configure the retention and delivery policies for the stream, based on the use-case.

A `Consumer` is a 'view' into the stream. A kyma subscription creates one consumer for each filter specified. In Kyma we use push based consumers.

## Steps to verify at least once delivery with JetStream backend

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
