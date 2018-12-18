apiVersion: v1
kind: Secret
metadata:
  name: remote-env-certificate-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
type: Opaque
data:
  global.remoteEnvCa: "__REMOTE_ENV_CA__"
  global.remoteEnvCaKey: "__REMOTE_ENV_CA_KEY__"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-certificate-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
data:
  global.tlsCrt: "__TLS_CERT__"
  global.tlsKey: "__TLS_KEY__"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: installation-config-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
data:
  global.isLocalEnv: "false"
  global.domainName: "__DOMAIN__"
  global.etcdBackup.containerName: "__ETCD_BACKUP_ABS_CONTAINER_NAME__"
  global.etcdBackup.enabled: "__ENABLE_ETCD_BACKUP__"
  nginx-ingress.controller.service.loadBalancerIP: "__REMOTE_ENV_IP__"
  cluster-users.users.adminGroup: "__ADMIN_GROUP__"
  etcd-stateful.replicaCount: "3"
  acceptanceTest.remoteEnvironment.disabled: "true"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: monitoring-config-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: monitoring
data:
  global.alertTools.credentials.slack.apiurl: "__SLACK_API_URL_VALUE__"
  global.alertTools.credentials.slack.channel: "__SLACK_CHANNEL_VALUE__"
  global.alertTools.credentials.victorOps.routingkey: "__VICTOR_OPS_ROUTING_KEY_VALUE__"
  global.alertTools.credentials.victorOps.apikey: "__VICTOR_OPS_API_KEY_VALUE__"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: connector-service-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: application-connector
data:
  connector-service.tests.skipSslVerify: "__SKIP_SSL_VERIFY__"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: core-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: core
data:
  console.cluster.headerLogoUrl: "assets/logo.svg"
  console.cluster.headerTitle: ""
  console.cluster.faviconUrl: "favicon.ico"
  minio.accessKey: "admin"
  minio.secretKey: "topSecretKey"
  minio.resources.limits.memory: 128Mi
  minio.resources.limits.cpu: 250m
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: istio
data:
  global.proxy.includeIPRanges: "10.0.0.1/8"

  security.enabled: "true"

  gateways.istio-ingressgateway.loadBalancerIP: "__EXTERNAL_PUBLIC_IP__"
  gateways.istio-ingressgateway.type: "LoadBalancer"

  pilot.resources.limits.memory: 2Gi
  pilot.resources.requests.memory: 512Mi

  mixer.resources.limits.memory: 1Gi
  mixer.resources.requests.memory: 256Mi
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: service-catalog-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: service-catalog
data:
  etcd-stateful.etcd.resources.limits.memory: 512Mi
