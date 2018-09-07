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
  name: installation-config
  namespace: kyma-installer
data:
  is_local_installation: "__IS_LOCAL_INSTALLATION__"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: installation-config-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
data:
  global.isLocalEnv: "__IS_LOCAL_INSTALLATION__"
  gateways.istio-ingressgateway.service.externalPublicIp: "__EXTERNAL_PUBLIC_IP__"
  global.domainName: "__DOMAIN__"
  nginx-ingress.controller.service.loadBalancerIP: "__REMOTE_ENV_IP__"    
  configurations-generator.kubeConfig.clusterName: "__DOMAIN__"
  cluster-users.users.adminGroup: "__ADMIN_GROUP__"
  global.etcdBackup.containerName: "__ETCD_BACKUP_ABS_CONTAINER_NAME__"
  global.etcdBackup.enabled: "__ENABLE_ETCD_BACKUP__"
  global.alertTools.credentials.slack.apiurl: "__SLACK_API_URL_VALUE__"
  global.alertTools.credentials.slack.channel: "__SLACK_CHANNEL_VALUE__"
  global.alertTools.credentials.victorOps.routingkey: "__VICTOR_OPS_ROUTING_KEY_VALUE__"
  global.alertTools.credentials.victorOps.apikey: "__VICTOR_OPS_API_KEY_VALUE__"
  "gateways.istio-ingressgateway.type": "__INGRESSGATEWAY_SERVICE_TYPE__"
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
