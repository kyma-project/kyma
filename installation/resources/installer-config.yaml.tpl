apiVersion: v1
kind: Secret
metadata:
  name: remote-env-certificate
  namespace: kyma-installer
type: Opaque
data:
  remote_env_ca: "__REMOTE_ENV_CA__"
  remote_env_ca_key: "__REMOTE_ENV_CA_KEY__"
---
apiVersion: v1
kind: Secret
metadata:
  name: cluster-certificate
  namespace: kyma-installer
type: Opaque
data:
  tls_cert: "__TLS_CERT__"
  tls_key: "__TLS_KEY__"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: installation-config
  namespace: kyma-installer
data:
  external_ip_address: "__EXTERNAL_IP_ADDRESS__"
  domain: "__DOMAIN__"
  remote_env_ip: "__REMOTE_ENV_IP__"
  k8s_apiserver_url: "__K8S_APISERVER_URL__"
  k8s_apiserver_ca: "__K8S_APISERVER_CA__"
  admin_group: "__ADMIN_GROUP__"
  enable_etcd_backup_operator: "__ENABLE_ETCD_BACKUP_OPERATOR__"
  etcd_backup_abs_container_name: "__ETCD_BACKUP_ABS_CONTAINER_NAME__"
  components: "cluster-prerequisites,cluster-essentials,istio,prometheus-operator,provision-bundles,dex,core,remote-environments" 
