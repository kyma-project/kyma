---
title: Troubleshooting using tracing
type: Troubleshooting
---

The tracing functionality available in Kyma can help you to pinpoint the error's root cause and solve the problem. See [this](docs/components/tracing/#overview-overview) document to learn more about tracing.

## No microservice or lambda configured to receive an Event

In this case, an external system sends the Event, but
a lambda or microservice with an Event trigger does not exist.

As a result, you can see traces for `publish` only, and the trace details show you only the tags for the `event-publish-knative-service`.

![](./assets/troubleshoot-only-publish-detail.png)

## Configured microservice or lambda returns an error

In this case, a microservice or lambda exists and reacts on
the Event trigger. However, due to a code issue, the microservice or lambda 
fails to process the Event.

As a result, the `webhook`, `push`, and `name-of-lambda` services in the trace are marked with error.

![](./assets/troubleshoot-error-in-lambda.png)

To see the error details, click one of the service spans, such as the one for `push` service.
![](./assets/troubleshoot-error-in-lambda-details.png)

Since the Event Bus keeps on retrying to deliver the Event until it is successful, you 
can see multiple spans for the `webhook-service`.

![](./assets/troubleshoot-error-multiple-spans.png)
