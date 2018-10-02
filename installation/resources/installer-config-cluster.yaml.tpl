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
kind: Secret
metadata:
  name: ui-test-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
type: Opaque
data:
  test.auth.username: "__UI_TEST_USER__"
  test.auth.password: "__UI_TEST_PASSWORD__"
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
  global.alertTools.credentials.slack.apiurl: "__SLACK_API_URL_VALUE__"
  global.alertTools.credentials.slack.channel: "__SLACK_CHANNEL_VALUE__"
  global.alertTools.credentials.victorOps.routingkey: "__VICTOR_OPS_ROUTING_KEY_VALUE__"
  global.alertTools.credentials.victorOps.apikey: "__VICTOR_OPS_API_KEY_VALUE__"
  nginx-ingress.controller.service.loadBalancerIP: "__REMOTE_ENV_IP__"
  configurations-generator.kubeConfig.clusterName: "__DOMAIN__"
  cluster-users.users.adminGroup: "__ADMIN_GROUP__"
  service-catalog.etcd-stateful.replicaCount: "3"
  minio.accessKey: "admin"
  minio.secretKey: "topSecretKey"
  acceptanceTest.remoteEnvironment.disabled: "true"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ec-default-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: ec-default
data:
  deployment.args.sourceType: commerce
  service.externalapi.nodePort: "32001"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: hmc-default-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: hmc-default
data:
  deployment.args.sourceType: marketing
  service.externalapi.nodePort: "32000"
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
  global.proxy.includeIPRanges: "10.244.0.0/16,10.240.0.0/16"

  gateways.istio-ingressgateway.loadBalancerIP: "__EXTERNAL_PUBLIC_IP__"
  gateways.istio-ingressgateway.type: "LoadBalancer"

  pilot.resources.limits.memory: 2Gi
  pilot.resources.requests.memory: 512Mi
  mixer.resources.limits.memory: 512Mi
  mixer.resources.requests.memory: 128Mi