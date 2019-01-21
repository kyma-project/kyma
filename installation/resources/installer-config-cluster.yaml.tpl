apiVersion: v1
kind: Secret
metadata:
  name: application-connector-certificate-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
type: Opaque
data:
  global.applicationConnectorCa: "__REMOTE_ENV_CA__"
  global.applicationConnectorCaKey: "__REMOTE_ENV_CA_KEY__"
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
  global.domainName: "__DOMAIN__"
  global.applicationConnectorDomainName: "__APPLICATION_CONNECTOR_DOMAIN__"
  global.loadBalancerIP: "__EXTERNAL_PUBLIC_IP__"
  global.etcdBackup.containerName: "__ETCD_BACKUP_ABS_CONTAINER_NAME__"
  global.etcdBackup.enabled: "__ENABLE_ETCD_BACKUP__"
  nginx-ingress.controller.service.loadBalancerIP: "__REMOTE_ENV_IP__"
  cluster-users.users.adminGroup: "__ADMIN_GROUP__"
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
  name: istio-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: istio
data:
  gateways.istio-ingressgateway.loadBalancerIP: "__EXTERNAL_PUBLIC_IP__"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: knative-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: knative
data:
  knative.domainName: "__DOMAIN__"
