# Eventing Module

The Eventing module ships Eventing Manager, which is a standard Kubernetes operator that observes the state of Eventing resources and reconciles them according to the desired state.

With Kyma Eventing, you can focus on your business workflows and trigger them with events to implement asynchronous flows within Kyma. Generally, eventing consists of event producers (or publishers) and consumers (or subscribers) that send events to or receive events from an event processing backend.

The objective of Eventing in Kyma is to simplify the process of publishing and subscribing to events. Kyma uses proven eventing backend technology to provide a seamless experience to the user with their end-to-end business flows. The user does not have to implement or integrate any intermediate backend or protocol.

Kyma Eventing uses the following technology:

- [NATS JetStream](https://docs.nats.io/) as backend within the cluster
- [HTTP POST](https://www.w3schools.com/tags/ref_httpmethods.asp) requests to simplify sending and receiving events
- Declarative [Subscription custom resource (CR)](./resources/evnt-cr-subscription.md) to subscribe to events

## Kyma Eventing Flow

Kyma Eventing follows the PubSub messaging pattern: Kyma publishes messages to a messaging backend, which filters these messages and sends them to interested subscribers. Kyma does not send messages directly to the subscribers as shown below:

![PubSub](../assets/evnt-pubsub.svg)

Eventing in Kyma from a user’s perspective works as follows:

- Offer an HTTP end point, for example a Function to receive the events.
- Specify the events the user is interested in using the Kyma [Subscription CR](./resources/evnt-cr-subscription.md).
- Send [CloudEvents](https://cloudevents.io/) or legacy events (deprecated) to the following HTTP end points on our [Eventing Publisher Proxy](https://github.com/kyma-project/eventing-publisher-proxy/blob/main/README.md) service.
  - `/publish` for CloudEvents.
  - `<application_name>/v1/events` for legacy events.

For more information, read [Eventing architecture](evnt-architecture.md).

## Glossary

- **Event Types**
  - `CloudEvents`: Events that conform to the [CloudEvents specification](https://cloudevents.io/) - a common specification for describing event data. The specification is currently under [CNCF](https://www.cncf.io/).
  - `Legacy events` (deprecated): Events or messages published to Kyma that do not conform to the CloudEvents specification. All legacy events published to Kyma are converted to CloudEvents.
- **Streams and Consumers**
  - `Streams`: A stream stores messages for the published events. Kyma uses only one stream, with _**file**_ storage, for all the events. You can configure the retention and delivery policies for the stream, depending on the use case.
  - `Consumers`: A consumer reads or consumes the messages from the stream. Kyma Subscription creates one consumer for each specified filter. Kyma uses push-based consumers.
- **Delivery Guarantees**
  - `at least once` delivery: With NATS JetStream, Kyma ensures that for each event published, all the subscribers subscribed to that event receive the event at least once.
  - `max bytes and discard policy`: NATS JetStream uses these configurations to ensure that no messages are lost when the storage is almost full. By default, Kyma ensures that no new messages are accepted when the storage reaches 90% capacity.  
