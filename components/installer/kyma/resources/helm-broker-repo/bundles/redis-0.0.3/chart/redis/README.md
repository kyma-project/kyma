```
_____          _ _     
|  __ \        | (_)    
| |__) |___  __| |_ ___
|  _  // _ \/ _` | / __|
| | \ \  __/ (_| | \__ \
|_|  \_\___|\__,_|_|___/
```

## Overview

[Redis](http://redis.io/) is an advanced key-value cache and store. It is often referred to as a data structure server, since keys can contain strings, hashes, lists, sets, sorted sets, bitmaps, and hyperloglogs.

## Prerequisites

- Kubernetes 1.4+ with Beta APIs enabled
- PV provisioner support in the underlying infrastructure

## Details

To install Redis:

```bash
$ helm install stable/redis
```

This chart bootstraps a [Redis](https://github.com/bitnami/bitnami-docker-redis) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

### Install the chart

To install the chart with the release name `my-release`:

```bash
$ helm install --name my-release stable/redis
```

The command deploys Redis on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters you can  configure during installation.

> **NOTE:** List all releases using `helm list`.

### Uninstall the chart

To delete the `my-release` deployment:

```bash
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

### Configuration

The following table lists the configurable parameters of the Redis chart and their default values.

| Parameter                  | Description                           | Default                                                   |
| -------------------------- | ------------------------------------- | --------------------------------------------------------- |
| `image`                    | Redis image                           | `bitnami/redis:{VERSION}`                                 |
| `imagePullPolicy`          | Image pull policy                     | `IfNotPresent`                                            |
| `usePassword`              | Use password                          | `true`                                         |
| `redisPassword`            | Redis password                        | Randomly generated                                        |
| `args`                     | Redis command-line args               | []                                                        |
| `persistence.enabled`      | Use a PVC to persist data             | `true`                                                    |
| `persistence.existingClaim`| Use an existing PVC to persist data   | `nil`                                                     |
| `persistence.storageClass` | Storage class of backing PVC          | `generic`                                                 |
| `persistence.accessMode`   | Use volume as ReadOnly or ReadWrite   | `ReadWriteOnce`                                           |
| `persistence.size`         | Size of data volume                   | `8Gi`                                                     |
| `resources`                | CPU/Memory resource requests/limits   | Memory: `256Mi`                                           |
| `metrics.enabled`          | Start a side-car prometheus exporter  | `false`                                                   |
| `metrics.image`            | Exporter image                        | `oliver006/redis_exporter`                                |
| `metrics.imageTag`         | Exporter image                        | `v0.11`                                                   |
| `metrics.imagePullPolicy`  | Exporter image pull policy            | `IfNotPresent`                                            |
| `metrics.resources`        | Exporter resource requests/limit      | Memory: `256Mi`                                           |
| `nodeSelector`             | Node labels for pod assignment        | {}                                                        |
| `tolerations`              | Toleration labels for pod assignment  | []                                                        |
| `networkPolicy.enabled`    | Enable NetworkPolicy                  | `false`                                                   |
| `networkPolicy.allowExternal` | Do not require client label for connections | `true`                                            |

The above parameters map to the env variables defined in [bitnami/redis](http://github.com/bitnami/bitnami-docker-redis). For more information, refer to the [bitnami/redis](http://github.com/bitnami/bitnami-docker-redis) image documentation.

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example:

```bash
$ helm install --name my-release \
  --set redisPassword=secretpassword \
    stable/redis
```

The above command sets the Redis server password to `secretpassword`.

Alternatively, you can provide a YAML file that specifies the values for the parameters while installing the chart. For example:

```bash
$ helm install --name my-release -f values.yaml stable/redis
```

> **NOTE:** You can use the default [values.yaml](values.yaml).

### NetworkPolicy

To enable network policy for Redis, install a networking plugin that implements the Kubernetes [NetworkPolicy spec](https://kubernetes.io/docs/tasks/administer-cluster/declare-network-policy#before-you-begin), and set `networkPolicy.enabled` to `true`.

For Kubernetes v1.5 and v1.6, turn on NetworkPolicy by setting the DefaultDeny namespace annotation.

> **NOTE:** This enforces policy for all pods in the namespace.

```
    kubectl annotate namespace default "net.beta.kubernetes.io/network-policy={\"ingress\":{\"isolation\":\"DefaultDeny\"}}"
```

With NetworkPolicy enabled, only pods with the generated client label can connect to Redis. The output displays this label after a successful install.

### Persistence

The [Bitnami Redis](https://github.com/bitnami/bitnami-docker-redis) image stores the Redis data and configurations at the `/bitnami/redis` path of the container.

By default, the chart mounts a [PersistentVolume](http://kubernetes.io/docs/user-guide/persistent-volumes/) volume at this location. The system creates the volume using dynamic volume provisioning. If a PersistentVolumeClaim already exists, specify it during installation.

#### Existing PersistentVolumeClaim

1. Create the PersistentVolume.
1. Create the PersistentVolumeClaim.
1. Install the chart.
```bash
$ helm install --set persistence.existingClaim=PVC_NAME redis
```

### Metrics
Optionally, the chart can start a metrics exporter for [prometheus](https://prometheus.io). The system does not expose the metrics endpoint (port 9121), and it is expected that you collect the metrics from inside the Kubernetes cluster using something similar to the [example Prometheus scrape configuration](https://github.com/prometheus/prometheus/blob/master/documentation/examples/prometheus-kubernetes.yml).
