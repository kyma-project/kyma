---
title: Architecture
---

When you create a lambda or a service to perform a given business functionality, you must define which Events trigger it. Define triggers by creating the Subscription CR in which you instruct the Event Bus to forward the Events of a particular type to your lambda. 
For example, whenever the `order-created` Event comes in, the Event Bus consumes it by saving it in NATS Streaming and persistence. Then it sends it to the receiver specified in the Subscription definition.

> **NOTE:** The Event Bus creates a separate Event Trigger for each Subscription.

## Event consumption

![Configure and Consume Events](./assets/configure-consume-events.svg)


1. A user creates a lambda or a service to be triggered by an Event coming from an external solution. 
    >**NOTE**: For a service, the user must create a Kyma Subscription resource manually. For a lambda, it is created automatically.
2. The **subscription-controller-knative** component reacts on the creation of Kyma Subscription. It [verifies](#event-validation) if the Event type has Subscription permissions to if so, it creates Knative Channel and Knative Subscription resources.
3. The **nats-controller** reacts on the creation of a Knative Channel and creates the required Kubernetes and Istio services.
4. The **nats-dispatcher** reacts on the creation of a Knative Subscription and creates the NATS Streaming subscription. 
5. The **nats-dispatcher** picks the Event and dispatches it to the configured lambda or the service URL as an HTTP POST request. The lambda reacts on the received Event.

## Event publishing

![Publish Events](./assets/publish-events.svg)

1. The external application integrated with Kyma makes a REST API request to the Application Connector's Events Gateway to indicate that a new Event is available. The request provides the Application Connector with the Event metadata. 
2. The Application Connector enriches the Event with the details of its source.

    > **NOTE:** There is always one dedicated instance of the Application Connector for every instance of an external solution connected to Kyma.

3. The Application Connector makes a REST API call to the **publish-knative** component and sends the enriched Event.
4. **publish-knative** makes the HTTP payload compatible with Knative and sends the Event to the relevant **knative-channel** service URL which is inferred based on **source id**, **event type** and **event type version** parameters.
5. Istio Virtual Service forwards the Event further to the **nats-dispatcher** service served by the **nats-dispatcher** Pod.
6. The NATS dispatcher stores the Event in NATS Streaming which stores the Event details in the Persistence storage volume.



## Event validation 

The **subscription-controller-knative** checks if the Namespace can receive Events from the Application. It performs the check for each Kyma Subscription created in a Namespace for a particular Event type with a version for a specific Application.

### Validation flow

See the diagram and a step-by-step description of the Event verification process.

![Event validation process](./assets/event-validation.svg)

1. The Kyma user defines a lambda or a service.
2. The Kyma user creates a Subscription custom resource.
3. The **subscription-controller-knative** reads the new Subscription.
4. The **subscription-controller-knative** reads the EventActivation CR to verify if it exists in the Namespace for a certain Application.
5. The **subscription-controller-knative**  updates the Subscription resource accordingly with the activation status `true` or `false`. The Event Bus uses this status to allow or prohibit Event delivery.