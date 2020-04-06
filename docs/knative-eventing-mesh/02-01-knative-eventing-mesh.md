---
title: Architecture
---

The architecture of Knative Eventing Mesh relies heavily on the functionality provided by the Knative Eventing components. To ensure a stable event flow between the sender and the subscriber, the Eventing Mesh wires Knative and Kyma components together.


This diagram shows how the Eventing Mesh components work together.

![Eventing implementation](./assets/eventing-mesh-implementation.svg)

1. The user creates an [Application CR](https://kyma-project.io/docs/components/application-connector/#custom-resource-application) and binds it to a Namespace. 

2. The Application Operator watches the creation of the Application CR and creates an [HTTPSource CR](#custom-resource-http-source) which defines the source sending the events.

3. The Event Source Controller watches the creation of the HTTPSource CR and deploys these resources:

    * [HTTP Source Adapter](https://github.com/kyma-project/kyma/tree/master/components/event-sources/adapter/http) which is an HTTP server deployed inside the `kyma-integration` Namespace. This adapter acts as a gateway to the Channel, and is responsible for exposing an endpoint to which the Application sends the events. 

    * [Channel](https://knative.dev/docs/eventing/channels/) which defines the way messages are dispatched in the Namespace. Its underlying implementation is responsible for forwarding events to the Broker or additional Channels. Kyma uses NATS Streaming as its default Channel, but you can change it to Kafka implementation. 

4. The Application Broker watches the creation of the Application CR and performs the following actions:

    * Exposes the Events API of an external system as a ServiceClass. Once the user provisions this ServiceClass in the Namespace, the Application Broker makes events available to use.

    * Deploys Knative Subscription and defines the Broker as the subscriber for the Channel to allow communication between them.

    * Adds the `knative-eventing-injection` label to the user's Namespace. As a result, the Namespace controller creates the [Broker](https://knative.dev/docs/eventing/broker-trigger/) which automatically receives the name `default`. The Broker acts as an entry point for the events and them at the cluster-local endpoint it exposes.

5. The user creates the [Trigger](https://knative.dev/docs/eventing/broker-trigger/) which references the Broker and defines the subscriber along with the conditions for filtering events. This way, certain subscribers receive only the events they are interested in.

For details on the Trigger specification, [this](/components/knative-eventing-mesh/#details-event-processing) document.
