---
title: Architecture
---

The architecture of Knative Eventing Mesh relies heavily on the functionality provided by [Knative Eventing](https://knative.dev/docs/eventing/). To ensure stable event flow between the sender and the subscriber, Knative Eventing Mesh wires Knative elements closely with the existing Kyma components.

## Knative Eventing Mesh implementation

This diagram shows the Knative Eventing Mesh implementation along with its components.

![Eventing implementation](./assets/eventing-mesh-implementation.svg)

1. The user creates the [Application CR](https://kyma-project.io/docs/components/application-connector/#custom-resource-application) and binds it to the Namespace. 

2. The Application Operator watches the creation of the Application CR and creates the [HTTPSource CR](#custom-resource-http-source) which defines the source sending the events.

3. The Event Source Controller watches the creation of the HTTPSource CR and deploys these resources:

    * [HTTP Adapter](https://github.com/kyma-project/kyma/tree/master/components/event-sources/adapter/http) which is an HTTP server deployed inside the `kyma-integration` Namespace. The adapter acts as a gateway to the Knative Channel, and is responsible for exposing an endpoint the Application sends the events to. 

    * [Knative Channel](https://knative.dev/docs/eventing/channels/) which defines the way messages are dispatched in the Namespace. Its underlying implementation, such as NATS Streaming or Kafka Channel, is responsible for forwarding events to multiple destinations. 

4. The Application Broker watches the creation of the Application CR and performs the following actions:

    * Exposes event definitions as an event ServiceClass. Once the user deploys the Service using this ServiceClass, the Application Broker provisions it to make events available for services.

    * Adds the `knative-eventing-injection` label to the user Namespace. As a result, the Namespace controller creates the [Knative Broker](https://knative.dev/docs/eventing/broker-trigger/) which acts as an entry point for the events. 

    * Creates the Knative Subscription and defines the Broker as the Subscriber for the Channel to allow communication.

5. The user creates [Knative Trigger](https://knative.dev/docs/eventing/broker-trigger/) which references Knative Broker and defines the subscriber along with the conditions for filtering events. This way, certain subscribers receive only those events they are interested in. For details on Knative Trigger specification, see the **Trigger Filtering** section of [this](https://knative.dev/docs/eventing/broker-trigger/) document.

## Event flow 

This diagram explains the event flow in Kyma, from the moment the Application sends an event, to the point when the event triggers the function.

![Eventing flow](./assets/eventing-mesh-flow.svg)

1. The Application sends events to the HTTP Adapter which forwards them to the resource such as Knative Broker.
   
    >**NOTE:** The HTTP adapter accepts only CloudEvents in version 1.0. 

2. Knative Subscription defines the Broker as the Subscriber. This way, Knative Channel can communicate with the Broker to send events.

3. Knative Channel listens for incoming events. When it receives an event, the underlying messaging layer dispatches it to Knative Broker.

4. Knative Broker sends events to Knative Trigger which registered interest in receiving them.

5. Knative Trigger checks if the attributes of an incoming event match its specification. If they do, Knative Trigger sends the event to the Subscriber.