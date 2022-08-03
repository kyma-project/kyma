---
title: Eventing Architecture
---

Eventing uses Event Publisher Proxy and Eventing Controller to connect to the default NATS JetStream backend. They work together to process and deliver events in Kyma.

## Event processing and delivery

The event processing and delivery flow uses the NATS server to process events and send them to subscribers.
This diagram explains the event flow in Kyma, from the moment an event source sends an event, to the point when the event triggers the Function.

![Eventing flow](./assets/evnt-architecture.svg)

1. The Eventing Controller watches the Subscription custom resource. It detects if there are any new incoming events.

2. The Eventing Controller creates an infrastructure for the NATS server.

3. An event source publishes events to the Event Publisher Proxy.

4. The Event Publisher Proxy sends events to the NATS server.

5. The NATS server dispatches events to the Eventing Controller.

6. The Eventing Controller dispatches events to subscribers (microservices or Functions).


## Event Publisher Proxy

Event Publisher Proxy receives legacy and Cloud Events, and publishes them to the configured eventing backend. All the legacy events are automatically converted to Cloud Events.

## Eventing Controller

Eventing Controller manages the internal infrastructure in order to receive an event. It watches Subscription custom resources. When an event is received, Eventing Controller dispatches the message to the configured sink.

## JetStream

Kyma now supports JetStream by default, which is a persistence offering from NATS, that guarantees `at least once` delivery. It is built-in within our default NATS backend.

The key advantages of JetStream over Core NATS are:

- At least once delivery of JetStream compared to at most once delivery of NATS.
- Streaming: Streams receive and store messages that are published and subscribers can consume these messages at any time.
- Persistent stream storage: Messages are retained in the stream storage even when the NATS server is restarted.
