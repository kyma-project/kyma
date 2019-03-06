
apiVersion: v1
kind: Secret
metadata:
  name: etcd-backup-abs-credentials-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    kyma-project.io/installation: ""
type: Opaque
data:
  etcd-operator.backupOperator.abs.storageAccount: "__ETCD_BACKUP_ABS_ACCOUNT__"
  etcd-operator.backupOperator.abs.storageKey: "__ETCD_BACKUP_ABS_KEY__"