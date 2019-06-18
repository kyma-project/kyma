---
title: Istio customization
type: Details
---

Istio is installed using the official charts from the currently supported release. However, those charts are customized for Kyma
Istio installs using the official, charts from the currently supported release. The charts that are currently
used are stored in the `resources/istio` directory.

## Istio components

| Component name | Enabled? | 
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

## Istio configuration 
List of configuration changes to Istio in Kyma:
- Only ingressgatewy is enabled
- Secret Discovery Service ([SDS](https://www.envoyproxy.io/docs/envoy/latest/configuration/secret#config-secret-discovery-service)) in ingressgateway is enabled
- Automatic sidecar injection is enabled by default, but the following namespaces are excluded:
    + `istio-system`
    + `kube-system`
- New resource limits for istio sidecars is introduced (CPU: 100m, memory: 128Mi)
- Mutual TLS ([mTLS](https://istio.io/docs/concepts/security/#mutual-tls-authentication)) is enabled cluster wide
- Global tracing is set to use the Zipkin installation provided by Kyma (`zipkin.kyma-system`)
- Ingressgatewy is expanded to handle `hostPorts 80, 443` in the case of a local (minikube) installation
- DestinationRules are created by default, which disable mTLS for the following services:
    + `istio-ingressgateway.istio-system.svc.cluster.local`
    + `tiller-deploy.kube-system.svc.cluster.local`

### Customization subchart
Part of the configuration is done in a private `customization` subchart, which is added to the official Istio charts. The component can be found in `resources/istio/charts/customization`. 