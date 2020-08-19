# Service Catalog

Service Catalog is a Kubernetes Incubator project that provides a
Kubernetes-native workflow for integrating with
[Open Service Brokers](https://www.openservicebrokerapi.org/)
to provision and bind to application dependencies like databases, object
storage, message-oriented middleware, and more.

For more information,
[visit the project on github](https://github.com/kubernetes-sigs/service-catalog).

## Prerequisites

- Kubernetes 1.13+
- `charts/catalog` already exists in your local machine

## Installing the Chart

To install the chart with the release name `catalog`:

```bash
$ helm install . --name catalog --namespace catalog
```

## Uninstalling the Chart

To uninstall/delete the `catalog` deployment:

```bash
$ helm delete --purge catalog
```

The command removes all the Kubernetes components associated with the chart and
deletes the release.

## Configuration

The following tables lists the configurable parameters of the Service Catalog
chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image` | Service catalog image to use | `eu.gcr.io/kyma-project/external/quay.io/kubernetes-service-catalog/service-catalog:v0.3.0` |
| `imagePullPolicy` | `imagePullPolicy` for the service catalog | `Always` |
| `imagePullSecrets`|  The pre-existing secrets to use to pull images from a private registry | `[]` | 
| `webhook.updateStrategy` | `updateStrategy` for the service catalog webhook deployment | `RollingUpdate` |
| `webhook.minReadySeconds` | how many seconds an webhook server pod needs to be ready before killing the next, during update | `1` |
| `webhook.annotations` | Annotations for webhook pods | `{}` |
| `webhook.nodeSelector` | A nodeSelector value to apply to the webhook pods. If not specified, no nodeSelector will be applied | |
| `webhook.service.type` | Type of service; valid values are `LoadBalancer` , `NodePort` and `ClusterIP` | `NodePort` |
| `webhook.service.nodePort.securePort` | If service type is `NodePort`, specifies a port in allowable range (e.g. 30000 - 32767 on minikube); The TLS-enabled endpoint will be exposed here | `30443` |
| `webhook.service.clusterIP` | If service type is ClusterIP, specify clusterIP as `None` for `headless services` OR specify your own specific IP OR leave blank to let Kubernetes assign a cluster IP |  |
| `webhook.verbosity` | Log level; valid values are in the range 0 - 10 | `10` |
| `webhook.healthcheck.enabled` | Enable readiness and liveliness probes | `true` |
| `webhook.resources` | Resources allocation (Requests and Limits) | `{requests: {cpu: 100m, memory: 20Mi}, limits: {cpu: 100m, memory: 30Mi}}` |
| `controllerManager.replicas` | `replicas` for the service catalog controllerManager pod count | `1` |
| `controllerManager.updateStrategy` | `updateStrategy` for the service catalog controllerManager deployments | `RollingUpdate` |
| `controllerManager.minReadySeconds` | how many seconds a controllerManager pod needs to be ready before killing the next, during update | `1` |
| `controllerManager.annotations` | Annotations for controllerManager pods | `{}` |
| `controllerManager.nodeSelector` | A nodeSelector value to apply to the controllerManager pods. If not specified, no nodeSelector will be applied | |
| `controllerManager.healthcheck.enabled` | Enable readiness and liveliness probes | `true` |
| `controllerManager.verbosity` | Log level; valid values are in the range 0 - 10 | `10` |
| `controllerManager.resyncInterval` | How often the controller should resync informers; duration format (`20m`, `1h`, etc) | `5m` |
| `controllerManager.osbApiRequestTimeout` | The maximum amount of timeout to any request to the broker; duration format (`60s`, `3m`, etc) | `60s` |
| `controllerManager.brokerRelistInterval` | How often the controller should relist the catalogs of ready brokers; duration format (`20m`, `1h`, etc) | `24h` |
| `controllerManager.brokerRelistIntervalActivated` | Whether or not the controller supports a --broker-relist-interval flag. If this is set to true, brokerRelistInterval will be used as the value for that flag. | `true` |
| `controllerManager.profiling.disabled` | Disable profiling via web interface host:port/debug/pprof/ | `false` |
| `controllerManager.profiling.contentionProfiling` | Enables lock contention profiling, if profiling is enabled | `false` |
| `controllerManager.leaderElection.activated` | Whether the controller has leader election enabled | `false` |
| `controllerManager.serviceAccount` | Service account | `service-catalog-controller-manager` |
| `controllerManager.enablePrometheusScrape` | Whether the controller will expose metrics on /metrics | `false` |
| `controllerManager.resources` | Resources allocation (Requests and Limits) | `{requests: {cpu: 100m, memory: 20Mi}, limits: {cpu: 100m, memory: 30Mi}}` |
| `controllerManager.service.type` | Type of service; valid values are `LoadBalancer` , `NodePort` and `ClusterIP` | `ClusterIP` |
| `controllerManager.service.nodePort.securePort` | If service type is `NodePort`, specifies a port in allowable range (e.g. 30000 - 32767 on minikube); The TLS-enabled endpoint will be exposed here | `30444` |
| `controllerManager.service.clusterIP` | If service type is ClusterIP, specify clusterIP as `None` for `headless services` OR specify your own specific IP OR leave blank to let Kubernetes assign a cluster IP |  |
| `rbacEnable` | If true, create & use RBAC resources | `true` |
| `originatingIdentityEnabled` | Whether the OriginatingIdentity feature should be enabled | `true` |
| `asyncBindingOperationsEnabled` | Whether or not alpha support for async binding operations is enabled | `false` |
| `namespacedServiceBrokerDisabled` | Whether or not alpha support for namespace scoped brokers is disabled | `false` |

Specify each parameter using the `--set key=value[,key=value]` argument to
`helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be
provided while installing the chart. For example:

```bash
$ helm install . --name catalog --namespace catalog --values values.yaml
```
