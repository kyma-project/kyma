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
  initContainers.pluginContainer.image: "__PROVIDER_PLUGIN_IMAGE__"
  configuration.provider: "__CLOUD_PROVIDER__"
  configuration.volumeSnapshotLocation.name: "__CLOUD_PROVIDER__"
  configuration.volumeSnapshotLocation.bucket: "__BUCKET__"
  configuration.volumeSnapshotLocation.config.apiTimeout: "__APITIMEOUT__"
  configuration.volumeSnapshotLocation.config.resourceGroup: "__RESOURCEGROUP__"
  configuration.backupStorageLocation.name: "__CLOUD_PROVIDER__"
  configuration.backupStorageLocation.bucket: "__BUCKET__"
  configuration.backupStorageLocation.config.resourceGroup: "__RESOURCEGROUP__"
  configuration.backupStorageLocation.config.storageAccount: "__STORAGEACCOUNT__"
  credentials.secretContents.cloud: "__CLOUD_CREDENTIALS_FILE_CONTENT_BASE64__"
