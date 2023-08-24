---
title: Enable mTLS for custom metrics    
---

> **NOTE:** Prometheus and Grafana are [deprecated](https://github.com/kyma-project/website/blob/main/content/blog-posts/2022-12-09-monitoring-deprecation/index.md) and are planned to be removed. If you want to install a custom stack, take a look at [Install a custom kube-prometheus-stack in Kyma](https://github.com/kyma-project/examples/tree/main/prometheus).

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
        insecureSkipVerify: true  # Prometheus does not support Istio security naming; skip verifying the target Pod certificate
  namespaceSelector:
    any: true
```
