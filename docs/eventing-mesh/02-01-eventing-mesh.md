---
title: Architecture
---

The architecture of Knative Eventing Mesh relies heavily on the functionality provided by [Knative Eventing](https://knative.dev/docs/eventing/). To ensure a stable event flow between the Sender and the Subscriber, the Knative Eventing Mesh must include certain Knative elements. The first diagram shows how the Knative elements are implemented and wired with the already exisitng Kyma components, whereas the second diagram shows you how the events are processed by in Kyma.

## 


![User actions](./assets/eventing-mesh-user-actions.svg)

1. The user creates the [Application CR](https://kyma-project.io/docs/components/application-connector/#custom-resource-application) and binds it to the Namespace. 
2. The Application Operator watches the creation of an Application CR and creates the HTTP Source CR, which defines the event sources the events can come from. 
3. The Event Source Controller watches the creation of the HTTP Source CR and deploys these resources:
    * [HTTP source adapter](https://github.com/kyma-project/kyma/tree/master/components/event-sources/adapter/http) which is an HTTP server deployed inside the `kyma-integration` Namespace. The adapter acts as a gateway to the Knative Channel, and its responsibility is to simply expose an endpoint to receive events. 
    * [Knative Channel](https://knative.dev/docs/eventing/channels/) Custom Resource, which defines the event forwarding layer. 
4. The Application Broker watches the creation of the Application CR and performs the following actions:
    * Exposes event definitions as an event ServiceClass. Once the Application Broker provisions the ServiceClass, the events become available for services. 
    * Creates the Knative Subscription] Custom Resource which enables forwarding events received by the Knative Channel to subscribers or other Channels. 
    * Injects the [Knative Broker](https://knative.dev/docs/eventing/broker-trigger/) to the User Namespace. The Broker acts as an entry point for the events in the User Namespace. 
5. The user creates the [Knative Trigger](https://knative.dev/docs/eventing/broker-trigger/) which defines the subscriber along with the conditions for filtering out events. This way, certain subscribers receive only the events they are interested in. For details on Trigger CR specification, see the **Trigger Filtering** section of [this](https://knative.dev/docs/eventing/broker-trigger/) document.

This diagram shows you how the events are processed in Kyma. 

![Eventing Architecture](./assets/knative-event-mesh-send-events.svg)

1. The Application sends events to the HTTP source adapter. 

    >**NOTE:** Currently only CloudEvents in version 1.0. 

2. Knative Channel listens for incoming events and maps them to Knative Broker.
3. The Knative Broker forwards events to the Knative Trigger.
4. The Knative Trigger receives events and forwards them to subscribers, such as lambda functions or services.
5. The event meets specified conditions which and triggers the function.