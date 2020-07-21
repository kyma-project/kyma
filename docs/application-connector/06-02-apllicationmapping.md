---
title: ApplicationMapping
type: Custom Resource
---

The `applicationmappings.application.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to enable APIs and events from an Application as a ServiceClass in a given Namespace. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd applicationmappings.applicationconnector.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample ApplicationMapping resource which enables the `test` Application in the `production` Namespace. In this example, all services provided by the Application are enabled:

```yaml
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: ApplicationMapping
metadata:
  name: test
  namespace: production
```

Using ApplicationMapping, you can also enable only the selected services in a given Namespace. See the example:

```yaml
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: ApplicationMapping
metadata:
  name: test
  namespace: production
spec:
  services:
    - id: ac031e8c-9aa4-4cb7-8999-0d358726ffaa
    - id: bef3143c-d1a5-674c-8dc9-ab4788896fba
```

The `services` list contains IDs of enabled services.

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter   |      Required      |  Description |
|----------|:-------------:|------|
| **metadata.name** | Yes | Specifies the name of the CR and the Application. |
| **metadata.namespace** | Yes | Specifies the Namespace to which the Application is bound. |
| **spec.services[]** | No | Lists enabled services. If the list is specified, only the selected services are enabled. If the list is empty, all services of the Application are enabled.|
| **spec.services[].id** | No | Specifies the ID of the enabled service.

## Related resources and components

These components use this CR:

| Component   |   Description |
|----------|------|
| Application Broker |  Uses this CR to enable the provisioning of ServiceClasses in a given Namespace. |
| Console Backend Service | Uses this CR to filter the enabled Applications. It also allows you to create or delete ApplicationMappings. |
