# Knative Eventing

## Overview

This chart includes [knative eventing](https://github.com/knative/docs/tree/master/docs/eventing) release files.

Included releases:
 * https://github.com/knative/eventing/releases/download/v0.12.0/eventing.yaml

Kyma-specific changes:
 * The `knative-eventing` Namespace is no longer created. This happens during the installation process.
 * The `in-memory-channel` no longer exists, as Kyma uses NATS Streaming-based provisioner out of the box.
 * The image versions are changed to use the release tag.
 * Added memory and cpu requests/limits to the `eventing-controller` and `sources-controller` Deployments.
 * Configured NATSS as default ClusterChannelProvisioner in `default-channel-webhook` ConfigMap
 * Configured config-tracing as per Kyma setup
 * A new label `kyma-project.io/dashboard: event-mesh` added for event-mesh dashboard
