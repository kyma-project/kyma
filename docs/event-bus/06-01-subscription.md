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

This is a sample resource that creates an `order.created` Event trigger for a lambda.

```yaml
apiVersion: eventing.kyma-project.io/v1alpha1
kind: Subscription
metadata:
  name: hello-with-data-subscription
  labels:
    example: event-bus-lambda-subscription
spec:
  endpoint: http://hello-with-data.{NAMESPACE}:8080/
  include_subscription_name_header: true
  event_type: order.created
  event_type_version: v1
  source_id: stage.commerce.kyma.local
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter                                 | Mandatory | Description                                                                                                                |
|-----------------------------------------|:---------:|---------------------------------------------------------------------------------------------------------------------------|
| **metadata.name**                         | **YES**   | Specifies the name of the CR.                                                                                              |
| **spec.endpoint**                         | **YES**   | The HTTP endpoint to which events are delivered as a POST request.                                                         |
| **spec.include_subscription_name_header** | **YES**   | The boolean flag indicating if the name of the Subscription should be included in the HTTP headers while delivering the Event. |
| **spec.event_type**                       | **YES**   | The Event type to which the Event trigger is registered. For example, **order.created**.                                                                 |
| **spec.event_type_version**               | **YES**   | The version of the Event type.                                                                                             |
| **spec.source_id**                        | **YES**   | Identifies the origin of events. This can be an external solution or a defined identifier for internally generated events.|
