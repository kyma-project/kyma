---
title: Glossary
---

- **Event Types**
    - `CloudEvents`: Events that confirm to the [CloudEvents specification](https://cloudevents.io/) - a common specification for describing event data. The specification is currently under [CNCF](https://www.cncf.io/).
    - `Legacy events` (deprecated): Events or messages published to Kyma that do not conform to the CloudEvents specification. All legacy events published to Kyma are converted to CloudEvents.
- **Streams and Consumers**
    - `Streams`: A stream stores messages for the published events. Kyma uses only one stream, with _**file**_ storage, for all the events. You can configure the retention and delivery policies for the stream, depending on the use case.
    - `Consumers`: A consumer reads or consumes the messages from the stream. Kyma Subscription creates one consumer for each specified filter. Kyma uses push-based consumers.
- **Delivery Guarantees**
    - `at least once` delivery: With NATS JetStream, Kyma now ensures that for each event published, all the subscribers subscribed to that event receive the event at least once.
    - `max bytes and discard policy`: NATS JetStream uses these configurations to ensure that no messages are lost when the storage is almost full. By default, Kyma ensures that no new messages are accepted when the storage reaches 90% capacity.  
 