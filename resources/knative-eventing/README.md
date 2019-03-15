# Knative Eventing

## Overview

This chart packs the [knative eventing](https://github.com/knative/docs/tree/master/eventing) release files.

Included releases:
 * https://github.com/knative/eventing/releases/download/v0.4.1/release.yaml

Kyma-specific changes:
 * Every CRD has the `helm.sh/hook: crd-install` annotation set. This forces Helm to install the CRDs before other resources.
