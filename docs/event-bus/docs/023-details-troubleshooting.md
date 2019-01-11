---
title: Troubleshooting
type: Details
---

In some cases, you can encounter some problems related to eventing. This
document introduces several ways to troubleshoot such problems.

## General Troubleshooting Guidelines

* If the lambda or the service does not receive any Events, check the following:
  - Confirm that the EventActivation custom resource is in place.
  - Ensure that the webhook defined for the lambda or the service is up and
    running.
  - Make sure the Events are published.

* If errors appear while sending Events:
  - Check if the `publish` application is up and running.
  - Make sure that NATS Streaming is up and running.

 If these general guidelines do not help, go to the next section of this
 document.

## Search by tags
You can search traces using tags. Tags are key-value pairs configured for each service.

See the full list of tags for a service from the details of that service's span.

For example, these are the tags for the `publish-service`:
* `event-type`
* `event-type-ver`
* `event-id`
* `source-id`

To search the traces, you can use either a single tag such as `event-type="order.created"`, or multiple tags such as `event-type="order.created" event-type-ver="v1"`.

## Troubleshooting using Kyma Tracing

Tracing allows you to troubleshoot different problems that you might encounter
while using Kyma. Understanding the common scenarios and how the expected traces
look like gives you a better grasp on how to quickly use Kyma tracing
capabilities to pinpoint the root cause. See the exemplary scenario for
reference.

### Scenario: I have no microservice or lambda configured to receive an Event

This scenario assumes that there is an Event sent from the external system but
there is no lambda or microservice configured with the Event trigger.

As a result, only the trace for the `publish` and initial services are visible.

![](./assets/troubleshoot-only-publish-overview.png)

In the trace details, you can see the tags for the `publish-service`.

![](./assets/troubleshoot-only-publish-detail.png)

### Scenario: Configured microservice or lambda returns an error

This scenario assumes that there is a microservice or lambda configured to recieve
the event trigger. However, due to a bug in the code, the microservice or lambda 
failed to process the Event.

As a result, you can see the `webhook`, `push`, and `name-of-lambda` services in the trace and they are marked with error.

![](./assets/troubleshoot-error-in-lambda.png)

To see the error details, click on one of the service spans. For example, choose the span for the `push` service.
![](./assets/troubleshoot-error-in-lambda-details.png)

Since the Event Bus keeps on retrying to deliver the Event until it is successful, you 
can see multiple spans for the `webhook-service`.

![](./assets/troubleshoot-error-multiple-spans.png)
