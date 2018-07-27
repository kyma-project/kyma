##################################################
# ConfigMap for Installer (Template)             #
#                                                #
# This map is used to populate                   #
# environment variables for Kyma Installer.      #
#                                                #
##################################################
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
  ectd_backup_abs_container_name: "__ECTD_BACKUP_ABS_CONTAINER_NAME__"
  components: "cluster-prerequisites,tiller,cluster-essentials,istio,prometheus-operator,provision-bundles,dex,core,remote-environments" 
