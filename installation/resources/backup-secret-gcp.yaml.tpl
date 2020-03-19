apiVersion: v1
kind: Secret
metadata:
  name: backup-credentials-overrides
  namespace: kyma-installer
  labels:
    kyma-project.io/installation: ""
    installer: overrides
    component: backup
type: Opaque
data:
  configuration.provider: "__CLOUD_PROVIDER__"
  configuration.volumeSnapshotLocation.bucket: "__BSL_BUCKET__"
  configuration.backupStorageLocation.bucket: "__BSL_BUCKET__"
  credentials.secretContents.cloud: "__CLOUD_CREDENTIALS_FILE_CONTENT_BASE64__"
