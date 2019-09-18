# Knative Serving

## Overview

This chart includes [knative-serving resources](https://github.com/knative/docs/tree/master/docs/serving) release files.

Included releases:
 * https://github.com/knative/serving/releases/download/v0.8.1/serving.yaml

Kyma-specific changes:
 * The `config-domain` is made configurable by specifying the `.Values.global.domainName` as the helm template.
 * The `knative-serving` Namespace is no longer created. This happens during the installation process.
 * The image versions are changed to use the release tag.
 * The `knative-ingress-gateway` is now a copy of `kyma-gateway`.
 * Changed CPU for minikube
