---
title: Monitoring in Istio
---

## Default setup

The [monitoring chart](https://github.com/kyma-project/kyma/blob/main/resources/monitoring/values.yaml) is preconfigured to collect all metrics relevant for observing the in-cluster [Istio](https://istio.io/latest/docs/concepts/observability/) Service Mesh, including the proxy-level, service-level, and control-plane metrics. The whole implementation of our monitoring solution is built around [Istio's observability best practices](https://istio.io/latest/docs/ops/best-practices/observability/).

![Prometheus Setup](./assets/prometheus-setup.svg)

1. The concept of collecting the [service-level](https://istio.io/latest/docs/concepts/observability/#service-level-metrics) metrics is based on the Istio Proxy implemented by Envoy. Istio Proxy collects all communication details inside the service mesh in a decentralized way. After scraping these high-cardinality metrics from the envoys, the metrics must be aggregated on a service level to get the final service-related details.

2. A dedicated Prometheus instance (Prometheus-Istio) scrapes and aggregates the service-level metrics. That instance is configured with the smallest possible data retention time because the raw metrics scraped from the Istio Proxies have high-cardinality and don't need to be kept further. 
The Istio-Prometheus instance is a Deployment named `monitoring-prometheus-istio-server`, with a hardcoded configuration that must not be changed. It also has no PersistentVolume attached. This instance never discovers additional metric endpoints from such resources as ServiceMonitors.

3. The main Prometheus instance scrapes these aggregated Istio metrics through the `/federate` endpoint of the Prometheus-Istio instance and any other metric endpoints from such resources as ServiceMonitors.
The main Prometheus instance supports scraping metrics using [`Strict mTLS`](https://istio.io/latest/docs/tasks/security/authentication/authn-policy/#globally-enabling-istio-mutual-tls-in-strict-mode). For this to work, Prometheus is configured to scrape metrics using Istio certificates.  
   
4. Prometheus is deployed with a sidecar proxy which rotates SDS certificates and outputs them to a volume mounted to the corresponding Prometheus container. It is configured to not intercept or redirect any traffic. 
   
5. By default, metrics from Kyma components are scraped using mTLS. As an exception, components deployed without sidecar proxy (for example, controllers like Prometheus operator) and Istio system components (for example the Istio sidecars proxies itself) are scraped using http, see also the [Istio's setup recommendation](https://istio.io/latest/docs/ops/integrations/prometheus/#tls-settings).

>**NOTE:** Learn how to [deploy](../../../03-tutorials/00-observability/obsv-01-observe-application-metrics.md#deploy-the-example-configuration) a sample `Go` service exposing metrics, which are scraped by Prometheus using mTLS.


>**NOTE:** You can find more information about the monitoring architecture in the [Technical References](../../../05-technical-reference/00-architecture/obsv-01-architecture-monitoring.md).

## Istio monitoring flow

See the diagram for a broader view of how the Istio-related instance fits into the monitoring setup in Kyma:

![Istio Monitoring](./assets/monitoring-istio.svg)

## Enable Grafana vizualization

By default, `monitoring-prometheus-istio-server` is not provided as a data source in Grafana.

1. To enable `monitoring-prometheus-istio-server` as a data source in Grafana, provide a YAML file with the following values:

  ```yaml
  monitoring  
    prometheus-istio:
      grafana:
        datasource:
          enabled: "true"
  ```

2. [Deploy](../../../04-operation-guides/operations/03-change-kyma-config-values.md) the values YAML file.

3. Restart the Grafana deployment with the following command:

  ```bash
  kubectl rollout restart -n kyma-system deployment monitoring-grafana
  ```

## Enable mTLS for custom metrics

To enable Strict mTLS scraping for a component, configure the Istio TLS certificates in the corresponding ServiceMonitor:

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
