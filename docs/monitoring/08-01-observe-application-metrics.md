---
title: Observe application metrics
type: Tutorials
---

This tutorial shows how you can check the list and changing values of all metrics exposed by a sample Go service by redirecting the metrics to a localhost and the default Prometheus server port.

This tutorial uses the [`monitoring-custom-metrics`](https://github.com/kyma-project/examples/tree/master/monitoring-custom-metrics) example and one of its services named `sample-metrics-8081` service. The service exposes its metrics on the standard `/metrics` endpoint that is available under port `8081`. You deploy the service `deployment/deployment.yaml` along with the service monitor `deployment/service-monitor.yaml` that instructs Prometheus to pull metrics:
- Of the service with the `k8s-app: metrics` label
- From the `/metrics` endpoint
- At a `10s` interval

This tutorial focuses on the `cpu_temperature_celsius` metric that is one of the custom metrics exposed by the `sample-metrics-8081` service. Using this metric logic implemented in the example, you can observe how the CPU temperature changes from 60 to 90 degrees each time the Prometheus makes a call to the `/metrics` endpoint.

## Prerequisites

To complete the tutorial you must meet one of these prerequisites and have:
- Cluster with Kyma 1.3 or higher
- Local Kyma 1.3 or higher installation, with the Monitoring component installed

> **NOTE:** The Monitoring component is not installed by default as part of the Kyma Lite package.

## Steps

Follow these steps to:
- Deploy the sample service with its default configuration
- Redirect the metrics to a localhost
- Redirect the metrics to the Prometheus server to observe the metrics in the Prometheus UI
- Clean up the deployed example

### Deploy the example configuration

Follow these steps:

1. Create the `testing-monitoring` Namespace.

```
kubectl create namespace testing-monitoring
```

2. Deploy the sample service in the `testing-monitoring` Namespace.

```
kubectl create -f https://raw.githubusercontent.com/kyma-project/examples/master/monitoring-custom-metrics/deployment/deployment.yaml --namespace=testing-monitoring
```

3. Deploy the service monitor in the `kyma-system` Namespace that is a default Namespace for all Service Monitors.

```

kubectl apply -f https://raw.githubusercontent.com/kyma-project/examples/master/monitoring-custom-metrics/deployment/service-monitor.yaml
```
3. Test your deployment.

```
kubectl get pods -n testing-monitoring
```

You should get a result similar to this one:

```
NAME                              READY   STATUS    RESTARTS   AGE
sample-metrics-6f7c8fcf4b-mlgbx   2/2     Running   0          26m
```

### View metrics on a localhost

Follow these steps:

1. Run the `port-forward` command on the `sample-metrics-8081` service for port `8081` to check the metrics.

```
kubectl port-forward svc/sample-metrics-8081 -n testing-monitoring 8081:8081

```

2. Open a browser and access [`http://localhost:8081/metrics`](http://localhost:8081/metrics).

You can see the `cpu_temperature_celsius` metric and its current value on the list of all metrics exposed by the `sample-metrics-8081` service.

![metrics on port 8081](./assets/sample-metrics-2.png)

Thanks to the example logic, the custom metric value changes each time you refresh the localhost address.

### View metrics on the Prometheus UI

You can also observe the metric on the Prometheus UI and see how its value changes in the pre-defined `10s` interval in which Prometheus pulls the metric value from the service endpoint.

Follow these steps:

1. Run the `port-forward` command on the `monitoring-prometheus` service:

```bash
kubectl port-forward svc/monitoring-prometheus -n kyma-system 9090:9090

```

2. Access the the [Prometheus UI](http://localhost:9090/targets#job-sample-metrics-8081) service endpoint and its details on the **Targets** list.

![Prometheus Dashboard](./assets/pm-dashboard-1.png)

2. Open the **Graph** tab, search for the `cpu_temperature_celsius` metric in the **Expression** search box, and click the **Execute** button to check the last value pulled by Prometheus.

![Prometheus Dashboard](./assets/pm-dashboard-2.png)

The Prometheus UI shows a new value every 10 seconds upon refreshing the page.

### Clean up the configuration

When you finish the tutorial, remove the deployed example and all its resources from the cluster.

Follow these steps:

1. Remove the deployed Service Monitor from the `kyma-system` Namespace.

    ```bash
    kubectl delete servicemonitor -l example=monitoring-custom-metrics -n kyma-system
    ```

2. Remove the example Deployment from the `testing-monitoring` Namespace.

    ```bash
    kubectl delete all -l example=monitoring-custom-metrics -n testing-monitoring
    ```
