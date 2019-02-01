---
title: Event flow requirements
type: Details
---

The Event Bus enables a successful flow of Events in Kyma when:

- You have enabled the [EventActivation](docs/components/event-bus#concepts) controller.
- You have created a [Subscription](/docs/components/event-bus#custom-resource-subscription) custom resource and registered the webhook for a lambda or service to consume the Events.
- The Events are [published](#details-event-flow-requirements-event-publishing).

## Details

Read the following subsections for details on the requirements.

### Activate Events

Use the EventActivation to ensure the Event flow between the Namespace and the Application (App). You can also simply [bind](/docs/components/application-connector#getting-started-bind-an-application-to-a-namespace) the App to the Namespace.

The diagram shows you the Event activation flow for a given Namespace.

![EventActivation.png](./assets/event-activation.svg)

The App sends the Events to the Event Bus and uses the EventApplication controller to ensure the Namespace receives the Events.  If you define a lambda in the `prod123` Namespace, it receives the **order.created** Event from the App through the Event Bus. The lambda in `test123` Namespace does not receive any Events, since you have not enabled the EventActivation controller.


### Consume Events

Configure the lambdas and services to use `push` for consuming Events coming from the App. To make sure the lambda or the service receive the Events, register a webhook for them by creating a [Subscription custom resource](/docs/components/event-bus#custom-resource-subscription). 

### Publish Events

Make sure that the external solution sends Events to Kyma.
