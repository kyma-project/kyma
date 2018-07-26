apiVersion: v1
kind: Secret
metadata:
  name: etcd-backup-abs-credentials
  namespace: kyma-installer
type: Opaque
data:
  storage-account: "__ETCD_BACKUP_ABS_ACCOUNT__"
  storage-key: "__ETCD_BACKUP_ABS_KEY__"
