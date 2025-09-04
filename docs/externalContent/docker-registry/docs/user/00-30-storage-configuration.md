# Docker Registry Storage Configuration

The DockerRegistry CR allows you to store images in five ways: filesystem, Azure, s3, GCP, and BTP Object Store. This document describes how to configure DockerRegistry CR to cooperate with all these storage types.

## Filesystem

The filesystem storage is a built-in storage type based on the PersistentVolumeClaim CR, which is part of the Kubernetes functionality. This is a default DockerRegistry CR configuration, and no additional configuration is needed.

All images pushed to this storage are removed when the Docker Registry is uninstalled/reconfigured, or the cluster is removed. Stored images can't be shared between clusters.

### Sample CR

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
    name: default
    namespace: kyma-system
spec: {}
```

## Azure

The Azure Storage can be configured in the DockerRegistry **spec.storage.azure** field. The only thing that is required is the **secretName** field, which must contain the name of the Secret with Azure configuration located in the same namespace. The Secret must have the following values:

* **accountKey** - contains the key used to authenticate to the Azure Storage
* **accountName** - contains the name used to authenticate to the Azure Storage
* **container** - contains the name of the storage container

The images can be stored centrally and shared between clusters so that different registries can reuse specific layers or whole images. Images will not be removed after deleting the cluster or uninstalling the Docker Registry module.

### Sample CR

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
    name: default
    namespace: kyma-system
spec:
    storage:
        azure:
            secretName: azure-storage
```

### Sample Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: azure-storage
  namespace: kyma-system
data:
  accountKey: "YWNjb3VudEtleQ=="
  accountName: "YWNjb3VudE5hbWU="
  container: "Y29udGFpbmVy"
```

## s3

Similarly to Azure, the s3 storage can be configured in the DockerRegistry **spec.storage.s3** field. The only required fields are **bucket**, which contains the s3 bucket name, and **region**, which specifies the bucket location. This storage type allows you to provide additional optional configuration, described in [DockerRegistry CR](resources/06-20-docker-registry-cr.md). One of the optional configurations is the **secretName** that contains the authentication method to the s3 storage in the following format:

* **accountKey** - contains the key used to authenticate to the s3 storage
* **secretKey** - contains the name used to authenticate to the s3 storage

### Sample CR

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
  name: default
  namespace: kyma-system
spec:
  storage:
    s3:
      bucket: "bucketName"
      region: "eu-central-1"
      regionEndpoint: "s3-eu-central-1.amazonaws.com"
      encrypt: false
      secure: true
      secretName: "s3-storage"
```

### Sample Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: s3-storage
  namespace: kyma-system
data:
  accessKey: "YWNjZXNzS2V5"
  secretKey: "c2VjcmV0S2V5"
```

## Google Cloud Storage

Google Cloud Storage (GCS) can be configured using the **spec.storage.gcs** field. The only required field is the **bucket**, which contains the GCS bucket name. This storage type allows you to provide additional optional configuration described in [DockerRegistry CR](resources/06-20-docker-registry-cr.md). One of the optional configurations is the **secretName**, which contains the authentication method to the GCS, which is a private service account key in the JSON format.

### Sample Custom Resource

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
  name: default
  namespace: kyma-system
spec:
  storage:
    gcs:
      bucket: "bucketName"
      secretName: "gcs-secret"
      rootdirectory: "dir"
      chunkSize: 5242880
```

### Sample Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gcs-secret
  namespace: kyma-system
data:
  accountkey: "Z3Njc2VjcmV0"
```

## BTP Object Store

BTP Object Store can be configured using the **spec.storage.btpObjectStore** field. The only required field is the **secretName**, which contains the BTP Object Store Secret name.
The Secret is provided to an instance of BTP Object Store by a service binding. The underlying object store depends on the hyperscaler used for the BTP subaccount, AWS or GCP.
Azure hyperscaler is not supported.

### Sample Custom Resource

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
  name: default
  namespace: kyma-system
spec:
  storage:
    btpObjectStore:
      secretName: "btp-object-store-secret"
```

## PVC storage

PVC storage can be configured using the **spec.storage.pvc** field. The only required field is the **name**, which contains the PersistentVolumeClaim name.

### Sample Custom Resource

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: DockerRegistry
metadata:
  name: default
  namespace: kyma-system
spec:
  storage:
    pvc:
      name: "existing-pvc"
```
