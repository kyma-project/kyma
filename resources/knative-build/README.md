# Knative build

## Overview

This chart packs the [Knative build](https://github.com/knative/build) release files.

Included releases:
 * https://github.com/knative/build/releases/download/v0.5.0/build.yaml
 * https://raw.githubusercontent.com/knative/serving/v0.5.1/third_party/config/build/clusterrole.yaml

Kyma-specific changes:
 * Every CRD has the `helm.sh/hook: crd-install` annotation set. This forces Helm to install the CRDs before other resources.
 * The duplicate of the `images.caching.internal.knative.dev` CRD is removed from the release.
 * The creation of the `knative-build` namespace is removed from the release. The namespace will already be created by the Kyma installer during installation process.
 * The image versions are changed to use the release tag.