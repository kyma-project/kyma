# Knative

## Overview

This chart packs the [knative](https://github.com/knative/docs) release files.

Included releases:
 * https://github.com/knative/serving/releases/download/v0.2.1/release-no-mon.yaml
 * https://github.com/knative/eventing/releases/download/v0.2.0/release.yaml

Our changes:
 * Every CRD have annotation `"helm.sh/hook": "crd-install"` so Helm installs it before rest of resources.
 * There is duplicated CRD `images.caching.internal.knative.dev` in serving release. One copy is removed.
 * `knative-ingressgateway` Service's `type` is changed to `NodePort`.
 * `knative-shared-gateway` Gateway have TLS enabled
 * KNative resources are created only if variable `global.knative` is true