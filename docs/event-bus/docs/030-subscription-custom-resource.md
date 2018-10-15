---
title: Subscription
type: Custom Resource
---

The `subscriptions.eventing.kyma.cx` Custom Resource Definition (CRD) is a detailed description of the kind of data and the format used to create an event trigger for lambda/microservice in Kyma. After creating a new custom resource, the event trigger is registered in the event bus and events will be delivered to the endpoint specified in the custom resource. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd subscriptions.eventing.kyma.cx -o yaml
```

## Sample custom resource

This is a sample resource that creates an event trigger for a lambda with `order.created` event.

```
apiVersion: eventing.kyma.cx/v1alpha1
kind: Subscription
metadata:
  name: hello-with-data-subscription
  labels:
    example: event-bus-lambda-subscription
spec:
  endpoint: http://hello-with-data.<environment>:8080/
  push_request_timeout_ms: 2000
  max_inflight: 400
  include_subscription_name_header: true
  event_type: order.created
  event_type_version: v1
  source_id: stage.commerce.kyma.local
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter                                 | Mandatory | Description                                                                                                                 |
|:-----------------------------------------:|:---------:|:----------------------------------------------------------------------------------------------------------------------------|
| **metadata.name**                         | **YES**   | Specifies the name of the CR.                                                                                               |
| **spec.endpoint**                         | **YES**   | The HTTP endpoint to which events will be delivered as a POST request.                                                      |
| **spec.push_request_timeout_ms**          | **YES**   | The HTTP request timeout. After the timeout expires, event will be redelivered.                                             |
| **spec.max_inflight**                     | **YES**   | The max concurrent HTTP requests to deliver events.                                                                         |
| **spec.include_subscription_name_header** | **YES**   | Boolean flag to indicate if the name of the subscription should be included in the HTTP headers while delivering the event. |
| **spec.event_type**                       | **YES**   | The event type to which the event trigger will be registered. e.g. `order.created`                                          |
| **spec.event_type_version**               | **YES**   | The version of the event type.                                                                                              |
| **spec.source_id**                        | **YES**   | This field identifies the external solution from which the event was sent to Kyma.                                          |

