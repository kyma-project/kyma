
# GCP SQL Proxy

[sql-proxy](https://cloud.google.com/sql/docs/postgres/sql-proxy) The Cloud SQL Proxy provides secure access to your Cloud SQL Postgres/MySQL instances without having to whitelist IP addresses or configure SSL.

Accessing your Cloud SQL instance using the Cloud SQL Proxy offers these advantages:

* Secure connections: The proxy automatically encrypts traffic to and from the database; SSL certificates are used to verify client and server identities.
* Easier connection management: The proxy handles authentication with Google Cloud SQL, removing the need to provide static IP addresses of your GKE/GCE Kubernetes nodes.

## Introduction

This chart creates a Google Cloud SQL proxy deployment and service on a Kubernetes cluster using the Helm package manager.
You need to enable Cloud SQL Administration API and create a service account for the proxy as per these [instructions](https://cloud.google.com/sql/docs/postgres/connect-container-engine).

## Prerequisites

- Kubernetes cluster on Google Container Engine (GKE)
- Kubernetes cluster on Google Compute Engine (GCE)
- Cloud SQL Administration API enabled
- GCP Service account for the proxy with `Cloud SQL Admin` role, and `Cloud SQL Admin API` enabled.

## Installing the Chart

1. Install the Chart from a remote URL with the release name `pg-sqlproxy` into the Namespace `sqlproxy`. Set the GCP service account and SQL instances and ports:

```console
$ helm upgrade pg-sqlproxy rimusz/gcloud-sqlproxy --namespace sqlproxy \
    --set serviceAccountKey="$(cat service-account.json | base64 | tr -d '\n')" \
    --set cloudsql.instances[0].instance=INSTANCE \
    --set cloudsql.instances[0].project=PROJECT \
    --set cloudsql.instances[0].region=REGION \
    --set cloudsql.instances[0].port=5432 -i
```

2. Replace Postgres/MySQL host:
- If access is from the same Namespace, replace with `pg-sqlproxy-gcloud-sqlproxy`.
- If it is from a different Namespace, replace with `pg-sqlproxy-gcloud-sqlproxy.sqlproxy`.

> **Tip**: List all releases using `helm list`

> **Tip**: If you encounter a YAML parse error on `gcloud-sqlproxy/templates/secrets.yaml`, you might need to set `-w 0` option to `base64` command.

> **Tip**: If you are using a MySQL instance, you may want to replace `pg-sqlproxy` with `mysql-sqlproxy` and `5432` with `3306`.

> **Tip**: Because of limitations on the length of port names, the `instance` value for each of the instances must be unique for the first 15 characters.

## Uninstalling the Chart

To uninstall/delete the `my-release-name` deployment:

```console
$ helm delete my-release-name
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the `gcloud-sqlproxy` chart and their default values.

| Parameter                         | Description                             | Default                                                                                     |
| --------------------------------- | --------------------------------------  | ---------------------------------------------------------                                   |
| `image`                           | SQLProxy image                          | `gcr.io/cloudsql-docker/gce-proxy`                                                        |
| `imageTag`                        | SQLProxy image tag                      | `1.16`                                                                                      |
| `imagePullPolicy`                 | Image pull policy                       | `IfNotPresent`                                                                              |
| `replicasCount`                   | Replicas count                          | `1`                                                                                         |
| `serviceAccountKey`               | Service account key JSON file. Provide it encoded with base64 if no existing Secret is used to create a new Secret. | |
| `existingSecret`                  | Name of an existing secret to be used for the cloud-sql credentials | `""`                                                            |
| `existingSecretKey`               | The key to use in the provided existing secret   | `""`                                                                               |
| `usingGCPController`              | enable the use of the GCP Service Account Controller     | `""`                                                                       |
| `serviceAccountName`              | specify a service account name to use with GCP Controller | `""`                                                                                        |
| `cloudsql.instances`              | List of PostgreSQL/MySQL instances      | [{instance: `instance`, project: `project`, region: `region`, port: 5432}] must be provided |
| `resources`                       | CPU/Memory resource requests/limits     | Memory: `100/150Mi`, CPU: `100/150m`                                                        |
| `lifecycleHooks`                  | Container lifecycle hooks               | `{}`                                                                                        |
| `autoscaling.enabled`             | Enable CPU/Memory horizontal pod autoscaler | `false`                                                                                 |
| `autoscaling.minReplicas`         | Autoscaler minimum Pod replica count    | `1`                                                                                         |
| `autoscaling.maxReplicas`         | Autoscaler maximum Pod replica count    | `3`                                                                                         |
| `autoscaling.targetCPUUtilizationPercentage` | Scaling target for CPU Utilization Percentage | `50`                                                                       |
| `autoscaling.targetMemoryUtilizationPercentage` | Scaling target for Memory Utilization Percentage | `50`                                                                 |
| `terminationGracePeriodSeconds`   | Number of seconds to wait before Pod is killed  | `30` (Kubernetes default)                                                                   |
| `podAnnotations`                  | Pod Annotations                         |                                                                                             |
| `priorityClassName`                  | Priority Class Name                  | `""`                                                                                         |
| `nodeSelector`                    | Node Selector                           |                                                                                             |
| `podDisruptionBudget`             | Pod disruption budget                   | `maxUnavailable: 1` if `replicasCount` > 1, does not create the PDB otherwise               |
| `service.type`                    | Kubernetes LoadBalancer type            | `ClusterIP`                                                                                 |
| `service.internalLB`              | Creates service with `cloud.google.com/load-balancer-type: "Internal"` | Default `false`. If `true`, you must also set the `service.type=LoadBalancer`. |
| `rbac.create`                     | Create RBAC configuration w/ SA         | `false`                                                                                     |
| `serviceAccount.create` | Create a service account | `true` |
| `serviceAccount.annotations` | Annotations for the service account | `{}` |
| `serviceAccount.name` |  Service account name | Generated using the fullname template |
| `networkPolicy.enabled`           | Enable NetworkPolicy                    | `false`                                                                                     |
| `networkPolicy.ingress.from`      | List of sources to have access to the Pods selected for this rule. If empty, allows all sources. | `[]`                  |
| `extraArgs`                       | Additional container arguments          | `{}`                                                                                        |
| `livenessProbe.enabled`           | Would you like a livenessProbe to be enabled  | `false`                                                                               |
| `livenessProbe.port`              | The port which will be checked by the probe   | 5432                                                                                  |
| `livenessProbe.initialDelaySeconds` | Delay before liveness probe is initiated    | 30                                                                                    |
| `livenessProbe.periodSeconds`     | How often to perform the probe. Provide the interval in seconds.  | 10                                                                                    |
| `livenessProbe.timeoutSeconds`    | Number of seconds after which the probe times out.  | 5                                                                                     |
| `livenessProbe.failureThreshold`  | Minimum consecutive failures for the probe to be considered failed after having succeeded.  | 6                                       |
| `livenessProbe.successThreshold`  | Minimum consecutive successes for the probe to be considered successful after having failed | 1                                       |
| `readinessProbe.enabled`          | would you like a readinessProbe to be enabled | `false`                                                                               |
| `readinessProbe.port`              | The port which will be checked by the probe  | 5432                                                                                  |
| `readinessProbe.initialDelaySeconds` | Delay before readiness probe is initiated  | 5                                                                                     |
| `readinessProbe.periodSeconds`    | How often to perform the probe. Provide the interval in seconds.  | 10                                                                                    |
| `readinessProbe.timeoutSeconds`   | Number of seconds after which the probe times out                      | 5                                                                                     |
| `readinessProbe.failureThreshold` | Minimum consecutive failures for the probe to be considered failed after having succeeded.  | 6                                       |
| `readinessProbe.successThreshold` | Minimum consecutive successes for the probe to be considered successful after having failed | 1                                       |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Provide the `extraArgs` using dot notation. For example, `--set extraArgs.log_debug_stdout=true` passes `--log_debug_stdout=false` to the SQL Proxy command.

Alternatively, provide a YAML file that specifies the values for the above parameters while installing the chart. For example,

```console
$ helm install --name my-release -f values.yaml rimusz/gcloud-sqlproxy
```

> **TIP**: You can use the default [values.yaml](values.yaml).

### Autogenerating the GCP service account
By enabling the flag `usingGCPController` and having a GCP Service Account Controller deployed in your cluster, it is possible to autogenerate and inject the service account used for connecting to the database. For more information see [this](https://github.com/kiwigrid/helm-charts/tree/master/charts/gcp-serviceaccount-controller) document.

## Documentation

- [Cloud SQL Proxy for Postgres](https://cloud.google.com/sql/docs/postgres/sql-proxy)
- [Cloud SQL Proxy for MySQL](https://cloud.google.com/sql/docs/mysql/sql-proxy)
- [GKE samples](https://github.com/GoogleCloudPlatform/container-engine-samples/tree/master/cloudsql)


## Upgrading

**From < 0.10.0 to >= 0.10.0**

If the chart name is included in the release name, the chart name is used as a full name.
For example, a service `gcloud-sqlproxy-gcloud-sqlproxy` shows up as `gcloud-sqlproxy`.

**From < 0.11.0 to >= 0.11.0**

As of `0.11.0`, recommended labels are used. Take into account anything that may target your release's objects via labels.

## Support

Kubernetes versions older than 1.9 are not supported by this chart.
