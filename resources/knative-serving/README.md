# Knative Serving

## Overview

This chart includes [knative-serving resources](https://github.com/knative/docs/tree/master/docs/serving) release files.

Included releases:
 * https://github.com/knative/serving/releases/download/v0.12.1/serving.yaml

Kyma-specific changes:
 * The `config-domain` is made configurable by specifying the `.Values.global.domainName` as the helm template.
 * The `knative-serving` Namespace is no longer created. This happens during the installation process.
 * The image versions are changed to use the release tag.
 * The `knative-ingress-gateway` is now a copy of `kyma-gateway`.
 * Changed CPU for minikube
 * Include [istio-knative-extras.yaml](https://github.com/knative/serving/blob/1cb31d16/third_party/istio-1.3.5/istio-knative-extras.yaml) which enables support for Knative Serving's `cluster-local` Gateway. This is required to create [private cluster-local Services](https://knative.dev/docs/serving/cluster-local-route/).
   * Comment all RBAC objects related to `istio-multi` and `istio-reader` ServiceAccounts, which are already part of the `istio` Chart.
   * Comment all `chart`, `heritage`, and `release` labels, which are leftovers from Helm Template.
 * A new label `kyma-project.io/dashboard: event-mesh` added for event-mesh dashboard
