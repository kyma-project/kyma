# Integrate With Loki

## Overview

| Category| |
| - | - |
| Signal types | logs |
| Backend type | custom in-cluster |
| OTLP-native | yes |

Learn how to use [Loki](https://github.com/grafana/loki/tree/main/production/helm/loki) in [OTLP mode](https://grafana.com/docs/loki/latest/send-data/otel/) as a logging backend with Kyma's [LogPipeline](../../logs.md).

> [!WARNING]
> This guide uses Grafana Loki, which is distributed under [AGPL-3.0](https://github.com/grafana/loki/blob/main/LICENSE) only. Using components that have this license might affect the license of your project. Inform yourself about the license used by Grafana Loki under [https://grafana.com/licensing/](https://grafana.com/licensing/)).

![setup](./../assets/loki.drawio.svg)

## Table of Content

- [Prerequisites](#prerequisites)
- [Preparation](#preparation)
- [Loki Installation](#loki-installation)
- [Kyma Telemetry Integration](#kyma-telemetry-integration)
- [Grafana installation](#grafana-installation)
- [Grafana Exposure](#grafana-exposure)

## Prerequisites

- Kyma as the target deployment environment
- The [Telemetry module](https://kyma-project.io/#/telemetry-manager/user/README) is [added](https://kyma-project.io/#/02-get-started/01-quick-install)
- [Kubectl version that is within one minor version (older or newer) of `kube-apiserver`](https://kubernetes.io/releases/version-skew-policy/#kubectl)
- Helm 3.x

## Preparation

1. Export your namespace as a variable with the following command:

    ```bash
    export K8S_NAMESPACE="loki"
    ```

2. Export the Helm release names that you want to use. It can be any name, but be aware that all resources in the cluster will be prefixed with that name. Run the following command:

    ```bash
    export HELM_LOKI_RELEASE="loki"
    ```

3. Update your Helm installation with the required Helm repository:

    ```bash
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo update
    ```

## Loki Installation

Depending on your scalability needs and storage requirements, you can install Loki in different [Deployment modes](https://grafana.com/docs/loki/latest/fundamentals/architecture/deployment-modes/). The following instructions install Loki in a lightweight in-cluster solution that does not fulfil production-grade qualities. Consider using a scalable setup based on an object storage backend instead (see [Install the simple scalable Helm chart](https://grafana.com/docs/loki/latest/installation/helm/install-scalable/)).

### Install Loki

You install the Loki stack with a Helm upgrade command, which installs the chart if not present yet.

```bash
helm upgrade --install --create-namespace -n ${K8S_NAMESPACE} ${HELM_LOKI_RELEASE} grafana/loki \
  -f https://raw.githubusercontent.com/grafana/loki/main/production/helm/loki/single-binary-values.yaml \
  --set-string 'loki.podLabels.sidecar\.istio\.io/inject=true' \
  --set 'singleBinary.resources.requests.cpu=1' \
  --set 'loki.auth_enabled=false' \
  --set 'log.storage.type: filesystem'
```

The previous command uses an example [values.yaml](https://github.com/grafana/loki/blob/main/production/helm/loki/single-binary-values.yaml) from the Loki repository for setting up Loki in the 'SingleBinary' mode. Additionally, it applies:

- Istio sidecar injection for the Loki instance
- a reduced CPU request setting for smaller cluster setups
- disabled multi-tenancy for easier setup
- enabled filesystem storage for a demo setup (not recommended)

Alternatively, you can create your own `values.yaml` file and adjust the command.

### Verify Loki Installation

Check that the `loki` Pod has been created in the Namespace and is in the `Running` state:

```bash
kubectl -n ${K8S_NAMESPACE} get pod -l app.kubernetes.io/name=loki
```

## Kyma Telemetry Integration

### Create a LogPipeline Resource

To ingest the application logs from within your cluster to Loki, use Kyma's LogPipeline feature based on OTLP. It enables the `application` input to automatically tail application logs from `stdout/stderr`. If you want to push logs natively, additionally use the [log gateway endpoint](https://kyma-project.io/#/telemetry-manager/user/logs?id=_1-create-a-logpipeline).

Apply the LogPipeline:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: telemetry.kyma-project.io/v1alpha1
   kind: LogPipeline
   metadata:
     name: custom-loki
   spec:
      input:
         application:
            enabled: true
      output:
         otlp:
            protocol: http
            endpoint:
               value: http://${HELM_LOKI_RELEASE}.${K8S_NAMESPACE}.svc.cluster.local:3100
            path: otlp
   EOF
   ```

When the status of the applied LogPipeline resource turns to `Running`, the underlying collector is reconfigured and log shipment to your Loki instance is active.

### Verify the Setup by Accessing Logs Using the Loki API

1. To access the Loki API, use kubectl port forwarding. Run:

   ```bash
   kubectl -n ${K8S_NAMESPACE} port-forward svc/$(kubectl  get svc -n ${K8S_NAMESPACE} -l app.kubernetes.io/name=loki -ojsonpath='{.items[0].metadata.name}') 3100
   ```

1. To get the latest logs from Loki, run a [range query](https://grafana.com/docs/loki/latest/reference/loki-http-api/#query-logs-within-a-range-of-time) returning the last 100 items:

   ```bash
   curl -G -s  "http://localhost:3100/loki/api/v1/query_range" \
     --data-urlencode \
     'query={service_name!=""}' \
   ```

## Grafana Installation

Because Grafana provides a very good Loki integration, you might want to install it as well.

### Install Grafana

1. To deploy Grafana, run:

   ```bash
   helm upgrade --install --create-namespace -n ${K8S_NAMESPACE} grafana grafana/grafana -f https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/loki/grafana-values.yaml
   ```

1. To enable Loki as Grafana data source, run:

   ```bash
   cat <<EOF | kubectl apply -n ${K8S_NAMESPACE} -f -
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: grafana-loki-datasource
     labels:
       grafana_datasource: "1"
   data:
      loki-datasource.yaml: |-
         apiVersion: 1
         datasources:
         - name: Loki
           type: loki
           access: proxy
           url: "http://${HELM_LOKI_RELEASE}:3100"
           version: 1
           isDefault: false
           jsonData: {}
   EOF
   ```

1. Get the password to access the Grafana UI:

   ```bash
   kubectl get secret --namespace ${K8S_NAMESPACE} grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
   ```

1. To access the Grafana UI with kubectl port forwarding, run:

   ```bash
   kubectl -n ${K8S_NAMESPACE} port-forward svc/grafana 3000:80
   ```

1. In your browser, open Grafana under `http://localhost:3000` and log in with user admin and the password you retrieved before.
  
## Grafana Exposure

### Expose Grafana

1. To expose Grafana using the Kyma API Gateway, create an APIRule:

   ```bash
   kubectl -n ${K8S_NAMESPACE} apply -f https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/loki/apirule.yaml
   ```

1. Get the public URL of your Loki instance:

   ```bash
   kubectl -n ${K8S_NAMESPACE} get virtualservice -l apirule.gateway.kyma-project.io/v1beta1=grafana.${K8S_NAMESPACE} -ojsonpath='{.items[*].spec.hosts[*]}'
   ```

### Add a Link for Grafana to the Kyma Dashboard

1. Download the `kyma-dashboard-configmap.yaml` file and change `{GRAFANA_LINK}` to the public URL of your Grafana instance.

   ```bash
   curl https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/loki/kyma-dashboard-configmap.yaml -o kyma-dashboard-configmap.yaml
   ```

1. Optionally, adjust the ConfigMap: You can change the label field to change the name of the tab. If you want to move it to another category, change the category tab.

1. Apply the ConfigMap, and go to Kyma dashboard. Under the Observability section, you should see a link to the newly exposed Grafana. If you already have a busola-config, merge it with the existing one:

   ```bash
   kubectl apply -f dashboard-configmap.yaml 
   ```

## Clean Up

1. To remove the installation from the cluster, run:

   ```bash
   helm delete -n ${K8S_NAMESPACE} ${HELM_LOKI_RELEASE}
   helm delete -n ${K8S_NAMESPACE} promtail
   helm delete -n ${K8S_NAMESPACE} grafana
   ```

2. To remove the deployed LogPipeline instance from cluster, run:

   ```bash
   kubectl delete LogPipeline custom-loki
   ```
