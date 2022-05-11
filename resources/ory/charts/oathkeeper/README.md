# oathkeeper

![Version: 0.23.1](https://img.shields.io/badge/Version-0.23.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.38.19-beta.1](https://img.shields.io/badge/AppVersion-v0.38.19--beta.1-informational?style=flat-square)

A Helm chart for deploying ORY Oathkeeper in Kubernetes

**Homepage:** <https://www.ory.sh/>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| ORY Team | office@ory.sh | https://www.ory.sh/ |

## Source Code

* <https://github.com/ory/oathkeeper>
* <https://github.com/ory/k8s>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../oathkeeper-maester | oathkeeper-maester(oathkeeper-maester) | 0.23.1 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Configure node affinity |
| configmap.hashSumEnabled | bool | `true` | switch to false to prevent checksum annotations being maintained and propogated to the pods |
| demo | bool | `false` | If enabled, a demo deployment with exemplary access rules and JSON Web Key Secrets will be generated. |
| deployment.annotations | object | `{}` |  |
| deployment.automountServiceAccountToken | bool | `false` |  |
| deployment.extraContainers | object | `{}` | If you want to add extra sidecar containers. |
| deployment.extraEnv | list | `[]` |  |
| deployment.extraVolumeMounts | list | `[]` | Extra volume mounts, allows mounting the extraVolumes to the container. |
| deployment.extraVolumes | list | `[]` | Extra volumes you can attach to the pod. |
| deployment.labels | object | `{}` |  |
| deployment.nodeSelector | object | `{}` | Node labels for pod assignment. |
| deployment.resources | object | `{}` |  |
| deployment.securityContext.allowPrivilegeEscalation | bool | `false` |  |
| deployment.securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| deployment.securityContext.privileged | bool | `false` |  |
| deployment.securityContext.readOnlyRootFilesystem | bool | `true` |  |
| deployment.securityContext.runAsNonRoot | bool | `true` |  |
| deployment.securityContext.runAsUser | int | `1000` |  |
| deployment.serviceAccount | object | `{"annotations":{},"create":true,"name":""}` | Specify the serviceAccountName value. In some situations it is needed to provides specific permissions to Hydra deployments Like for example installing Hydra on a cluster with a PosSecurityPolicy and Istio. Uncoment if it is needed to provide a ServiceAccount for the Hydra deployment.** |
| deployment.serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| deployment.serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| deployment.serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| deployment.tolerations | list | `[]` | Configure node tolerations. |
| deployment.tracing | object | `{"datadog":{"enabled":false}}` | Configuration for tracing providers. Only datadog is currently supported through this block. If you need to use a different tracing provider, please manually set the configuration values via "oathkeeper.config" or via "deployment.extraEnv". |
| fullnameOverride | string | `""` | Full chart name override |
| global | object | `{"ory":{"oathkeeper":{"maester":{"mode":"controller"}}}}` | Mode for oathkeeper controller -- Two possible modes are: controller or sidecar |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| image.repository | string | `"oryd/oathkeeper"` | ORY Oathkeeper image |
| image.tag | string | `"v0.38.19-beta.1"` | ORY Oathkeeper version |
| imagePullSecrets | list | `[]` | Image pull secrets |
| ingress | object | `{"api":{"annotations":{},"className":"","enabled":false,"hosts":[{"host":"api.oathkeeper.localhost","paths":[{"path":"/","pathType":"ImplementationSpecific"}]}]},"proxy":{"annotations":{},"className":"","defaultBackend":{},"enabled":false,"hosts":[{"host":"proxy.oathkeeper.localhost","paths":[{"path":"/","pathType":"ImplementationSpecific"}]}]}}` | Configure ingress |
| ingress.api.enabled | bool | `false` | En-/Disable the api ingress. |
| ingress.proxy | object | `{"annotations":{},"className":"","defaultBackend":{},"enabled":false,"hosts":[{"host":"proxy.oathkeeper.localhost","paths":[{"path":"/","pathType":"ImplementationSpecific"}]}]}` | Configure ingress for the proxy port. |
| ingress.proxy.enabled | bool | `false` | En-/Disable the proxy ingress. |
| maester | object | `{"enabled":true}` | Configures controller setup |
| nameOverride | string | `""` | Chart name override |
| oathkeeper | object | `{"accessRules":{},"config":{"access_rules":{"repositories":["file:///etc/rules/access-rules.json"]},"serve":{"api":{"port":4456},"prometheus":{"port":9000},"proxy":{"port":4455}}},"managedAccessRules":true,"mutatorIdTokenJWKs":{}}` | Configure ORY Oathkeeper itself |
| oathkeeper.accessRules | object | `{}` | If set, uses the given access rules. |
| oathkeeper.config | object | `{"access_rules":{"repositories":["file:///etc/rules/access-rules.json"]},"serve":{"api":{"port":4456},"prometheus":{"port":9000},"proxy":{"port":4455}}}` | The ORY Oathkeeper configuration. For a full list of available settings, check:   https://github.com/ory/oathkeeper/blob/master/docs/config.yaml |
| oathkeeper.managedAccessRules | bool | `true` | If you enable maester, the following value should be set to "false" to avoid overwriting the rules generated by the CDRs. Additionally, the value "accessRules" shouldn't be used as it will have no effect once "managedAccessRules" is disabled. |
| oathkeeper.mutatorIdTokenJWKs | object | `{}` | If set, uses the given JSON Web Key Set as the signing key for the ID Token Mutator. |
| pdb | object | `{"enabled":false,"spec":{"minAvailable":1}}` | PodDistributionBudget configuration |
| replicaCount | int | `1` | Number of ORY Oathkeeper members |
| secret.enabled | bool | `true` | switch to false to prevent creating the secret |
| secret.filename | string | `"mutator.id_token.jwks.json"` | default filename of JWKS (mounted as secret) |
| secret.hashSumEnabled | bool | `true` | switch to false to prevent checksum annotations being maintained and propogated to the pods |
| secret.mountPath | string | `"/etc/secrets"` | default mount path for the kubernetes secret |
| secret.nameOverride | string | `""` | Provide custom name of existing secret, or custom name of secret to be created |
| secret.secretAnnotations."helm.sh/hook" | string | `"pre-install, pre-upgrade"` |  |
| secret.secretAnnotations."helm.sh/hook-delete-policy" | string | `"before-hook-creation"` |  |
| secret.secretAnnotations."helm.sh/hook-weight" | string | `"0"` |  |
| secret.secretAnnotations."helm.sh/resource-policy" | string | `"keep"` |  |
| service | object | `{"api":{"annotations":{},"enabled":true,"labels":{},"name":"http","port":4456,"type":"ClusterIP"},"metrics":{"annotations":{},"enabled":true,"labels":{},"name":"http","port":80,"type":"ClusterIP"},"proxy":{"annotations":{},"enabled":true,"labels":{},"name":"http","port":4455,"type":"ClusterIP"}}` | Configures the Kubernetes service |
| service.api | object | `{"annotations":{},"enabled":true,"labels":{},"name":"http","port":4456,"type":"ClusterIP"}` | Configures the Kubernetes service for the api port. |
| service.api.annotations | object | `{}` | If you do want to specify annotations, uncomment the following lines, adjust them as necessary, and remove the curly braces after 'annotations:'. kubernetes.io/ingress.class: nginx kubernetes.io/tls-acme: "true" |
| service.api.enabled | bool | `true` | En-/disable the service |
| service.api.labels | object | `{}` | If you do want to specify additional labels, uncomment the following lines, adjust them as necessary, and remove the curly braces after 'labels:'. e.g.  app: oathkeeper |
| service.api.name | string | `"http"` | The service port name. Useful to set a custom service port name if it must follow a scheme (e.g. Istio) |
| service.api.port | int | `4456` | The service port |
| service.api.type | string | `"ClusterIP"` | The service type |
| service.metrics | object | `{"annotations":{},"enabled":true,"labels":{},"name":"http","port":80,"type":"ClusterIP"}` | Configures the Kubernetes service for the metrics port. |
| service.metrics.annotations | object | `{}` | If you do want to specify annotations, uncomment the following lines, adjust them as necessary, and remove the curly braces after 'annotations:'. kubernetes.io/ingress.class: nginx kubernetes.io/tls-acme: "true" |
| service.metrics.enabled | bool | `true` | En-/disable the service |
| service.metrics.labels | object | `{}` | If you do want to specify additional labels, uncomment the following lines, adjust them as necessary, and remove the curly braces after 'labels:'. e.g.  app: oathkeeper |
| service.metrics.name | string | `"http"` | The service port name. Useful to set a custom service port name if it must follow a scheme (e.g. Istio) |
| service.metrics.port | int | `80` | The service port |
| service.metrics.type | string | `"ClusterIP"` | The service type |
| service.proxy | object | `{"annotations":{},"enabled":true,"labels":{},"name":"http","port":4455,"type":"ClusterIP"}` | Configures the Kubernetes service for the proxy port. |
| service.proxy.annotations | object | `{}` | If you do want to specify annotations, uncomment the following lines, adjust them as necessary, and remove the curly braces after 'annotations:'. kubernetes.io/ingress.class: nginx kubernetes.io/tls-acme: "true" |
| service.proxy.enabled | bool | `true` | En-/disable the service |
| service.proxy.labels | object | `{}` | If you do want to specify additional labels, uncomment the following lines, adjust them as necessary, and remove the curly braces after 'labels:'. e.g.  app: oathkeeper |
| service.proxy.name | string | `"http"` | The service port name. Useful to set a custom service port name if it must follow a scheme (e.g. Istio) |
| service.proxy.port | int | `4455` | The service port |
| service.proxy.type | string | `"ClusterIP"` | The service type |
| serviceMonitor | object | `{"labels":{},"scheme":"http","scrapeInterval":"60s","scrapeTimeout":"30s","tlsConfig":{}}` | Parameters for the Prometheus ServiceMonitor objects. Reference: https://docs.openshift.com/container-platform/4.6/rest_api/monitoring_apis/servicemonitor-monitoring-coreos-com-v1.html |
| serviceMonitor.labels | object | `{}` | Provide additionnal labels to the ServiceMonitor ressource metadata |
| serviceMonitor.scheme | string | `"http"` | HTTP scheme to use for scraping. |
| serviceMonitor.scrapeInterval | string | `"60s"` | Interval at which metrics should be scraped |
| serviceMonitor.scrapeTimeout | string | `"30s"` | Timeout after which the scrape is ended |
| serviceMonitor.tlsConfig | object | `{}` | TLS configuration to use when scraping the endpoint |
| sidecar | object | `{"envs":{},"image":{"repository":"oryd/oathkeeper-maester","tag":"v0.1.2"}}` | Options for the sidecar |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.5.0](https://github.com/norwoodj/helm-docs/releases/v1.5.0)
