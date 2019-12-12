---
title: Set MinIO to the AWS S3 Gateway mode
type: Tutorials
---

>**CAUTION:** AWS S3 Gateway mode was only tested manually on Kyma 1.6. Currently, there is no automated pipeline for it in Kyma.

By default, you install Kyma with the Asset Store in MinIO stand-alone mode. This tutorial shows how to set MinIO to the AWS S3 Gateway mode using an [override](/root/kyma/#configuration-helm-overrides-for-kyma-installation).

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Amazon Web Services (AWS)](https://aws.amazon.com) account

## Steps

You can set MinIO to the AWS S3 Gateway mode both during and after Kyma installation. In both cases, you need to create and configure an access key for the AWS account, apply a Secret and ConfigMap with an override onto a cluster or Minikube, and trigger the Kyma installation process.

>**CAUTION:** Buckets created in MinIO without using Bucket CRs are not recreated or migrated while switching to the MinIO Gateway mode.

### Create an AWS access key

Create an AWS access key for an IAM user. Follow these steps:

1. Access the [AWS Identity and Access Management console](https://console.aws.amazon.com/iam/)
2. In the left navigation panel, select **Users**.
3. Choose the user whose access keys you want to create, and select the **Security credentials** tab.
4. In the **Access keys** section, select **Create access key**.
5. Export the access and secret keys as environment variables:

    ```bash
    export AWS_ACCESS_KEY={your-access-ID}
    export AWS_SECRET_KEY={your-secret-key}
    ```

### Configure MinIO Gateway mode

Export an AWS S3 service [endpoint](https://docs.aws.amazon.com/general/latest/gr/rande.html) as an environment variable:

```bash
export AWS_SERVICE_ENDPOINT=https://{endpoint-address}
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
  minio.accessKey: "$(echo "${AWS_ACCESS_KEY}" | base64)"
  minio.secretKey: "$(echo "${AWS_SECRET_KEY}" | base64)"
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
  minio.s3gateway.enabled: "true"
  minio.s3gateway.serviceEndpoint: "${AWS_SERVICE_ENDPOINT}"
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
