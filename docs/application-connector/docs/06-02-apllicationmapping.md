---
title: ApplicationMapping
type: Custom Resource
---

The `applicationmappings.application.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to enable APIs and Events from an Application (App) as a ServiceClass in a given Namespace. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd applicationmappings.applicationconnector.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample ApplicationMapping resource which enables the `test` Application in the `production` Namespace:

```
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: ApplicationMapping
metadata:
  name: test
  namespace: production
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR and the App. |
| **metadata.namespace** |    **YES**   | Specifies the Namespace in which the App is enabled. |

## Related resources and components

These are the resources related to this CR:

| Custom resource   |   Description |
|:----------:|:------|
| ApplicationMapping |  Uses this CR to expose the services of an App in a given Namespace. |

These components use this CR:

| Component   |   Description |
|:----------:|:------|
| Application Broker |  Uses this CR to enable the provisioning of ServiceClasses in a given Namespace. |
| UI API Layer | Uses this CR to filter the enabled Apps. It also allows you to create or delete ApplicationMappings. |
