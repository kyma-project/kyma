---
title: AddonsConfiguration
type: Custom Resource
---

The `addonsconfiguration.addons.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define define Namespace-scoped addons fetched by the Helm Broker. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd addonsconfiguration.addons.kyma-project.io -o yaml
```

> **NOTE:** Only users with the **kyma-admin** role can modify the AddonsConfiguration CR. To learn more about roles in Kyma, read [this](/components/security/#details-roles-in-kyma) document.

## Sample custom resource

This is a sample AddonsConfiguration which provides Namespace-scoped addons. If any of the **status** fields of the CR is marked as `Failed`, none of the addons registered with the CR is available in the Service Catalog.

>**NOTE:** All CRs must have the `addons.kyma-project.io` finalizer which prevents the CR from deletion until the Controller completes the deletion logic successfully. If you don't set a finalizer, the Controller sets it automatically.

```yaml
apiVersion: addons.kyma-project.io/v1alpha1
kind: AddonsConfiguration
metadata:
  name: addons-cfg-sample
  namespace: default
  finalizers:
  - addons.kyma-project.io
  label:
spec:
  reprocessRequest: 0
  repositories:
    - url: https://github.com/kyma-project/addons/releases/download/0.6.0/index.yaml
    - url: https://github.com/kyma-project/addons/releases/download/0.6.0/index-testing.yaml
    - url: https://broker.url
status:
  phase: Failed
  lastProcessedTime: "2018-01-03T07:38:24Z"
  observedGeneration: 1
  repositories:
    - url: https://github.com/kyma-project/addons/releases/download/0.6.0/index.yaml
      status: Ready
      addons:
        - name: gcp-service-broker
          version: 0.0.2
          status: Failed
          reason: ConflictInSpecifiedRepositories
          message: "Specified repositories have addons with the same ID: [url: https://github.com/kyma-project/addons/releases/download/0.6.0/index-testing.yaml, addons: testing:0.0.1]"
        - name: aws-service-broker
          version: 0.0.2
          status: Failed
          reason: ConflictWithAlreadyRegisteredAddons
          message: "An addon with the same ID is already registered: [ConfigurationName: addons-cfg, url: https://github.com/kyma-project/addons/releases/download/0.4.0/index.yaml, addons: aws-service-broker:0.0.1]"
        - name: azure-service-broker
          version: 0.0.1
          status: Ready
    - url: https://github.com/kyma-project/addons/releases/download/0.6.0/index-testing.yaml
      status: Ready
      addons:
        - name: testing
          version: 0.0.1
          status: Failed
          reason: ConflictInSpecifiedRepositories
          message: "Specified repositories have addons with the same ID: [url: https://github.com/kyma-project/addons/releases/download/0.6.0/index.yaml, addons: gcp-service-broker:0.0.2]"
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

>**NOTE:** The Controller fetches and processes all addons, even if any of them fails. Thanks to that, at the end of the process you can see the status of all processed addons. You can read information about all detected problems in the **status** entry of a given CR.

## Custom resource parameters

This table lists all possible parameters of a given resource together with their descriptions:

| Parameter                              | Required          | Description            |
|----------------------------------------|:------------------:|------------------------|
| **metadata.name**                      | Yes            | Specifies the name of the CR.         |
| **metadata.namespace**                 | Yes            | Specifies the Namespace in which the CR is available.        |
| **metadata.finalizers**                | Yes            | Specifies the finalizer which prevents the CR from deletion until the Controller completes the deletion logic. The default finalizer is `addons.kyma-project.io`.       |
| **metadata.labels**                    | No            | Specifies a key-value pair that helps you to organize and filter your CRs. The label indicating the default addon configuration is `addons.kyma-project.io/managed: "true"`.       |
| **spec.reprocessRequest**              | No             | Allows you to manually trigger the reprocessing action of this CR. It is a strictly increasing, non-negative integer counter.   |
| **spec.repositories.url**              | Yes            | Provides the full URL to the index file of addons repositories.    |
| **spec.repositories.secretRef.name**     | No           | Defines the name of a Secret which provides values for the URL template.    |
| **spec.repositories.secretRef.namespace**| No           | Defines the Namespace which stores a Secret that provides values for the URL template.    |
| **status.phase**                       | Not applicable | Describes the status of processing the CR by the Helm Broker Controller. It can be `Ready`, `Failed`, or `Pending`.       |
| **status.lastProcessedTime**           | Not applicable | Specifies the last time when the Helm Broker Controller processed the CR.     |
| **status.observedGeneration**          | Not applicable | Specifies the most recent generation that the Helm Broker Controller observed.               |
| **status.repositories.url**            | Not applicable | Provides the full URL to the index file with addons definitions.         |
| **status.repositories.status**         | Not applicable | Describes the status of processing a given repository by the Helm Broker Controller.     |
| **status.repositories.reason**         | Not applicable | Provides the reason why the repository processing failed. [Here](https://github.com/kyma-project/helm-broker/blob/master/pkg/apis/addons/v1alpha1/reason.go) you can find a complete list of reasons.     |
| **status.repositories.message**        | Not applicable | Provides a human-readable message why the repository processing failed. [Here](https://github.com/kyma-project/helm-broker/blob/master/pkg/apis/addons/v1alpha1/reason.go) you can find a complete list of messages.     |
| **status.repositories.addons.name**    | Not applicable | Defines the name of the addon.         |
| **status.repositories.addons.version** | Not applicable | Defines the version of the addon.        |
| **status.repositories.addons.status**  | Not applicable | Describes the status of processing a given addon by the Helm Broker Controller.           |
| **status.repositories.addons.reason**  | Not applicable | Provides the reason why the addon processing failed. [Here](https://github.com/kyma-project/helm-broker/blob/master/pkg/apis/addons/v1alpha1/reason.go) you can find a complete list of reasons.      |
| **status.repositories.addons.message** | Not applicable | Provides a human-readable message on processing progress, success, or failure. [Here](https://github.com/kyma-project/helm-broker/blob/master/pkg/apis/addons/v1alpha1/reason.go) you can find a complete list of messages. |

> **NOTE:** The Helm Broker Controller automatically adds all parameters marked as **Not applicable** to the AddonsConfiguration CR.

## Related resources and components

These components use this CR:

| Component   |   Description |
|-------------|---------------|
| Helm Broker |  Fetches Namespace-scoped addons provided by this CR. |
