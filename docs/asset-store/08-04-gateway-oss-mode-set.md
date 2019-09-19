---
title: Set Minio to the Alibaba OSS Gateway mode
type: Tutorials
---

>**CAUTION:** Alibaba OSS Gateway Mode was tested only manually on Kyma 1.6. Currently, we don't have an automated pipeline like for other gateway modes.

By default, you install Kyma with the Asset Store in Minio stand-alone mode. This tutorial shows how to set Minio to the Alibaba Object Storage Service (OSS) Gateway mode using an [override](/root/kyma/#configuration-helm-overrides-for-kyma-installation).

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Alibaba Cloud](https://alibabacloud.com) account

## Steps

You can set Minio to the Alibaba OSS Gateway mode both during and after Kyma installation. In both cases, you need to create and configure a access key for Alibaba Cloud account, apply a Secret and Config Map with an override onto a cluster or Minikube, and trigger the Kyma installation process. 

>**CAUTION:** Buckets created in Minio without using Bucket CRs are not recreated or migrated while switching to the Minio Gateway mode.

### Create a Alibaba Cloud access key

Create a Alibaba Cloud access key for user. Follow these steps:

1. Access the [Resource Access Management console](https://ram.console.aliyun.com)
2. In the left navigation panel, choose **User**.
3. Choose the user whose access keys you want to create.
4. Click **Create AccessKey** in the **User AccessKey** section.
5. Export the access key and secret key as an environment variable:

    ```bash
    export ALIBABA_ACCESS_KEY={your-access-ID}
    export ALIBABA_SECRET_KEY={your-secret-key}
    ```

### Configure Minio Gateway mode

Export Alibaba OSS service endpoint (https://www.alibabacloud.com/help/doc-detail/31837.htm) as an environment variable:

```bash
export ALIBABA_SERVICE_ENDPOINT=https://{endpoint-address}
```

Apply the following Secret and ConfigMap with an override onto a cluster or Minikube. Run:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: asset-store-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: assetstore
    kyma-project.io/installation: ""
type: Opaque
data:
  minio.accessKey: "$(< ${ALIBABA_ACCESS_KEY} base64)"
  minio.secretKey: "$(< ${ALIBABA_SECRET_KEY} base64)"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: asset-store-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: assetstore
    kyma-project.io/installation: ""
data:
  minio.persistence.enabled: "false"
  minio.ossgateway.enabled: "true"
  minio.ossgateway.endpointURL: "${ALIBABA_SERVICE_ENDPOINT}"
  minio.DeploymentUpdate.type: RollingUpdate
  minio.DeploymentUpdate.maxSurge: "0"
  minio.DeploymentUpdate.maxUnavailable: "50%"
EOF
```
>**CAUTION:** When you install Kyma locally from sources, you need to manually add the ConfigMap and the Secret to the `installer-config-local.yaml.tpl` template located under the `installation/resources` subfolder before you run the installation script.

### Trigger installation

Trigger Kyma installation or update by labeling the Installation custom resource. Run:

```bash
kubectl label installation/kyma-installation action=install
```
