# PostgreSQL

## Overview

[PostgreSQL](https://www.postgresql.org/) is an object-relational database management system (ORDBMS) with an emphasis on extensibility and on standards-compliance.

## Prerequisites

- Kubernetes 1.10+
- PV provisioner support in the underlying infrastructure

## Details

### Configuration

The following tables lists the configurable parameters of the PostgreSQL chart and their default values.

| Parameter                                     | Description                                                                                                            | Default                                                     |
| --------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| `global.imageRegistry`                        | Global Docker Image registry                                                                                           | `nil`                                                       |
| `global.postgresql.postgresqlDatabase`        | PostgreSQL database (overrides `postgresqlDatabase`)                                                                   | `nil`                                                       |
| `global.postgresql.postgresqlUsername`        | PostgreSQL username (overrides `postgresqlUsername`)                                                                   | `nil`                                                       |
| `global.postgresql.existingSecret`            | Name of existing secret to use for PostgreSQL passwords (overrides `existingSecret`)                                   | `nil`                                                       |
| `global.postgresql.postgresqlPassword`        | Name of existing secret to use for PostgreSQL passwords (overrides `postgresqlPassword`)                               | `nil`                                                       |
| `global.postgresql.servicePort`               | PostgreSQL port (overrides `service.port`)                                                                             | `nil`                                                       |
| `global.postgresql.replicationPassword`       | Replication user password (overrides `replication.password`)                                                           | `nil`                                                       |
| `global.imagePullSecrets`                     | Global Docker registry secret names as an array                                                                        | `[]` (does not add image pull secrets to deployed pods)     |
| `image.registry`                              | PostgreSQL Image registry                                                                                              | `docker.io`                                                 |
| `image.repository`                            | PostgreSQL Image name                                                                                                  | `bitnami/postgresql`                                        |
| `image.tag`                                   | PostgreSQL Image tag                                                                                                   | `{TAG_NAME}`                                                |
| `image.pullPolicy`                            | PostgreSQL Image pull policy                                                                                           | `IfNotPresent`                                              |
| `image.pullSecrets`                           | Specify Image pull secrets                                                                                             | `nil` (does not add image pull secrets to deployed pods)    |
| `image.debug`                                 | Specify if debug values should be set                                                                                  | `false`                                                     |
| `volumePermissions.image.registry`            | Init container volume-permissions image registry                                                                       | `docker.io`                                                 |
| `volumePermissions.image.repository`          | Init container volume-permissions image name                                                                           | `bitnami/minideb`                                           |
| `volumePermissions.image.tag`                 | Init container volume-permissions image tag                                                                            | `latest`                                                    |
| `volumePermissions.image.pullPolicy`          | Init container volume-permissions image pull policy                                                                    | `Always`                                                    |
| `volumePermissions.securityContext.runAsUser` | User ID for the init container                                                                                         | `0`                                                         |
| `usePasswordFile`                             | Have the secrets mounted as a file instead of env vars                                                                 | `false`                                                     |
| `replication.enabled`                         | Enable replication                                                                                                     | `false`                                                     |
| `replication.user`                            | Replication user                                                                                                       | `repl_user`                                                 |
| `replication.password`                        | Replication user password                                                                                              | `repl_password`                                             |
| `replication.slaveReplicas`                   | Number of slaves replicas                                                                                              | `1`                                                         |
| `replication.synchronousCommit`               | Set synchronous commit mode. Allowed values: `on`, `remote_apply`, `remote_write`, `local` and `off`                   | `off`                                                       |
| `replication.numSynchronousReplicas`          | Number of replicas that will have synchronous replication. Note: Cannot be greater than `replication.slaveReplicas`.   | `0`                                                         |
| `replication.applicationName`                 | Cluster application name. Useful for advanced replication settings                                                     | `my_application`                                            |
| `existingSecret`                              | Name of existing secret to use for PostgreSQL passwords                                                                | `nil`                                                       |
| `postgresqlUsername`                          | PostgreSQL admin user                                                                                                  | `postgres`                                                  |
| `postgresqlPassword`                          | PostgreSQL admin password                                                                                              | _random 10 character alphanumeric string_                   |
| `postgresqlDatabase`                          | PostgreSQL database                                                                                                    | `nil`                                                       |
| `postgresqlDataDir`                           | PostgreSQL data dir folder                                                                                             | `/bitnami/postgresql` (same value as persistence.mountPath) |
| `postgresqlInitdbArgs`                        | PostgreSQL initdb extra arguments                                                                                      | `nil`                                                       |
| `postgresqlInitdbWalDir`                      | PostgreSQL location for transaction log                                                                                | `nil`                                                       |
| `postgresqlConfiguration`                     | Runtime Config Parameters                                                                                              | `nil`                                                       |
| `postgresqlExtendedConf`                      | Extended Runtime Config Parameters (appended to main or default configuration)                                         | `nil`                                                       |
| `pgHbaConfiguration`                          | Content of pg\_hba.conf                                                                                                | `nil (do not create pg_hba.conf)`                           |
| `configurationConfigMap`                      | ConfigMap with the PostgreSQL configuration files (Note: Overrides `postgresqlConfiguration` and `pgHbaConfiguration`). The value is evaluated as a template. | `nil`                                                       |
| `extendedConfConfigMap`                       | ConfigMap with the extended PostgreSQL configuration files. The value is evaluated as a template.                      | `nil`                                                       |
| `initdbScripts`                               | Dictionary of initdb scripts                                                                                           | `nil`                                                       |
| `initdbScriptsConfigMap`                      | ConfigMap with the initdb scripts (Note: Overrides `initdbScripts`). The value is evaluated as a template.             | `nil`                                                       |
| `initdbScriptsSecret`                         | Secret with initdb scripts that contain sensitive information (Note: can be used with `initdbScriptsConfigMap` or `initdbScripts`). The value is evaluated as a template. | `nil`                                           |
| `service.type`                                | Kubernetes Service type                                                                                                | `ClusterIP`                                                 |
| `service.port`                                | PostgreSQL port                                                                                                        | `5432`                                                      |
| `service.nodePort`                            | Kubernetes Service nodePort                                                                                            | `nil`                                                       |
| `service.annotations`                         | Annotations for PostgreSQL service                                                                                     | {}                                                          |
| `service.loadBalancerIP`                      | loadBalancerIP if service type is `LoadBalancer`                                                                       | `nil`                                                       |
| `service.loadBalancerSourceRanges`            | Address that are allowed when svc is LoadBalancer                                                                      | []                                                          |
| `schedulerName`                               | Name of the k8s scheduler (other than default)                                                                         | `nil`                                                       |
| `persistence.enabled`                         | Enable persistence using PVC                                                                                           | `true`                                                      |
| `persistence.existingClaim`                   | Provide an existing `PersistentVolumeClaim`, the value is evaluated as a template.                                     | `nil`                                                       |
| `persistence.mountPath`                       | Path to mount the volume at                                                                                            | `/bitnami/postgresql`                                       |
| `persistence.subPath`                         | Subdirectory of the volume to mount at                                                                                 | `""`                                                        |
| `persistence.storageClass`                    | PVC Storage Class for PostgreSQL volume                                                                                | `nil`                                                       |
| `persistence.accessModes`                     | PVC Access Mode for PostgreSQL volume                                                                                  | `[ReadWriteOnce]`                                           |
| `persistence.size`                            | PVC Storage Request for PostgreSQL volume                                                                              | `8Gi`                                                       |
| `persistence.annotations`                     | Annotations for the PVC                                                                                                | `{}`                                                        |
| `master.nodeSelector`                         | Node labels for pod assignment (postgresql master)                                                                     | `{}`                                                        |
| `master.affinity`                             | Affinity labels for pod assignment (postgresql master)                                                                 | `{}`                                                        |
| `master.tolerations`                          | Toleration labels for pod assignment (postgresql master)                                                               | `[]`                                                        |
| `master.podAnnotations`                       | Map of annotations to add to the pods (postgresql master)                                                              | `{}`                                                        |
| `master.podLabels`                            | Map of labels to add to the pods (postgresql master)                                                                   | `{}`                                                        |
| `master.extraVolumeMounts`                    | Additional volume mounts to add to the pods (postgresql master)                                                        |  `[]`                                                       |
| `master.extraVolume`                          | Additional volumes to add to the pods (postgresql master)                                                              |  `[]`                                                       |
| `slave.nodeSelector`                          | Node labels for pod assignment (postgresql slave)                                                                      | `{}`                                                        |
| `slave.affinity`                              | Affinity labels for pod assignment (postgresql slave)                                                                  | `{}`                                                        |
| `slave.tolerations`                           | Toleration labels for pod assignment (postgresql slave)                                                                | `[]`                                                        |
| `slave.podAnnotations`                        | Map of annotations to add to the pods (postgresql slave)                                                               | `{}`                                                        |
| `slave.podLabels`                             | Map of labels to add to the pods (postgresql slave)                                                                    | `{}`                                                        |
| `slave.extraVolumeMounts`                     | Additional volume mounts to add to the pods (postgresql slave)                                                         |  `[]`                                                       |
| `slave.extraVolume`                           | Additional volumes to add to the pods (postgresql slave)                                                               |  `[]`                                                       |
| `terminationGracePeriodSeconds`               | Seconds the pod needs to terminate gracefully                                                                          | `nil`                                                       |
| `resources`                                   | CPU/Memory resource requests/limits                                                                                    | Memory: `256Mi`, CPU: `250m`                                |
| `securityContext.enabled`                     | Enable security context                                                                                                | `true`                                                      |
| `securityContext.fsGroup`                     | Group ID for the container                                                                                             | `1001`                                                      |
| `securityContext.runAsUser`                   | User ID for the container                                                                                              | `1001`                                                      |
| `serviceAccount.enabled`                      | Enable service account (Note: Service Account will only be automatically created if `serviceAccount.name` is not set)  | `false`                                                     |
| `serviceAcccount.name`                        | Name of existing service account                                                                                       | `nil`                                                       |
| `livenessProbe.enabled`                       | Would you like a livenessProbe to be enabled                                                                            | `true`                                                      |
| `networkPolicy.enabled`                       | Enable NetworkPolicy                                                                                                   | `false`                                                     |
| `networkPolicy.allowExternal`                 | Don't require client label for connections                                                                             | `true`                                                      |
| `livenessProbe.initialDelaySeconds`           | Delay before liveness probe is initiated                                                                               | 30                                                          |
| `livenessProbe.periodSeconds`                 | How often to perform the probe                                                                                         | 10                                                          |
| `livenessProbe.timeoutSeconds`                | When the probe times out                                                                                               | 5                                                           |
| `livenessProbe.failureThreshold`              | Minimum consecutive failures for the probe to be considered failed after having succeeded.                             | 6                                                           |
| `livenessProbe.successThreshold`              | Minimum consecutive successes for the probe to be considered successful after having failed                            | 1                                                           |
| `readinessProbe.enabled`                      | would you like a readinessProbe to be enabled                                                                          | `true`                                                      |
| `readinessProbe.initialDelaySeconds`          | Delay before readiness probe is initiated                                                                               | 5                                                           |
| `readinessProbe.periodSeconds`                | How often to perform the probe                                                                                         | 10                                                          |
| `readinessProbe.timeoutSeconds`               | When the probe times out                                                                                               | 5                                                           |
| `readinessProbe.failureThreshold`             | Minimum consecutive failures for the probe to be considered failed after having succeeded.                             | 6                                                           |
| `readinessProbe.successThreshold`             | Minimum consecutive successes for the probe to be considered successful after having failed                            | 1                                                           |
| `metrics.enabled`                             | Start a prometheus exporter                                                                                            | `false`                                                     |
| `metrics.service.type`                        | Kubernetes Service type                                                                                                | `ClusterIP`                                                 |
| `service.clusterIP`                           | Static clusterIP or None for headless services                                                                         | `nil`                                                       |
| `metrics.service.annotations`                 | Additional annotations for metrics exporter pod                                                                        | `{ prometheus.io/scrape: "true", prometheus.io/port: "9187"}` |
| `metrics.service.loadBalancerIP`              | loadBalancerIP if redis metrics service type is `LoadBalancer`                                                         | `nil`                                                       |
| `metrics.serviceMonitor.enabled`              | Set this to `true` to create ServiceMonitor for Prometheus operator                                                    | `false`                                                     |
| `metrics.serviceMonitor.additionalLabels`     | Additional labels that can be used so ServiceMonitor will be discovered by Prometheus                                  | `{}`                                                        |
| `metrics.serviceMonitor.namespace`            | Optional namespace in which to create ServiceMonitor                                                                   | `nil`                                                       |
| `metrics.serviceMonitor.interval`             | Scrape interval. If not set, the Prometheus default scrape interval is used                                            | `nil`                                                       |
| `metrics.serviceMonitor.scrapeTimeout`        | Scrape timeout. If not set, the Prometheus default scrape timeout is used                                              | `nil`                                                       |
| `metrics.image.registry`                      | PostgreSQL Image registry                                                                                              | `docker.io`                                                 |
| `metrics.image.repository`                    | PostgreSQL Image name                                                                                                  | `wrouesnel/postgres_exporter`                               |
| `metrics.image.tag`                           | PostgreSQL Image tag                                                                                                   | `v0.4.7`                                                    |
| `metrics.image.pullPolicy`                    | PostgreSQL Image pull policy                                                                                           | `IfNotPresent`                                              |
| `metrics.image.pullSecrets`                   | Specify Image pull secrets                                                                                             | `nil` (does not add image pull secrets to deployed pods)    |
| `metrics.securityContext.enabled`             | Enable security context for metrics                                                                                    | `false`                                                     |
| `metrics.securityContext.runAsUser`           | User ID for the container for metrics                                                                                  | `1001`                                                      |
| `metrics.livenessProbe.initialDelaySeconds`   | Delay before liveness probe is initiated                                                                               | 30                                                          |
| `metrics.livenessProbe.periodSeconds`         | How often to perform the probe                                                                                         | 10                                                          |
| `metrics.livenessProbe.timeoutSeconds`        | When the probe times out                                                                                               | 5                                                           |
| `metrics.livenessProbe.failureThreshold`      | Minimum consecutive failures for the probe to be considered failed after having succeeded.                             | 6                                                           |
| `metrics.livenessProbe.successThreshold`      | Minimum consecutive successes for the probe to be considered successful after having failed                            | 1                                                           |
| `metrics.readinessProbe.enabled`              | would you like a readinessProbe to be enabled                                                                          | `true`                                                      |
| `metrics.readinessProbe.initialDelaySeconds`  | Delay before liveness probe is initiated                                                                               | 5                                                           |
| `metrics.readinessProbe.periodSeconds`        | How often to perform the probe                                                                                         | 10                                                          |
| `metrics.readinessProbe.timeoutSeconds`       | When the probe times out                                                                                               | 5                                                           |
| `metrics.readinessProbe.failureThreshold`     | Minimum consecutive failures for the probe to be considered failed after having succeeded.                             | 6                                                           |
| `metrics.readinessProbe.successThreshold`     | Minimum consecutive successes for the probe to be considered successful after having failed                            | 1                                                           |
| `extraEnv`                                    | Any extra environment variables you would like to pass on to the pod. The value is evaluated as a template.            | `{}`                                                        |
| `updateStrategy`                              | Update strategy policy                                                                                                 | `{type: "RollingUpdate"}`                                   |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```console
$ helm install --name my-release \
  --set postgresqlPassword=secretpassword,postgresqlDatabase=my-database \
    stable/postgresql
```

The above command sets the PostgreSQL `postgres` account password to `secretpassword`. Additionally it creates a database named `my-database`.

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```console
$ helm install --name my-release -f values.yaml stable/postgresql
```

> **Tip**: You can use the default [values.yaml](values.yaml)

### Production configuration

This chart includes a `values-production.yaml` file where you can find some parameters oriented to production configuration in comparison to the regular `values.yaml`.

```console
$ helm install --name my-release -f ./values-production.yaml stable/postgresql
```

- Enable replication:
```diff
- replication.enabled: false
+ replication.enabled: true
```

- Number of slaves replicas:
```diff
- replication.slaveReplicas: 1
+ replication.slaveReplicas: 2
```

- Set synchronous commit mode:
```diff
- replication.synchronousCommit: "off"
+ replication.synchronousCommit: "on"
```

- Number of replicas that will have synchronous replication:
```diff
- replication.numSynchronousReplicas: 0
+ replication.numSynchronousReplicas: 1
```

- Start a prometheus exporter:
```diff
- metrics.enabled: false
+ metrics.enabled: true
```

To horizontally scale this chart, first download the [values-production.yaml](values-production.yaml) file to your local folder, then:

```console
$ helm install --name my-release -f ./values-production.yaml stable/postgresql
$ kubectl scale statefulset my-postgresql-slave --replicas=3
```
