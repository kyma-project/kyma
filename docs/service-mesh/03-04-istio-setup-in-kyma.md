---
title: Istio setup in Kyma
type: Details
---

Istio is installed using the official charts from the currently supported release. The charts are customized for Kyma and are stored in the `resources/istio` directory.

## Istio components

This list shows the available Istio components and the components enabled by default:

| Component | Enabled? |
| :--- | :---: |
| Gateways | ✅ |
| Sidecar Injector | ✅ |
| Galley | ✅ |
| Mixer | ✅ |
| Pilot | ✅ |
| Security | ✅ |
| Node agent | ⛔️ |
| Grafana | ⛔️ |
| Prometheus | ⛔️ |
| Servicegraph | ⛔️ |
| Tracing | ⛔️ |
| Kiali | ⛔️ |

## Kyma-specific configuration

These configuration changes are applied to customize Istio for use with Kyma:

- Only the Ingress Gateway is enabled.
- The [Secret Discovery Service](https://www.envoyproxy.io/docs/envoy/latest/configuration/secret#config-secret-discovery-service) is enabled in the Ingress Gateway.
- Automatic sidecar injection is enabled by default, excluding the `istio-system` and `kube-system` Namespaces.
- New resource limits for Istio sidecars are introduced: CPU: `100m`, memory: `128Mi`.
- [Mutual TLS (mTLS)](https://istio.io/docs/concepts/security/#mutual-tls-authentication) is enabled cluster-wide with the exception of the Istio Control Plane.  
- Global tracing is set to use the Zipkin installation provided by Kyma.
- Ingress Gateway is expanded to handle ports `80` and `443` for local Kyma deployments.
- DestinationRules are created by default, which disables mTLS for the `istio-ingressgateway.istio-system.svc.cluster.local` service.

## The Customization sub-chart

The Kyma-specific configuration is applied through the proprietary `customization` sub-chart added to the official Istio charts. The sub-chart is located in the `resources/istio/charts/customization` directory.
