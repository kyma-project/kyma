# Docker Registry Custom Resource

The `dockerregistries.operator.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the Docker Registry configuration that you want to install on your cluster. To get the up-to-date CRD and show the output in the YAML format, run this command:

   ```bash
   kubectl get crd dockerregistries.operator.kyma-project.io -o yaml
   ```

## Sample Custom Resource

The following Docker Registry custom resource (CR) shows the configuration of the Docker Registry.

   ```yaml
   apiVersion: operator.kyma-project.io/v1alpha1
   kind: DockerRegistry
   metadata:
     annotations:
       kubectl.kubernetes.io/last-applied-configuration: |
         {"apiVersion":"operator.kyma-project.io/v1alpha1","kind":"DockerRegistry","metadata":{"annotations":{},"name":"default","namespace":"kyma-system"},"spec":{}}
     creationTimestamp: "2024-05-16T10:18:25Z"
     finalizers:
     - dockerregistry-operator.kyma-project.io/deletion-hook
     generation: 1
     name: default
     namespace: kyma-system
     resourceVersion: "31542"
     uid: 30dbb8a0-2193-47b6-bdf7-358f78319eb8
   spec: {}
   status:
     conditions:
     - lastTransitionTime: "2024-05-16T10:18:25Z"
       message: Configuration ready
       reason: Configured
       status: "True"
       type: Configured
     - lastTransitionTime: "2024-05-16T10:18:45Z"
       message: DockerRegistry installed
       reason: Installed
       status: "True"
       type: Installed
     storage: filesystem
     externalAccess:
       enabled: "False"
     internalAccess:
       enabled: "True"
       pullAddress: localhost:32137
       pushAddress: dockerregistry.kyma-system.svc.cluster.local:5000
       secretName: dockerregistry-config
     served: "True"
     state: Ready
   ```

## Custom Resource Parameters

For details, see the [Docker Registry specification file](https://github.com/kyma-project/docker-registry/blob/main/components/operator/api/v1alpha1/dockerregistry_types.go).
<!-- TABLE-START -->
### dockerregistries.operator.kyma-project.io/v1alpha1

**Spec:**

| Parameter                               | Type   | Description                                                                                                                |
|-----------------------------------------|--------|----------------------------------------------------------------------------------------------------------------------------|
| **externalAccess**                      | object | Contains configuration of the registry external access through the Istio Gateway.                                          |
| **externalAccess.enabled**              | string | Specifies if the registry is exposed.                                                                                      |
| **externalAccess.gateway**              | string | Specifies the name of the Istio Gateway CR in the `NAMESPACE/NAME` format. Defaults to the `kyma-system/kyma-gateway`.     |
| **externalAccess.host**                 | string | Specifies the host on which the registry will be exposed. It must fit into at least one server defined in the Gateway.     |
| **storage**                             | object | Contains configuration of the registry images storage.                                                                     |
| **storage.deleteEnabled**               | string | Specifies if registry supports deletion of image blobs and manifests by digest.                                            |
| **storage.azure**                       | object | Contains configuration of the Azure Storage.                                                                               |
| **storage.azure.secretName** (required) | string | Specifies the name of the Secret that contains data needed to connect to the Azure Storage.                                |
| **storage.s3**                          | object | Contains configuration of the s3 storage.                                                                                  |
| **storage.s3.bucket** (required)        | string | Specifies the name of the s3 bucket.                                                                                       |
| **storage.s3.region** (required)        | string | Specifies the region of the s3 bucket.                                                                                     |
| **storage.s3.regionEndpoint**           | string | Specifies the endpoint of the s3 region.                                                                                   |
| **storage.s3.encrypt**                  | string | Specifies if data in the bucket is encrypted.                                                                              |
| **storage.s3.secure**                   | string | Specifies if registry uses the TLS communication with the s3.                                                              |
| **storage.s3.secretName**               | string | Specifies the name of the Secret that contains data needed to connect to the s3 storage.                                   |
| **storage.gcs.bucket** (required)       | string | Specifies the name of the GCS bucket.                                                                                      |
| **storage.gcs.secretName**              | string | A private service account key file in JSON format used for Service Account Authentication.                                 |
| **storage.gcs.rootdirectory**           | string | The root directory tree in which all registry files are stored. Defaults to the empty string (bucket root).                |
| **storage.gcs.chunksize**               | string | This is the chunk size used for uploading large blobs, must be a multiple of 256*1024. Defaults to 5242880.                |
| **storage.btpObjectStore.secretName**   | string | Specifies the name of the Secret that contains data needed to connect to BTP Object Store.                                 |
| **storage.pvc.name** (required)         | string | Specifies the name of the PersistentVolumeClaim.                                                                           |


**Status:**

| Parameter                                            | Type       | Description                                                                                                                                                                                                                                                                                                                                                    |
|------------------------------------------------------|------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **conditions**                                       | \[\]object | Conditions associated with CustomStatus.                                                                                                                                                                                                                                                                                                                       |
| **conditions.&#x200b;lastTransitionTime** (required) | string     | Specifies the last time the condition transitioned from one status to another. This should be when the underlying condition changes.  If that is not known, then using the time when the API field changed is acceptable.                                                                                                                                      |
| **conditions.&#x200b;message** (required)            | string     | Provides a human-readable message indicating details about the transition. This may be an empty string.                                                                                                                                                                                                                                                        |
| **conditions.&#x200b;observedGeneration**            | integer    | Represents **.metadata.generation** that the condition was set based upon. For instance, if **.metadata.generation** is currently `12`, but the **.status.conditions[x].observedGeneration** is `9`, the condition is out of date with respect to the current state of the instance.                                                                           |
| **conditions.&#x200b;reason** (required)             | string     | Contains a programmatic identifier indicating the reason for the condition's last transition. Producers of specific condition types may define expected values and meanings for this field and whether the values are considered a guaranteed API. The value should be a camelCase string. This field may not be empty.                                        |
| **conditions.&#x200b;status** (required)             | string     | Specifies the status of the condition. The value is either `True`, `False`, or `Unknown`.                                                                                                                                                                                                                                                                      |
| **conditions.&#x200b;type** (required)               | string     | Specifies the condition type in camelCase or in `foo.example.com/CamelCase`. Many **.conditions.type** values are consistent across resources like `Available`, but because arbitrary conditions can be useful (see **.node.status.conditions**), the ability to deconflict is important. The regex it matches is `(dns1123SubdomainFmt/)?(qualifiedNameFmt)`. |
| **storage**                                          | string     | Type of the used registry images storage.                                                                                                                                                                                                                                                                                                                      |
| **internalAccess**                                   | object     | Contains installed internal access configuration.                                                                                                                                                                                                                                                                                                              |
| **internalAccess.enabled**                           | string     | Specifies if internal access is enabled.                                                                                                                                                                                                                                                                                                                       |
| **internalAccess.secretName**                        | string     | Name of the Secret with data needed for internal connection to Docker Registry.                                                                                                                                                                                                                                                                                |
| **internalAccess.pushAddress**                       | string     | Address that can be used to push images from inside the cluster.                                                                                                                                                                                                                                                                                               |
| **internalAccess.pullAddress**                       | string     | Address that can be used by Kubernetes to make a communication with the registry.                                                                                                                                                                                                                                                                              |
| **externalAccess**                                   | object     | Contains installed external access configuration.                                                                                                                                                                                                                                                                                                              |
| **externalAccess.enabled**                           | string     | Specifies if external access is enabled.                                                                                                                                                                                                                                                                                                                       |
| **externalAccess.gateway**                           | string     | Specifies the name of the Istio Gateway CR.                                                                                                                                                                                                                                                                                                                    |
| **externalAccess.secretName**                        | string     | Name of the Secret with data needed for external connection to Docker Registry.                                                                                                                                                                                                                                                                                |
| **externalAccess.pushAddress**                       | string     | Address that can be used to push images from outside the cluster.                                                                                                                                                                                                                                                                                              |
| **externalAccess.pullAddress**                       | string     | Address that can be used by Kubernetes to make a communication with the registry.                                                                                                                                                                                                                                                                              |
| **served** (required)                                | string     | Signifies if the current Docker Registry is managed. Value can be `True` or `False`.                                                                                                                                                                                                                                                                        |
| **state**                                            | string     | Signifies the current state of Docker Registry. Value can be one of `Ready`, `Processing`, `Error`, or `Deleting`.                                                                                                                                                                                                                                                  |

<!-- TABLE-END -->

### Status Reasons

Processing of a Docker Registry CR can succeed, continue, or fail for one of these reasons:

## Docker Registry CR Conditions

This section describes the possible states of the Docker Registry CR. Three condition types, `Installed`, `Configured` and `Deleted`, are used.

| No  | CR State   | Condition type | Condition status | Condition reason | Remark                                             |
|-----|------------|----------------|------------------|------------------|----------------------------------------------------|
| 1   | Processing | Configured     | true             | Configured       | Docker Registry configuration verified             |
| 2   | Processing | Configured     | unknown          | Configuration    | Docker Registry configuration verification ongoing |
| 3   | Error      | Configured     | false            | ConfigurationErr | Docker Registry configuration verification error   |
| 4   | Error      | Configured     | false            | Duplicated       | Only one Docker Registry CR is allowed             |
| 5   | Ready      | Installed      | true             | Installed        | Docker Registry workloads deployed                 |
| 6   | Processing | Installed      | unknown          | Installation     | Deploying Docker Registry workloads                |
| 7   | Error      | Installed      | false            | InstallationErr  | Deployment error                                   |
| 8   | Deleting   | Deleted        | unknown          | Deletion         | Deletion in progress                               |
| 9   | Deleting   | Deleted        | true             | Deleted          | Docker Registry module deleted                     |
| 10  | Error      | Deleted        | false            | DeletionErr      | Deletion failed                                    |
