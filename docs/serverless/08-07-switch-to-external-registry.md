---
title: Switch to an external Docker registry in the runtime
type: Tutorials
---

This tutorial shows how you can [switch to an external Docker registry](#details-internal-and-external-registries-switching-registries-in-the-runtime) in a specific Namespace, with Serverless already installed on your cluster. This example shows the `default` Namespace but you can use any other. You will create a Secret custom resource (CR) with credentials to one these registry:

- [Docker Hub](https://hub.docker.com/)
- [Google Container Registry (GCR)](https://cloud.google.com/container-registry)
- [Azure Container Registry (ACR)](https://azure.microsoft.com/en-us/services/container-registry/)

Any Function deployed in the `default` Namespace will use these credentials to pull the images from the selected external registry. They will also allow Kaniko - the job build tool - to push any newly built or rebuilt images to this registry.

> **CAUTION:** Function images are not cached in the Docker Hub. The reason is that this registry is not compatible with the caching logic defined in [Kaniko](https://cloud.google.com/cloud-build/docs/kaniko-cache) that Serverless uses for building images.

## Prerequisites

<div tabs name="prerequisites" group="external-docker-registry">
  <details>
  <summary label="docker-hub">
  Docker Hub
  </summary>

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

  </details>
  <details>
  <summary label="gcr">
  GCR
  </summary>

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [gcloud](https://cloud.google.com/sdk/gcloud/)
- [Google Cloud Platform (GCP)](https://cloud.google.com) project

  </details>
  <details>
  <summary label="acr">
  ACR
  </summary>

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure)
- [Microsoft Azure](http://azure.com) subscription

  </details>
</div>

## Steps

### Create required cloud resources

To create cloud resources required for a given registry provider, follow the steps described in the [Set an external Docker registry](#tutorials-set-an-external-docker-registry-create-required-cloud-resources) tutorial.

### Create a Secret CR

Create a Secret CR in the `default` Namespace. Such a Secret must contain the `serverless.kyma-project.io/remote-registry: config` label and the required data (**username**, **password**, **serverAddress**, and **registryAddress**):

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
type: kubernetes.io/dockerconfigjson
metadata:
 name: serverless-registry-config
 namespace: default
 labels:
   serverless.kyma-project.io/remote-registry: config
data:
 username: {VALUE}
 password: {VALUE}
 serverAddress: {VALUE}
 registryAddress: {VALUE}
EOF
```

> **CAUTION:** If you want to create a cluster-wide Secret, you must create it in the `kyma-system` Namespace and add the `serverless.kyma-project.io/config: credentials` label. Read more about [requirements for Secret CRs](#details-switching-registries-in-the-runtime).

### Test the registry switch

[Create a Function](#tutorials-create-an-inline-function) in the `default` Namespace and check if the Function's Deployment points to the external registry using this command:

```bash
kubectl get pods -n default -l serverless.kyma-project.io/resource=deployment -o jsonpath='{ ...image }'
```
