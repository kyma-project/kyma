---
title: Event flow requirements
type: Details
---

The Event Bus enables a successful flow of Events in Kyma when:

- You have created an [EventActivation](docs/components/event-bus#concepts) controller.
- You have created a [Subscription](/docs/components/event-bus#custom-resource-subscription)  custom resource and register the webhook for a lambda or service to consume the Events.
- The Events are [published](#details-event-flow-requirements-event-publishing).

## Details

See the following subsections for details on each requirement.

### Activate Events

Use the EventActivation to ensure the Event flow between the Namespace and the Application (App). You can also simply [bind](/docs/components/application-connector#getting-started-bind-an-application-to-a-namespace) the App to the Namespace.

The diagram shows you the Event flow for a particular Namespace.

![EventActivation.png](./assets/event-activation.svg)

The App sends the Events to the Event Bus and uses the EventApplication controller to ensure the Namespace receives the Events.  If you define a lambda in the `prod123` Namespace, it receives the **order.created** Event from the App using the EventApplication controller. The lambda in `test123` Namespace does not receive any Events, since you have not enabled the  you need to enable the EventActivation. 



### Consume Events

Enable lambdas and services to consume Events in Kyma between any Namespace and an App using `push`. Deliver Events to the lambda or the service by registering a webhook for it. Create a [Subscription custom resource](/docs/components/event-bus#custom-resource-subscription) to register the webhook.


### Publish Events

Make sure that the external solution sends Events to Kyma.
