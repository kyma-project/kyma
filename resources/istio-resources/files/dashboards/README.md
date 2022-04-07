# Istio Dashboards
Istio Dashboards in Kyma are based on the official [Istio dashboards](https://istio.io/latest/docs/ops/integrations/grafana/#configuration) with minor modifications:
 - Removed `DS_PROMETHEUS` Grafana variable
 - Removed **__inputs** and **__requires** fields
 - Added tags: **"tags": ["service-mesh", "kyma"],**
