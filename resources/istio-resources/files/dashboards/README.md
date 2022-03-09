# Istio Dashboards
Dashboards are form official Istio dashboards with minor changes: https://istio.io/latest/docs/ops/integrations/grafana/#configuration
 - removing DS_PROMETHEUS Grafana variable
 - removing `__inputs` and `__requires` fields
 - adding tags: `"tags": ["service-mesh", "kyma"],`
