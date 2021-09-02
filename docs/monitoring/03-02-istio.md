---
title: Istio monitoring
type: Details
---

The monitoring chart is pre-configured to collect all metrics relevant for observing the in-cluster [Istio](https://istio.io/latest/docs/concepts/observability/) Service Mesh, including the proxy-level, service-level, and control-plane metrics.

The concept of collecting the [service-level](https://istio.io/latest/docs/concepts/observability/#service-level-metrics) metrics is built around the Istio Proxy implemented by Envoy. Istio Proxy collects all communication details inside the service mesh in a decentralized way. After scraping these high cardinality metrics from the envoys, the metrics need to be additionally aggregated on a service level to get the final service-related details.

Following the [Istio's observability best practice](https://istio.io/latest/docs/ops/best-practices/observability/), the scraping and aggregation of the service-level metrics is done in a dedicated Prometheus instance. That instance has the smallest possible data retention time configured as the raw metrics scraped from the Istio Proxies have high cardinality and are not further required to be kept. Instead, the main Prometheus instance scrapes the aggregated metrics through the `/federate` endpoint.

The Istio-related instance is a Deployment named `monitoring-prometheus-istio-server`. This instance has a short data retention time and hardcoded configuration that should not be changed. It also has no PersistentVolume attached. This instance never discovers additional metric endpoints from such resources as ServiceMonitors.

The monitoring chart is configured in such way, that it is possible to scrape metrics using [`Strict mTLS`](https://istio.io/latest/docs/tasks/security/authentication/authn-policy/#globally-enabling-istio-mutual-tls-in-strict-mode). For this to work, Prometheus is configured to scrape metrics using Istio certificates. Prometheus is deployed with a sidecar proxy which rotates SDS certificates and outputs them to a volume mounted to the corresponding Prometheus container. To stick to Istio's observability best practices, Prometheus redirection is configured to not intercept or redirect any traffic.

If a component wants to make use of Strict mTLS scraping, a Strict PeerAuthentication policy has to be added:

```yaml
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: sample-metrics
spec:
  selector:
    matchLabels:
      app: sample-metrics
  mtls:
    mode: "STRICT"
```

Furthermore, the corresponding ServiceMonitor needs to have the Istio TLS certificates:

```yaml
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: metrics
  namespace: kyma-system
  labels:
    prometheus: monitoring
    example: monitoring-custom-metrics
spec:
  selector:
    matchLabels:
      k8s-app: metrics
  targetLabels:
    - k8s-app
  endpoints:
    - port: web
      interval: 10s
      scheme: https
      tlsConfig: 
        caFile: /etc/prometheus/secrets/istio.default/root-cert.pem
        certFile: /etc/prometheus/secrets/istio.default/cert-chain.pem
        keyFile: /etc/prometheus/secrets/istio.default/key.pem
        insecureSkipVerify: true  # Prometheus does not support Istio security naming, thus skip verifying target pod ceritifcate
  namespaceSelector:
    any: true
```

See the diagram for a broader view of how the Istio-related instance fits into the monitoring setup in Kyma:

![Istio Monitoring](./assets/monitoring-istio.svg)

By default, `monitoring-prometheus-istio-server` is not provided as a data source in Grafana. However, this can be enabled by adding the override: 

 ```bash
cat <<EOF | kubectl apply -f -
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: monitoring-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: monitoring
    kyma-project.io/installation: ""
data:
    prometheus-istio.grafana.datasource.enabled: "true"
EOF
```

Run the [cluster update process](/root/kyma/#installation-update-kyma). After finishing the upgrade process, restart the Grafana deployment by using this command:

```bash
kubectl rollout restart -n kyma-system deployment monitoring-grafana
```
