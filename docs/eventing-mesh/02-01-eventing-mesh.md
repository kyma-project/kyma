---
title: Architecture
---

The diagram shows you how the events are processed from the moment the Application sends them until they are received by the function. 

![Eventing Architecture](./assets/knative-event-mesh-send-events.svg)

1. The Application sends events to the [HTTP source adapter](https://github.com/kyma-project/kyma/tree/master/components/event-sources/adapter/http) which is an HTTP server deployed inside the `kyma-integration` Namespace. The adapter acts as a gateway to the [Knative Channel](https://knative.dev/docs/eventing/channels/default-channels/), and its responsibility is to simply expose an endpoint to receive events. 

    >**NOTE:** Currently only CloudEvents in version 1.0. 

2. Knative Channel listens for incoming events and maps them to the [Knative Broker](https://knative.dev/docs/eventing/broker-trigger/) which acts as an entry point for the events in the User Namespace. 
3. The Knative Broker forwards events to the Knative Trigger which defines the subscriber along with the conditions for filtering out events. This way, certain subscribers receive only the events they are interested in. For details on Trigger CR specification, see the **Trigger Filtering** section of [this](https://knative.dev/docs/eventing/broker-trigger/) document.
4. The Knative Trigger receives events and forwards them to subscribers, such as lambda functions or services.
5. The event meets specified conditions which and triggers the function. 