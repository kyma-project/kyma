# Knative

## Overview

This chart packs the [knative](https://github.com/knative/docs) release files.

Included releases:
 * https://github.com/knative/serving/releases/download/v0.4.0/serving.yaml
 * https://github.com/knative/eventing/releases/download/v0.4.0/release.yaml

Kyma-specific changes:
 * Every CRD has the `helm.sh/hook: crd-install` annotation set. This forces Helm to install the CRDs before other resources.
 * The duplicate of the `images.caching.internal.knative.dev` CRD is removed from the serving release.

> **NOTE:** The Knative build component is not installed.
