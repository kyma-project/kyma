---
title: EventActivation
type: Custom Resource
---

The `eventactivations.applicationconnector.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to create an Event Bus Subscription and to get an event schema. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd eventactivations.applicationconnector.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that allows you to consume events sent from the service with the `ac031e8c-9aa4-4cb7-8999-0d358726ffaa` ID in a `production` Namespace.

```yaml
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: EventActivation
metadata:
  name: "ac031e8c-9aa4-4cb7-8999-0d358726ffaa"
  namespace: production
spec:
  displayName: "Orders"
  sourceId: "prod"
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter   |      Required      |  Description |
|----------|:-------------:|------|
| **metadata.name** | Yes | Specifies the name of the CR and the ID of the Application service. This field is also used to fetch event schemas from the MinIO storage.  |
| **metadata.namespace** | Yes | Specifies the Namespace in which the CR is created. |
| **spec.displayName** | Yes | Specifies a human-readable name of the Application service. |
| **spec.sourceId** | Yes | Used to construct a Publish-Subscribe (Pub/Sub) topic name where events are sent and from where they are consumed. |

## Related resources and components

These are the resources related to this CR:

| Custom resource   |   Description |
|---------|------|
| Application |  Describes a service from which the user receives events. |
| Subscription | Contains information on how to create an infrastructure for consuming events. Works only if the EventActivation is enabled.  |

These components use this CR:

| Component   |   Description |
|----------|------|
| Application Broker |  Uses this CR to enable the user to receive events from a given service. |
| Event Bus | Uses this CR to control the consumption of an event.  |
| Serverless | Lambda UI sends a GraphQL query to Console Backend Service to list EventActivations. |
| Console Backend Service |  Exposes the given CR to the Console UI. |
