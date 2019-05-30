# Istio Kyma patch

## Overview

This chart packs the [patch script](../../components/istio-kyma-patch/README.md) as a Kubernetes job.

By default, following changes are applied:
 * The `policies.authentication.istio.io` CRD is required. This means that security in Istio must be enabled.
 * Configuration of the [sidecar injector](../../components/istio-kyma-patch/README.md).
 * Monitoring and tracing related resources are deleted.
 * Sidecar injection is enabled in all Namespaces, except those labeled with `istio-injection: disabled`.
 * A DestinationRule CR disabling mTLS for requests to Helm Tiller.

Changes moved into the Istio chart:
- Configuration of Zipkin as a tracer, done by official Istio values (`global.tracer.zipkin.address` in the `values.yaml` file)
- HostPorts in ingress-gateway deployment, done by editing the Istio chart (`istio/charts/gateways` `templates/deployment.yaml` and `values.yaml`)

