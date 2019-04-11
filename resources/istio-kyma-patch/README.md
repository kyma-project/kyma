# Istio Kyma patch

## Overview

This chart packs the [patch script](../../components/istio-kyma-patch/README.md) as a Kubernetes job.

By default, following changes are applied:
 * The `policies.authentication.istio.io` CRD is required. This means that security in Istio must be enabled.
 * Configuration of the [sidecar injector](../../components/istio-kyma-patch/README.md).
 * The Egressgateway, Ingressgateway, policy, and telemetry are configured to use Zipkin from the `kyma-system` Namespace.
 * Monitoring and tracing related resources are deleted.
 * Sidecar injection is enabled in all Namespaces, except those labeled with `istio-injection: disabled`.
 * A DestinationRule CR disabling mTLS for requests to Helm Tiller.
