# Knative Eventing

## Overview

This chart includes [knative eventing](https://github.com/knative/docs/tree/master/docs/eventing) release files.

Included releases:
 * https://github.com/knative/eventing/releases/download/v0.5.0/release.yaml

Kyma-specific changes:
 * The `default-channel-config` is left empty since when this chart is installed, there is no provisioner ready.
 * The `knative-eventing` Namespace is no longer created. This happens during the installation process.
 * The `in-memory-channel` no longer exists, as Kyma uses NATS Streaming-based provisioner out of the box.
 * The image versions are changed to use the release tag.
