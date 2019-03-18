# Knative Eventing

## Overview

This chart includes [knative eventing](https://github.com/knative/docs/tree/master/docs/eventing) release files.

Included releases:
 * https://github.com/knative/eventing/releases/download/v0.4.1/release.yaml

Kyma-specific changes:
 * Every CRD has the `helm.sh/hook: crd-install` annotation set. This forces Helm to install the CRDs before other resources.
 * The `default-channel-config` is set to empty since when this chart is installed, we do not have a provisioner ready.
 * The creation of `knative-eventing` Namespace has been removed as this will taken care by installation process.
 * The `in-memory-channel` has been removed as Kyma use NATS Streaming based provisioner OOTB.
