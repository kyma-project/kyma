---
title: EventActivation
type: Custom Resource
---

The `eventactivations.applicationconnector.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to create an Event Bus Subscription and to get an Event schema. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd eventactivations.applicationconnector.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that allows you to consume Events sent from the service with the `ac031e8c-9aa4-4cb7-8999-0d358726ffaa` ID in a `production` Namespace.

```
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


| Parameter   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR and the ID of the Application service. This field is also used to fetch Event schemas from the Minio storage.  |
| **metadata.namespace** |    **YES**   | Specifies the Namespace in which the CR is created. |
| **spec.displayName** |    **YES**   | Specifies a human-readable name of the Application service. |
| **spec.sourceId** |    **YES**   | Used to construct a Publish-Subscribe (Pub/Sub) topic name where the Events are send and from where the Events are consumed. |

## Related resources and components

These are the resources related to this CR:

| Custom resource   |   Description |
|:----------:|:------|
| Application |  Describes a service from which the user receives Events. |
| Subscription | Contains information on how to create an infrastructure for consuming Events. Works only if the EventActivation is enabled.  |

These components use this CR:

| Component   |   Description |
|:----------:|:------|
| Application Broker |  Uses this CR to enable the user to receive Events from a given service. |
| Event Bus | Uses this CR to control the consumption of an Event.  |
| Serverless | Lambda UI sends a GraphQL query to UI API Layer to list EventActivations. |
| UI API Layer |  Exposes the given CR to the Console UI. |
