# Grafana

## Overview

Kyma comes with a set of dashboards for monitoring Kubernetes clusters. These dashboards display metrics that the Prometheus server collects.

In Kyma, you can find these dashboards under [dashboards](dashboards/).

These are the available dashboards:

* **Deployment** - Displays metrics on details such as memory, CPU, network and replicas for deployments running in different namespaces. Find the configuration of this dashboard in [this](dashboards/deployment-dashboard.json) file.
* **Istio** - Displays Istio metrics for services (HTTP and TCP) as well as the Service Mesh. Find the configuration of this dashboard in [this](dashboards/istio-mesh-dashboard.json) file.
* **Kubernetes Capacity Planning** - Displays the current memory usage, disk usage, system load, and other system status details. Find the configuration of this dashboard in [this](dashboards/kubernetes-capacity-planning-dashboard.json) file.
* **Kubernetes Cluster Health** - Displays the status of alerts, nodes, pods and control plan components. Find the configuration of this dashboard in [this](dashboards/kubernetes-cluster-health-dashboard.json) file.
* **Kubernetes Cluster Status** - Displays metrics on alerts, API servers, CPU utilitzation, schedulers, and more. Find the configuration of this dashboard in [this](dashboards/kubernetes-cluster-status-dashboard.json) file.
* **Kubernetes Control Plane Status** - Displays metrics on control planes. Find the configuration of this dashboard in [this](dashboards/kubernetes-control-plane-status-dashboard.json) file.
* **Kubernetes Resource Requests** - Displays details on CPU core and memory resource usage. Find the configuration of this dashboard in [this](dashboards/kubernetes-resource-requests-dashboard.json) file.
* **Mixer** -Displays metrics on incoming requests, response durations, connections, cluster membership, server error rate and more. Find the configuration of this dashboard in [this](dashboards/istio-mixer-dashboard.json) file.
* **Nodes** - Displays information pertaining to Kubernetes nodes utilization. Find the configuration of this dashboard in [this](dashboards/nodes-dashboard.json) file.
* **Pilot** - Displays metrics on request latency, discovery calls and various cache types. Find the configuration of this dashboard in [this](dashboards/istio-pilot-dashboard.json) file.
* **Pods** - Displays Pod metrics such as CPU and Memory. Find the configuration of this dashboard in [this](dashboards/pods-dashboard.json) file.
* **StatefulSet** - Displays Kubernetes StatefulSet metrics such as replica count, CPU and Memory. Find the configuration of this dashboard in [this](dashboards/statefulset-dashboard.json) file.

## Add a dashboard to Kyma

Grafana dashboards in Kyma are configured through a [ConfigMap](templates/dashboards-configmap.yaml). This dashboard consumes the data configuration of all the JSON files located in the [dashboards directory](dashboards/).

```yaml
data:
  ...
  {{- if .Values.keepOriginalDashboards }}
{{ (.Files.Glob "dashboards/*.json").AsConfig | indent 2 }}
  {{- end }}
```

To add a dashboard to Kyma:

1. Create or modify the dashboard using the Grafana UI.
2. Export the dashboard to a JSON format file. Name the file following this convention: `{dashboard_name}-dashboard.json`.
3. Clone the Kyma [master branch](https://github.com/kyma-project/kyma).
4. Copy the JSON file to the directory **[dashboards](dashboards/)** of your local installation.
5. Install Kyma locally and open it in a browser at https://console.kyma.local.
6. Access the Grafana console from Kyma by clicking **Administration > Diagnostics > Status & Metrics** in the left navigation.  
7. Sign in and check if the newly added dashboard is deployed.  
8. Create a pull request following the [GitHub workflow](https://github.com/kyma-project/community/blob/master/contributing/03-git-workflow.md) defined for Kyma.

## Add a custom dashboard in Grafana

Users can create their own **Grafana Dashboard** by using the Grafana UI as the dashboards are persisted even after the Pod restarts.

1. Create or modify a dashboard using Grafana UI.
2. Save the dashboard with a new name.

## Lambda dashboard

The lambda dashboard provides visualisation for specific lambda function metrics such as memory usage, CPU usage or success rate response.

You can access the dashboard directly from the lambda UI.

## Unique Dashboard Identifier

The Unique Dashboard Identifier or UID allows having consistent URLs for accessing dashboards from the lambda UI. 
This UID is defined in the `*-dashboard.json` file.

>**Note:** Changing the UID breaks the URL used to access specific dashboards from the lambda UI.

## Additional resources

There are several resources you can use to become more familiar with Grafana. The [Grafana Getting Started Guide](http://docs.grafana.org/guides/getting_started/) is an ideal starting point. Refer to the document [Export and Import Dashboards](http://docs.grafana.org/reference/export_import/) for a closer look at dashboards used to export and import data in Grafana. Grafana also provides in-depth documentation on the [Grafana Dashboard API](http://docs.grafana.org/http_api/dashboard/).


