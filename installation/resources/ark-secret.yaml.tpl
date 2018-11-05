apiVersion: v1
kind: Secret
metadata:
  name: ark-credentials-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
type: Opaque
data:
  volumeSnapshotLocation.provider: "__CLOUD_PROVIDER__"
  volumeSnapshotLocation.config.apiTimeout: "__VSL_CONFIG_APITIMEOUT__"
  volumeSnapshotLocation.config.resourceGroup: "__VSL_CONFIG_RESOURCEGROUP__"
  backupStorageLocation.provider: "__CLOUD_PROVIDER__"
  backupStorageLocation.objectStorage.bucket: "__BSL_BUCKET__"
  backupStorageLocation.config.resourceGroup: "__BSL_CONFIG_RESOURCEGROUP__"
  backupStorageLocation.config.storageAccount: "__BSL_CONFIG_STORAGEACCOUNT__"
  credentials.secretContents.AZURE_SUBSCRIPTION_ID: "__AZURE_SUBSCRIPTION_ID__"
  credentials.secretContents.AZURE_TENANT_ID: "__AZURE_TENANT_ID__"
  credentials.secretContents.AZURE_RESOURCE_GROUP: "__AZURE_RESOURCE_GROUP__"
  credentials.secretContents.AZURE_CLIENT_ID: "__AZURE_CLIENT_ID__"
  credentials.secretContents.AZURE_CLIENT_SECRET: "__AZURE_CLIENT_SECRET__"
  credentials.secretContents.cloud: "__CLOUD_CREDENTIALS_FILE_CONTENT_BASE64__"