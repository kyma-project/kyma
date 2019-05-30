apiVersion: v1
kind: Secret
metadata:
  name: velero-credentials-overrides
  namespace: kyma-installer
  labels:
    kyma-project.io/installation: ""
    installer: overrides
type: Opaque
data:
  configuration.provider: "__CLOUD_PROVIDER__"
  configuration.volumeSnapshotLocation.name: "__CLOUD_PROVIDER__"
  configuration.volumeSnapshotLocation.bucket: "__BSL_BUCKET__"
  configuration.volumeSnapshotLocation.provider: "__CLOUD_PROVIDER__"
  configuration.volumeSnapshotLocation.config.apiTimeout: "__VSL_CONFIG_APITIMEOUT__"
  configuration.volumeSnapshotLocation.config.resourceGroup: "__VSL_CONFIG_RESOURCEGROUP__"
  configuration.backupStorageLocation.provider: "__CLOUD_PROVIDER__"
  configuration.backupStorageLocation.name: "__CLOUD_PROVIDER__"
  configuration.backupStorageLocation.bucket: "__BSL_BUCKET__"
  configuration.backupStorageLocation.config.resourceGroup: "__BSL_CONFIG_RESOURCEGROUP__"
  configuration.backupStorageLocation.config.storageAccount: "__BSL_CONFIG_STORAGEACCOUNT__"
  credentials.secretContents.AZURE_SUBSCRIPTION_ID: "__AZURE_SUBSCRIPTION_ID__"
  credentials.secretContents.AZURE_TENANT_ID: "__AZURE_TENANT_ID__"
  credentials.secretContents.AZURE_RESOURCE_GROUP: "__AZURE_RESOURCE_GROUP__"
  credentials.secretContents.AZURE_CLIENT_ID: "__AZURE_CLIENT_ID__"
  credentials.secretContents.AZURE_CLIENT_SECRET: "__AZURE_CLIENT_SECRET__"
  credentials.secretContents.cloud: "__CLOUD_CREDENTIALS_FILE_CONTENT_BASE64__"