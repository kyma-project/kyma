# Istio Kyma patch

## Overview

This chart packs the [patch script](../../components/istio-kyma-patch/README.md) as a Kubernetes job.

By default, following options are verified:
- The `policies.authentication.istio.io` CRD is required. This means that security in Istio must be enabled.
- mTLS must be enabled
- Policy checks must be enabled
- Automatic sidecar injector must be enabled

Changes moved into the Istio chart:
- Configuration of Zipkin as a tracer, done by official Istio values (`global.tracer.zipkin.address` in the `values.yaml` file)
- HostPorts in ingress-gateway deployment, done by editing the Istio chart (`istio/charts/gateways` `templates/deployment.yaml` and `values.yaml`)
- Configuration of the sidecar injector, done by official Istio values (`sidecarInjectorWebhook.enableNamespacesByDefault`)
- A DestinationRule CR disabling mTLS for requests to Helm Tiller.
- Monitoring and tracing related resources are deleted.
- Sidecar injection is enabled in all Namespaces, except those labeled with `istio-injection: disabled`.
- Chosen namespaces are labeled with `istio-injection: disabled`
- Istio prometheus is disabled by default
