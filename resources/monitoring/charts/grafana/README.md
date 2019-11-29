# Grafana

## Overview

Kyma comes with a set of dashboards for monitoring Kubernetes clusters. These dashboards display metrics that the Prometheus server collects.

In Kyma, you can find these dashboards under [grafana](../templates/grafana/).

Some of the available dashboards:

* **Istio** - Displays Istio metrics for services (HTTP and TCP) as well as the Service Mesh. Find the configuration of this dashboard in [this](../../templates/grafana/kyma-dashboards/istio-mesh.yaml) file.
* **Mixer** -Displays metrics on incoming requests, response durations, connections, cluster membership, server error rate and more. Find the configuration of this dashboard in [this](../../templates/grafana/kyma-dashboards/istio-mixer.yaml) file.
* **Pilot** - Displays metrics on request latency, discovery calls and various cache types. Find the configuration of this dashboard in [this](../../templates/grafana/kyma-dashboards/istio-pilot.yaml) file.
* **Nodes** - Displays information pertaining to Kubernetes nodes utilization. Find the configuration of this dashboard in [this](../../templates/grafana/dashboards/nodes.yaml) file.
* **Pods** - Displays Pod metrics such as CPU and Memory. Find the configuration of this dashboard in [this](../../templates/grafana/dashboards/pods.yaml) file.
* **StatefulSet** - Displays Kubernetes StatefulSet metrics such as replica count, CPU and Memory. Find the configuration of this dashboard in [this](../../templates/grafana/dashboards/statefulset.yaml) file.

## Add a dashboard to Kyma

Grafana dashboards in Kyma are configured through ConfigMaps.

This is how a dashboard looks like:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: dashboard-name
  labels:
    {{- if $.Values.grafana.sidecar.dashboards.label }}
    {{ $.Values.grafana.sidecar.dashboards.label }}: "1"
    {{- end }}
    app: {{ template "prometheus-operator.name" $ }}-grafana
data:
  dashboard.json: |-
    {
      ... dashboard configuration as JSON
    }
```

To add a dashboard to Kyma:

1. Create or modify the dashboard using the Grafana UI.
2. Export the dashboard to a JSON format file. Create a ConfigMap file having this JSON in the `data` field. Name the file following this convention: `{dashboard_name}.yaml`.
4. Clone the Kyma [master branch](https://github.com/kyma-project/kyma).
5. Copy the YAML file to the directory **[kyma-dashboards](../../templates/grafana/kyma-dashboards/)** of your local installation.
6. Install Kyma locally and open it in a browser at https://console.kyma.local.
7. Access the Grafana console from Kyma by clicking **Administration > Diagnostics > Status & Metrics** in the left navigation.  
8. Sign in and check if the newly added dashboard is deployed.  
9. Create a pull request following the [GitHub workflow](https://github.com/kyma-project/community/blob/master/contributing/03-git-workflow.md) defined for Kyma.

## Add a Custom Dashboard in Grafana

Users can create their own **Grafana Dashboard** by using the Grafana UI as the dashboards are persisted even after the Pod restarts.

1. Create or modify a dashboard using Grafana UI.
2. Save the dashboard with a new name.

## Lambda dashboard

The lambda dashboard provides visualisation for specific lambda function metrics such as memory usage, CPU usage or success rate response.

You can access the dashboard directly from the lambda UI.

## Unique Dashboard Identifier

The Unique Dashboard Identifier or UID allows having consistent URLs for accessing dashboards from the lambda UI. 
This UID is defined in the dashboard YAML files.

>**Note:** Changing the UID breaks the URL used to access specific dashboards from the lambda UI.

## Additional Resources

There are several resources you can use to become more familiar with Grafana. The [Grafana Getting Started Guide](http://docs.grafana.org/guides/getting_started/) is an ideal starting point. Refer to the document [Export and Import Dashboards](http://docs.grafana.org/reference/export_import/) for a closer look at dashboards used to export and import data in Grafana. Grafana also provides in-depth documentation on the [Grafana Dashboard API](http://docs.grafana.org/http_api/dashboard/).
