---
title: Default Istio setup in Kyma
---

Istio in Kyma is installed with the help of the `istioctl` tool.
The tool is driven by a configuration file containing an instance of [IstioOperator](https://istio.io/docs/reference/config/istio.operator.v1alpha1/) custom resource.

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

- Automatic sidecar injection is enabled by default, excluding the `istio-system` and `kube-system` Namespaces. You can [disable it](/../../04-operation-guides/operations/smsh-01-istio-disable-sidecar-injection.md.).
<!--- New resource requests for Istio sidecars are introduced: CPU: `20m`, memory: `32Mi`.
- New resource limits for Istio sidecars are introduced: CPU: `200m`, memory: `128Mi`. zrobić opisówkę (All Istio resources) Resources dla Istio componentów są zmienione, tune up dla ewaluacyjnego profilu i produkcyjnego-->
- [Mutual TLS (mTLS)](https://istio.io/docs/concepts/security/#mutual-tls-authentication) is enabled cluster-wide in a STRICT mode.
- Global tracing is set to use the Zipkin installation provided by Kyma.
<!--- Ingress Gateway is expanded to handle ports `80` and `443` for local Kyma deployments. consult Karol-->
<!--- DestinationRules are created by default, which disables mTLS for the `kubernetes.default.svc.cluster.local` service. In local (Minikube) installation mTLS is also disabled for valid bez Minikube'a-->
`istio-ingressgateway.istio-system.svc.cluster.local` service.
<!--- The `istio-sidecar-injector` Mutating Webhook Configuration is patched to exclude Gardener resources in the kube-system namespace and the timeout is set to 10 seconds. - invalid but consult -->
<!--- The use of HTTP 1.0 is enabled in the outbound HTTP listeners by `PILOT_HTTP10` flag set in Istiod component environment variables. -  consult Karol -->
- (Add info Nie wystawiamy calej konfiguracji istio operator file’a tylko zawezony set configuracji - dodac note'a o CHnage Kyma settings + wywalic doka "Provide custom Istio configuration"The Istio installation in Kyma uses the [IstioOperator](https://istio.io/docs/reference/config/istio.operator.v1alpha1/) API.
Kyma provides the default IstioOperator configurations for <!-- change "production and evaluation" local (Minikube) and cluster installations-->, <!--but you can add a custom IstioOperator definition that overrides the default settings. - czesciowo mozesz skonfigurowac Istio, use values.yam.