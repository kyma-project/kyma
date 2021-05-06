---
title: Switching registries at runtime
type: Details
---

When you install Kyma with the default internal registry, Helm creates the `serverless-registry-config-default` Secret in the `kyma-system` Namespace. This Secret contains credentials used by Kubernetes to pull deployed Functions' images from the internal registry. These credentials also allow Kaniko to push any images to the registry each time a Function is created or updated.

> **NOTE:** If you [install Serverless with overrides](#tutorials-set-an-external-docker-registry), you disable the internal registry and specify an external one to use. The `serverless-registry-config-default` Secret will then contain credentials to the specified external registry instead of the internal one.

Once you have Serverless up and running, you can switch to an external registry:

- Per Namespace, and have even multiple external registries in a cluster, but no more than one per Namespace.
- Cluster-wide, with this configuration overwriting by default the Namespace-scoped one.

## Namespace-scoped external registry

To switch to an external registry in a given Namespace, create a Secret CR that:

- Is named `serverless-registry-config`.
- Is of type `kubernetes.io/dockerconfigjson`.
- Has the `serverless.kyma-project.io/remote-registry: config` label.
- Contains these keys with valid values pointing to the external registry:
  - **username**
  - **password**
  - **serverAddress**
  - **registryAddress**

See this example:

  ```yaml
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
  ```

  > **CAUTION:** If you have your own Secret CR in a Namespace and you don't want the system to override it with any cluster-wide configuration, add the `serverless.kyma-project.io/managed-by: user` label to that Secret CR.

## Cluster-wide external registry

To switch to one external registry in the whole cluster, you must create a Secret CR in the `kyma-system` Namespace. The Secret CR must meet the same [requirements](#namespace-scoped-external-registry) as in the case of the Namespace-scoped setup, but you must also add the `serverless.kyma-project.io/config: credentials` label. This label ensures the Secret CR gets propagated to all Namespaces. Such a cluster-wide configuration will take precedence over a Namespace-scoped one unless the Namespace-scoped configuration blocks it with the `serverless.kyma-project.io/managed-by: user` label.

> **CAUTION:** Do not remove the `serverless.kyma-project.io/config: credentials` label from the existing Secret CR in the `kyma-system` Namespace. If you do so, you will not be able to remove the Secret CR afterwards.

### How this works

This implementation has a fallback mechanism that works as follows:

1. Every 5 minutes, Function Controller checks if there is the `serverless-registry-config` Secret CR in the Namespace with your Function specifying alternative registry to push the Function's image to.
2. If it doesn't find such a Secret CR, Function Controller uses the credentials to the default registry specified in the `serverless-registry-config-default` Secret CR.

This mechanism also leaves room for a lot of flexibility as you can easily switch between external registries or move back to the internal one. If you remove the `serverless-registry-config` Secret CR or update it with credentials to a different external registry, you don't lose any images. Function Controller detects any changes in the Secret CR and the images are rebuilt automatically, using cache and delta updates. If you modify the username and password to the registry, the [admission webhook](#details-supported-webhooks-admission-webhook) automatically encodes these credentials to base64 and sets them as a value under the `.dockerconfigjson` entry in the Secret CR. These credentials later serve Kubernetes to pull images of deployed Function from the registry, and allow Kaniko to push any newly built or rebuilt images to this registry.
