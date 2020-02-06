---
title: Configure default Knative Channel
type: Details
---

## Overview

Kyma comes with NATS Streaming as its default Channel. You can see the configuration details in the [`default-ch-webhook`](../../resources/knative-eventing/charts/knative-eventing/templates/eventing.yaml) ConfigMap.

You can use a different messaging middleware, other than NATS Streaming, as the Kyma eventing operator.
To achieve that:

- Apply the Channel Resources for the messaging middleware you want to use. These resources connect with the running messaging middleware.
- Configure the `default-ch-webhook` ConfigMap in the `knative-eventing` Namespace to use that particular Channel.

If you want to edit the `default-ch-webhook` ConfigMap, run: 

```bash
kubectl -n knative-eventing edit configmaps default-ch-webhook
```

Read about the examples and the configuration details.

## In-memory channel

Follow this [guide](https://github.com/knative/eventing/tree/master/config/channels/in-memory-channel) to add the InMemoryChannel resources. This will deploy the InMemoryChannel CRD, Controller, and Dispatcher.

>**NOTE**: Before installing this provisioner, add the following annotation to the [`podTemplate.Spec`](https://github.com/knative/eventing/blob/master/config/channels/in-memory-channel/300-in-memory-channel.yaml) in the `in-memory-channel-controller` Deployment to remove the Istio sidecar.

```yaml
template:
  annotations:
    sidecar.istio.io/inject: "false"
    metadata:
      labels: *labels
```

You can change the default channel configuration by editing the ConfigMap `default-ch-webhook` in `knative-eventing` Namespace. For example, if you want to set In-Memory Channels as default provisioner, include the following data in the ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: default-ch-webhook
  namespace: knative-eventing
data:
  default-ch-config: |
    clusterDefault:
      apiVersion: messaging.knative.dev/v1alpha1
      kind: InMemoryChannel
    namespaceDefaults:
      some-namespace:
        apiVersion: messaging.knative.dev/v1alpha1
        kind: InMemoryChannel
```

> **NOTE**: This ConfigMap may specify a cluster-wide default channel and/or namespace-specific channel implementations.

## Google PubSub

Follow this [guide](https://github.com/google/knative-gcp/blob/master/docs/install/install-knative-gcp.md) to install Knative-GCP along with GCP PubSub Channel CRDs and deploy the `cloud-run-events` controller.

Edit  the `default-ch-webhook` ConfigMap located in the `knative-eventing` Namespace to include the following data.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: default-ch-webhook
  namespace: knative-eventing
data:
  default-ch-config: |
    clusterDefault:
      apiVersion: messaging.cloud.run/v1alpha1
      kind: Channel
      spec:
        project: <GCP Project Name>
```

> **NOTE**: You need to mention the GCP Project Name in the specification which will be used as the reference GCP project to create GCP PubSub Topics.
