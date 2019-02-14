---
title: Installation with Knative
type: Installation
---

You can install Kyma with [Knative](https://cloud.google.com/knative/) and use its solutions for handling events and serverless functions.

> **NOTE:** You can’t install Kyma with Knative on clusters with a pre-allocated ingress gateway IP address.

> **NOTE:** Knative intagration requires Kyma 0.6 or higher.

## Knative with local deployment from release

When you install Kyma locally from a release, follow [this](#installation-install-kyma-locally-from-the-release-install-kyma-on-minikube) guide. 
Ensure that you created the local Kubernetes cluster with `10240Mb` memory and `30Gb` disk size.
```
./scripts/minikube.sh --domain "kyma.local" --vm-driver "hyperkit" --memory 10240Mb --disk-size 30g
```

Run the following command before triggering the Kyma installation process:
```
kubectl -n kyma-installer patch configmap installation-config-overrides -p '{"data": {"global.knative": "true", "global.kymaEventBus": "false", "global.natsStreaming.clusterID": "knative-nats-streaming"}}'
```

## Knative with local deployment from sources

When you install Kyma locally from sources, add the `--knative` argument to the `run.sh` script. Run this command:

```
./run.sh --knative
```

## Knative with a cluster deployment

Run the following command before triggering the Kyma installation process:
```
kubectl -n kyma-installer patch configmap installation-config-overrides -p '{"data": {"global.knative": "true", "global.kymaEventBus": "false", "global.natsStreaming.clusterID": "knative-nats-streaming"}}'
```
