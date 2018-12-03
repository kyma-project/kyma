# Knative

## Overview

This chart packs the [knative](https://github.com/knative/docs) release files.

Included releases:
 * https://github.com/knative/serving/releases/download/v0.2.1/release-no-mon.yaml
 * https://github.com/knative/eventing/releases/download/v0.2.0/release.yaml

Kyma-specific changes:
 * Every CRD has the `helm.sh/hook: crd-install` annotation set. This forces Helm to install the CRDs before other resources.
 * The duplicate of the `images.caching.internal.knative.dev` CRD is removed from the serving release.
 * If the **isLocalEnv** variable is set to `knative-ingressgateway`, the Service's type automatically changes to `NodePort`.
