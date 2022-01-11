---
title: Default Istio setup in Kyma
---

Istio in Kyma is installed with the help of the `istioctl` tool. The tool is driven by a configuration file containing an instance of the [IstioOperator](https://istio.io/docs/reference/config/istio.operator.v1alpha1/) custom resource.

## Istio components

This list shows the available Istio components and addons. Check which of those are enabled in Kyma:

| Component | Enabled |
| :--- | :---: |
| Istiod | ✅ |
| Ingress Gateway | ✅️ |
| Egress Gateway | ⛔️ |
| CNI | ⛔️ |
| Grafana | ⛔️ |
| Prometheus | ⛔️ |
| Tracing | ⛔️ |
| Kiali | ⛔️ |

## Kyma-specific configuration

These configuration changes are applied to customize Istio for use with Kyma:

- Both [Istio control plane and data plane](https://istio.io/latest/docs/ops/deployment/architecture/) use distroless images compliant with Federal Information Processing Standards (FIPS). [Solo.io](https://www.solo.io/) provides the FIPS-certified images. To learn more, read about [Distroless FIPS-compliant Istio](https://www.solo.io/blog/distroless-fips-compliant-istio/).
- Automatic sidecar injection is enabled by default, excluding the `istio-system` and `kube-system` Namespaces. See how to [disable sidecar proxy injection](../../../04-operation-guides/operations/smsh-01-istio-disable-sidecar-injection.md).
- Resource requests and limits for Istio sidecars are modified to best suit the needs of the evaluation and production profiles.
- [Mutual TLS (mTLS)](https://istio.io/docs/concepts/security/#mutual-tls-authentication) is enabled cluster-wide in a STRICT mode.
- Global tracing is set to use the Zipkin protocol to send requests to the tracing component provided by Kyma. Learn more about [tracing]](../../../05-technical-reference/00-architecture/obsv-03-architecture-tracing.md).
- Ingress Gateway is expanded to handle ports `80`, `443`, and `31400` for local Kyma deployments.
- The `istio-sidecar-injector` Mutating Webhook Configuration is patched to exclude Gardener resources in the `kube-system` Namespace and the timeout is set to 10 seconds.
- The use of HTTP 1.0 is enabled in the outbound HTTP listeners by `PILOT_HTTP10` flag set in Istiod component environment variables.
- IstioOperator configuration file is modified. [Change Kyma settings](../../../04-operation-guides/operations/03-change-kyma-config-values.md) to customize the configuration.
