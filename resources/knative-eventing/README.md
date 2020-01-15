# Knative Eventing

## Overview

This chart includes [knative eventing](https://github.com/knative/docs/tree/master/docs/eventing) release files.

Included releases:
 * https://github.com/knative/eventing/releases/download/v0.10.0/eventing.yaml

Kyma-specific changes:
 * The `knative-eventing` Namespace is no longer created. This happens during the installation process.
 * The `in-memory-channel` no longer exists, as Kyma uses NATS Streaming-based provisioner out of the box.
 * The image versions are changed to use the release tag.
 * Added requests.memory, requests.cpu, limits.cpu and limits.memory for deployment/eventing-controller, deployment/webhook (values motivated from knative/serving charts)
 * Removed istio-proxy side-car for eventing-controller
 * Configured NATSS as default ClusterChannelProvisioner in `default-channel-webhook` ConfigMap
 * Configured config-tracing as per Kyma setup
 * A new label `kyma-project.io/event-mesh: "true"` added for event-mesh dashboard