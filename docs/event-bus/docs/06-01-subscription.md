---
title: Subscription
type: Custom Resource
---

The `subscriptions.eventing.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to create an Event trigger for a lambda or microservice in Kyma. After creating a new custom resource, the Event trigger is registered in the Event Bus and Events are delivered to the endpoint specified in the custom resource.

To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd subscriptions.eventing.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that creates an event trigger for a lambda with `order.created` event.

```yaml
apiVersion: eventing.kyma-project.io/v1alpha1
kind: Subscription
metadata:
  name: hello-with-data-subscription
  labels:
    example: event-bus-lambda-subscription
spec:
  endpoint: http://hello-with-data.{NAMESPACE}:8080/
  push_request_timeout_ms: 2000
  max_inflight: 400
  include_subscription_name_header: true
  event_type: order.created
  event_type_version: v1
  source_id: stage.commerce.kyma.local
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter                                 | Mandatory | Description                                                                                                                |
|:-----------------------------------------:|:---------:|:---------------------------------------------------------------------------------------------------------------------------|
| **metadata.name**                         | **YES**   | Specifies the name of the CR.                                                                                              |
| **spec.endpoint**                         | **YES**   | The HTTP endpoint to which events are delivered as a POST request.                                                         |
| **spec.push_request_timeout_ms**          | **YES**   | The HTTP request timeout. After the timeout has expired, event are redelivered.                                            |
| **spec.max_inflight**                     | **YES**   | The maximum number of concurrent HTTP requests to deliver events.                                                          |
| **spec.include_subscription_name_header** | **YES**   | Boolean flag indicating if the name of the subscription should be included in the HTTP headers while delivering the event. |
| **spec.event_type**                       | **YES**   | The event type to which the event trigger is registered.                                                                   |
| **spec.event_type_version**               | **YES**   | The version of the event type.                                                                                             |
| **spec.source_id**                        | **YES**   | Identifies the external the external solution which sent the event to Kyma.                                                |
