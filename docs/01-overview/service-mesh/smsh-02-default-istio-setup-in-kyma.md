---
title: Default Istio setup in Kyma
---

Istio in Kyma is installed with the help of the `istioctl` tool. The tool is driven by a configuration file containing an instance of the [IstioOperator](https://istio.io/docs/reference/config/istio.operator.v1alpha1/) custom resource.

## Istio components

This list shows the available Istio components and addons. Check which of those are enabled in Kyma:
- Istiod (Pilot)
- Ingress Gateway
- Grafana - installed as separate component - [monitoring](../../05-technical-reference/00-architecture/obsv-01-architecture-monitoring.md)
- Prometheus - installed as separate component - [monitoring](../../05-technical-reference/00-architecture/obsv-01-architecture-monitoring.md)

## Kyma-specific configuration

These configuration changes are applied to customize Istio for use with Kyma:

- Both [Istio control plane and data plane](https://istio.io/latest/docs/ops/deployment/architecture/) use distroless images. To learn more, read about [Harden Docker Container Images](https://istio.io/latest/docs/ops/configuration/security/harden-docker-images/).
- Automatic sidecar injection is disabled by default. See how to [enable sidecar proxy injection](../../04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection.md).
- Resource requests and limits for Istio sidecars are modified to best suit the needs of the evaluation and production profiles.
- [Mutual TLS (mTLS)](https://istio.io/docs/concepts/security/#mutual-tls-authentication) is enabled cluster-wide in a STRICT mode.
- Ingress Gateway is expanded to handle ports `80`, `443`, and `31400` for local Kyma deployments.
- The use of HTTP 1.0 is enabled in the outbound HTTP listeners by `PILOT_HTTP10` flag set in Istiod component environment variables.
- IstioOperator configuration file is modified. [Change Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md) to customize the configuration.
