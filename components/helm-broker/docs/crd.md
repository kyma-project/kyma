---
title: AddonsConfiguration
type: Custom Resource
---

The `addonsconfiguration.addons.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define an Addons configuration for the Helm Broker. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd addonsconfiguration.addons.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample AddonsConfiguration CR configuration:

```yaml
apiVersion: addons.kyma-project.io/v1alpha1
kind: AddonsConfiguration
metadata:
  name: addons-cfg--sample
  namespace: default
  finalizers:
  - addons.kyma-project.io
spec:
  reprocessRequest: 0
  repositories:
    - url: https://github.com/kyma-project/bundles/releases/download/0.6.0/index.yaml
    - url: https://github.com/kyma-project/bundles/releases/download/0.6.0/index-testing.yaml
    - url: https://broker.url
status:
  phase: Failed
  lastProcessedTime: "2018-01-03T07:38:24Z"
  observedGeneration: 1
  repositories:
    - url: https://github.com/kyma-project/bundles/releases/download/0.6.0/index.yaml
      status: Ready
      addons:
        - name: gcp-service-broker
          version: 0.0.2
          status: Failed 
          reason: ConflictInSpecifiedRepositories
          message: "Specified repositories have addons with the same ID: [url: https://github.com/kyma-project/bundles/releases/download/0.6.0/index-testing.yaml, addons: testing:0.0.1]"
        - name: aws-service-broker
          version: 0.0.2
          status: Failed 
          reason: ConflictWithAlreadyRegisteredAddons
          message: "An addon with the same ID is already registered: [ConfigurationName: addons-cfg, url: https://github.com/kyma-project/bundles/releases/download/0.4.0/index.yaml, addons: aws-service-broker:0.0.1]"
        - name: azure-service-broker
          version: 0.0.1
          status: Ready
    - url: https://github.com/kyma-project/bundles/releases/download/0.6.0/index-testing.yaml
      status: Ready
      addons:
        - name: testing
          version: 0.0.1
          status: Failed
          reason: ConflictInSpecifiedRepositories
          message: "Specified repositories have addons with the same ID: [url: https://github.com/kyma-project/bundles/releases/download/0.6.0/index.yaml, addons: gcp-service-broker:0.0.2]"
        - name: redis
          version: 0.0.3
          status: Failed
          reason: ValidationError
          message: "Addon validation failed due to error: schema /plans/default/update-instance-schema.json is larger than 64 kB"
    - url: https://broker.url
      status: Failed
      reason: FetchingIndexError
      message: "Fetching repository failed due to error: the index file was not found"
```

> **NOTE:** If AddonsConfigurations/ClusterAddonsConfiguration is marked as `Failed` then all its addons are not available in Service Catalog.

## Custom resource parameters

This table lists all possible parameters of a given resource together with their descriptions:

| Parameter                              | Mandatory          | Description                                                                                                                                                        |
|----------------------------------------|:------------------:|--------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **metadata.name**                      | **YES**            | Specifies the name of the CR.                                                                                                                                      |
| **metadata.namespace**                 | **YES**            | Defines the Namespace in which the CR is available.                                                                                                                |
| **spec.reprocessRequest**              | **NO**             | Is a strictly increasing, non-negative integer counter that can be incremented by a user to manually trigger the reprocessing action of given CR.                  |
| **spec.repositories.url**              | **YES**            | Defines the full URL to the index file of addons repositories.                                                                                                     |
| **status.phase**                       | **Not applicable** | Describes the status of processing the CR by the Helm Broker Controller. It can be `Ready`, `Failed`, or `Pending`.                                                |
| **status.lastProcessedTime**           | **Not applicable** | Provides the last time when the Helm Broker Controller processed the CR.                                                                                           |
| **status.observedGeneration**          | **Not applicable** | Specifies the most recent generation that the Helm Broker Controller observes.                                                                                     |
| **status.repositories.url**            | **Not applicable** | Defines the full URL to the index file of addons repositories.                                                                                                     |
| **status.repositories.status**         | **Not applicable** | Describes the status of processing the given repository by the Helm Broker Controller.                                                                             |
| **status.repositories.reason**         | **Not applicable** | Provides the reason why the repository processing failed. [Here](../pkg/apis/addons/v1alpha1/reason.go) you can find all available reasons.                        |
| **status.repositories.message**        | **Not applicable** | Describes a human-readable message why processing failed. [Here](../pkg/apis/addons/v1alpha1/reason.go) you can find all available messages.                       |
| **status.repositories.addons.name**    | **Not applicable** | Defines the name of the addon.                                                                                                                                     |
| **status.repositories.addons.version** | **Not applicable** | Defines the version of the addon.                                                                                                                                  |
| **status.repositories.addons.status**  | **Not applicable** | Describes the status of processing the given addon by the Helm Broker Controller.                                                                                  |
| **status.repositories.addons.reason**  | **Not applicable** | Provides the reason why the addon processing failed. [Here](../pkg/apis/addons/v1alpha1/reason.go) you can find all available reasons.                             |
| **status.repositories.addons.message** | **Not applicable** | Describes a human-readable message on processing progress, success, or failure. [Here](../pkg/apis/addons/v1alpha1/reason.go) you can find all available messages. |

> **NOTE:** The Helm Broker Controller automatically adds all parameters marked as **Not applicable** to the AddonsConfiguration CR.

> **NOTE:** The namespace is discarded in case of the ClusterAddonsConfiguration. Rest parameters are the same.
