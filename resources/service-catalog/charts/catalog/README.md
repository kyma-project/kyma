# Service Catalog

## Overview

Service Catalog is a Kubernetes Incubator project that provides a
Kubernetes-native workflow for integrating with
[Open Service Brokers](https://www.openservicebrokerapi.org/)
to provision and bind application dependencies, such as databases, object
storages, or message-oriented middlewares.

For more information,
go to [Kubernetes Incubator](https://github.com/kubernetes-incubator/service-catalog).

## Prerequisites

- Kubernetes 1.9+ with Beta APIs enabled
- `charts/catalog` copied to your local machine

## Details
### Install the chart

To install the chart with the `catalog` release name, run:

```bash
helm install . --name catalog --namespace catalog
```

### Uninstall the chart

To uninstall the `catalog` deployment, run:

```bash
helm delete --purge catalog
```

The command removes all Kubernetes components associated with the chart, and
deletes the release.

### Configuration

The following table lists the configurable parameters of the Service Catalog
chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| **image** | Specifies the Service Catalog image. | `quay.io/kubernetes-service-catalog/service-catalog:v0.1.41` |
| **imagePullPolicy** | Specifies **imagePullPolicy** for the Service Catalog. | `Always` |
| **webhook.updateStrategy** | Specifies **updateStrategy** for the Service Catalog webhook deployment. | `RollingUpdate` |
| **webhook.minReadySeconds** | The minimum number of seconds for which a newly created webhook Pod is ready without any of its containers crashing, for it to be considered available. | `1` |
| **webhook.annotations** | Provides annotations for webhook Pods. | `{}` |
| **webhook.nodeSelector** | A **nodeSelector** value to apply to the webhook Pods. If not specified, no nodeSelector will be applied. | |
| **webhook.service.type** | Specifies the type of service. The possible values are `LoadBalancer`, `NodePort`, and `ClusterIP`. | `NodePort` |
| **webhook.service.nodePort.securePort** | If the service type is `NodePort`, it specifies a port in the allowed range in which the TLS-enabled endpoint will be exposed (for example, `30000 - 32767` on Minikube). | `30443` |
| **webhook.service.clusterIP** | If the service type is `ClusterIP`, specify the cluster as `None` for the headless services, specify your own specific IP, or leave this parameter blank to let Kubernetes assign a cluster IP. |  |
| **webhook.verbosity** | The log level. The possible values range from `0 - 10`. | `10` |
| **webhook.healthcheck.enabled** | Enables readiness and liveliness probes. | `true` |
| **webhook.resources** | Specifies resources allocation (requests and limits). | `{requests: {cpu: 100m, memory: 20Mi}, limits: {cpu: 100m, memory: 30Mi}}` |
| **controllerManager.replicas** | Number of desired Pods for Service Catalog controllerManager. | `1` |
| **controllerManager.updateStrategy** | Specifies the update strategy for the Service Catalog controllerManager deployments. | `RollingUpdate` |
| **controllerManager.minReadySeconds** |The minimum number of seconds for which a newly created controllerManager Pod is ready without any of its containers crashing, for it to be considered available.  | `1` |
| **controllerManager.annotations** | Provides annotations for controllerManager Pods. | `{}` |
| **controllerManager.nodeSelector** | A nodeSelector value to apply to the controllerManager Pods. If not specified, no nodeSelector will be applied. | |
| **controllerManager.healthcheck.enabled** | Enables readiness and liveliness probes. | `true` |
| **controllerManager.verbosity** | The log level. The possible values range from `0 - 10`. | `10` |
| **controllerManager.resyncInterval** | Specifies how often the controller resyncs informers. The duration has a format of `20m`, `1h`, etc. | `5m` |
| **controllerManager.brokerRelistInterval** | Specifies how often the controller relists the catalogs of ready brokers. The duration has a format of `20m`, `1h`, etc. | `24h` |
| **controllerManager.brokerRelistIntervalActivated** | Specifies if the controller supports a `--broker-relist-interval` flag. If set to `true`, the `brokerRelistInterval` value will be used for that flag. | `true` |
| **controllerManager.profiling.disabled** | Disable profiling using the `{host}:{port}/debug/pprof/` web interface. | `false` |
| **controllerManager.profiling.contentionProfiling** | Enables lock contention profiling, if profiling is enabled. | `false` |
| **controllerManager.leaderElection.activated** | Specifies if the controller has leader election enabled. | `false` |
| **controllerManager.serviceAccount** | Specifies the service account. | `service-catalog-controller-manager` |
| **controllerManager.enablePrometheusScrape** | Specifies if the controller will expose metrics on the `/metrics` endpoint. | `false` |
| **controllerManager.resources** | Specifies resources allocation (requests and limits). | `{requests: {cpu: 100m, memory: 20Mi}, limits: {cpu: 100m, memory: 30Mi}}` |
| **rbacEnable** | If set to `true`, you can create and use RBAC resources. | `true` |
| **originatingIdentityEnabled** | Specifies if the OriginatingIdentity feature is enabled. | `true` |
| **asyncBindingOperationsEnabled** | Specifies if the alpha support for async binding operations is enabled. | `false` |
| **namespacedServiceBrokerDisabled** | Specifies is the alpha support for Namespace-scoped brokers is disabled. | `false` |

Specify each parameter using the `--set key=value[,key=value]` argument to
`helm install`.

Alternatively, you can create a YAML file that specifies the values for the parameters 
while installing the chart. For example:

```bash
helm install . --name catalog --namespace catalog --values values.yaml
```
