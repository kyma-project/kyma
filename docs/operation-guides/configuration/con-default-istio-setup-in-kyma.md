---
title: Default Istio setup in Kyma
type: Configuration
---

Istio in Kyma is installed with the help of the `istioctl` tool.
The tool is driven by a configuration file containing an instance of [IstioOperator](https://istio.io/docs/reference/config/istio.operator.v1alpha1/) custom resource.
There are two configuration files — one for local installation on Minikube and one for cluster installations.
The configurations are customized for Kyma and are stored in the `resources/istio/files` directory.

## Istio components

This list shows the available Istio components and the components enabled in Kyma:

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

- Automatic sidecar injection is enabled by default, excluding the `istio-system` and `kube-system` Namespaces. You can [disable it](task-istio-disable-sidecar-injection).
- New resource requests for Istio sidecars are introduced: CPU: `20m`, memory: `32Mi`.
- New resource limits for Istio sidecars are introduced: CPU: `200m`, memory: `128Mi`.
- [Mutual TLS (mTLS)](https://istio.io/docs/concepts/security/#mutual-tls-authentication) is enabled cluster-wide in a STRICT mode.
- Global tracing is set to use the Zipkin installation provided by Kyma.
- Ingress Gateway is expanded to handle ports `80` and `443` for local Kyma deployments.
- DestinationRules are created by default, which disables mTLS for the `kubernetes.default.svc.cluster.local` service. In local (Minikube) installation mTLS is also disabled for
`istio-ingressgateway.istio-system.svc.cluster.local` service.
- The `istio-sidecar-injector` Mutating Webhook Configuration is patched to exclude Gardener resources in the kube-system namespace and the timeout is set to 10 seconds.
- All Istio components have decreased resource requests and limits.
- The use of HTTP 1.0 is enabled in the outbound HTTP listeners by `PILOT_HTTP10` flag set in Istiod component environment variables.
