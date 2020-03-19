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
  configuration.volumeSnapshotLocation.bucket: "__BSL_BUCKET__"
  configuration.volumeSnapshotLocation.config.apiTimeout: "__VSL_CONFIG_APITIMEOUT__"
  configuration.volumeSnapshotLocation.config.resourceGroup: "__VSL_CONFIG_RESOURCEGROUP__"
  configuration.backupStorageLocation.bucket: "__BSL_BUCKET__"
  configuration.backupStorageLocation.config.resourceGroup: "__BSL_CONFIG_RESOURCEGROUP__"
  configuration.backupStorageLocation.config.storageAccount: "__BSL_CONFIG_STORAGEACCOUNT__"
  credentials.secretContents.cloud: "__CLOUD_CREDENTIALS_FILE_CONTENT_BASE64__"
