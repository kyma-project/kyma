---
title: ClusterAddonsConfiguration
type: Custom Resource
---

The `clusteraddonsconfiguration.addons.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define cluster-wide bundles fetched by the Helm Broker. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd clusteraddonsconfiguration.addons.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample ClusterAddonsConfiguration which provides cluster-wide bundles. If any status of ClusterAddonsConfiguration is marked as `Failed`, all of its bundles are not available in the Service Catalog.

```yaml
apiVersion: addons.kyma-project.io/v1alpha1
kind: AddonsConfiguration
metadata:
  name: addons-cfg--sample
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

## Custom resource parameters

This table lists all possible parameters of a given resource together with their descriptions:

| Parameter                 | Mandatory          | Description                   |
|---------------------------|:------------------:|-------------------------------|
| **metadata.name**                      | **YES**            | Specifies the name of the CR.    |
| **spec.reprocessRequest**              | **NO**             | Is a strictly increasing, non-negative integer counter that can be incremented by a user to manually trigger the reprocessing action of given CR.    |
| **spec.repositories.url**              | **YES**            | Provides the full URL to the index file of addons repositories.    |
| **status.phase**                       | **Not applicable** | Describes the status of processing the CR by the Helm Broker Controller. It can be `Ready`, `Failed`, or `Pending`.       |
| **status.lastProcessedTime**           | **Not applicable** | Specifies the last time when the Helm Broker Controller processed the CR.     |
| **status.observedGeneration**          | **Not applicable** | Specifies the most recent generation that the Helm Broker Controller observed.               |
| **status.repositories.url**            | **Not applicable** | Provides the full URL to the index file of addons repositories.         |
| **status.repositories.status**         | **Not applicable** | Describes the status of processing a given repository by the Helm Broker Controller.     |
| **status.repositories.reason**         | **Not applicable** | Provides the reason why the repository processing failed. [Here](https://github.com/kyma-project/kyma/blob/master/components/helm-broker/pkg/apis/addons/v1alpha1/reason.go) you can find a complete list of reasons.     |
| **status.repositories.message**        | **Not applicable** | Provides a human-readable message why the repository processing failed. [Here](https://github.com/kyma-project/kyma/blob/master/components/helm-broker/pkg/apis/addons/v1alpha1/reason.go) you can find a complete list of messages.     |
| **status.repositories.addons.name**    | **Not applicable** | Defines the name of the addon.         |
| **status.repositories.addons.version** | **Not applicable** | Defines the version of the addon.        |
| **status.repositories.addons.status**  | **Not applicable** | Describes the status of processing a given addon by the Helm Broker Controller.           |
| **status.repositories.addons.reason**  | **Not applicable** | Provides the reason why the addon processing failed. [Here](https://github.com/kyma-project/kyma/blob/master/components/helm-broker/pkg/apis/addons/v1alpha1/reason.go) you can find a complete list of reasons.      |
| **status.repositories.addons.message** | **Not applicable** | Provides a human-readable message on processing progress, success, or failure. [Here](https://github.com/kyma-project/kyma/blob/master/components/helm-broker/pkg/apis/addons/v1alpha1/reason.go) you can find a complete list of messages. |

> **NOTE:** The Helm Broker Controller automatically adds all parameters marked as **Not applicable** to the ClusterAddonsConfiguration CR.

## Related resources and components

These are the resources related to this CR:

| Custom resource |   Description |
|-----------------|---------------|
| {Related CRD kind} |  {Briefly describe the relation between the resources}. |

These components use this CR:

| Component   |   Description |
|-------------|---------------|
| Helm Broker |  Fetches cluster-wide bundles provided by this CR. |
