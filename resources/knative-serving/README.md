# Knative Serving

## Overview

>**NOTE**: This is an experimental module.
This chart includes [knative-serving resources](https://github.com/knative/docs/tree/master/docs/serving) release files.

Included releases:
 * https://github.com/knative/serving/releases/download/v0.6.1/serving.yaml
 * https://github.com/knative/eventing/releases/download/v0.6.1/release.yaml

Kyma-specific changes:
 * The `config-domain` is made configurable by specifying the `.Values.global.domainName` as the helm template.
 * The `knative-serving` Namespace is no longer created. This happens during the installation process.
 * The image versions are changed to use the release tag.
 * Knative Serving uses the Kyma Istio Ingress gateway.
