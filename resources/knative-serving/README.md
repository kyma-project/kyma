# Knative Serving

## Overview

This chart includes [knative-serving](https://github.com/knative/docs/tree/master/docs/serving) release files.

Included releases:
 * https://github.com/knative/serving/releases/download/v0.4.1/serving.yaml
 * https://github.com/knative/eventing/releases/download/v0.4.1/release.yaml

Kyma-specific changes:
 * Every CRD has the `helm.sh/hook: crd-install` annotation set. This forces Helm to install the CRDs before other resources.
 * The duplicate of the `images.caching.internal.knative.dev` CRD is removed from the serving release.
 * The `config-domain` is made configurable by specifying the `.Values.domainName` as the helm template.

> **NOTE:** The Knative build component is not installed.
