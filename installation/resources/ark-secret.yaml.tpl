apiVersion: v1
kind: Secret
metadata:
  name: ark-credentials-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
type: Opaque
data:
  configuration.persistentVolumeProvider.name: "__CLOUD_PROVIDER__"
  configuration.persistentVolumeProvider.config.apiTimeout: "__PVP_CONFIG_APITIMEOUT__"
  configuration.backupStorageProvider.name: "__CLOUD_PROVIDER__"
  configuration.backupStorageProvider.bucket: "__BSP_BUCKET__"
  credentials.secretContents.AZURE_SUBSCRIPTION_ID: "__AZURE_SUBSCRIPTION_ID__"
  credentials.secretContents.AZURE_TENANT_ID: "__AZURE_TENANT_ID__"
  credentials.secretContents.AZURE_RESOURCE_GROUP: "__AZURE_RESOURCE_GROUP__"
  credentials.secretContents.AZURE_CLIENT_ID: "__AZURE_CLIENT_ID__"
  credentials.secretContents.AZURE_CLIENT_SECRET: "__AZURE_CLIENT_SECRET__"
  credentials.secretContents.AZURE_STORAGE_ACCOUNT_ID: "__AZURE_STORAGE_ACCOUNT_ID__"
  credentials.secretContents.AZURE_STORAGE_KEY: "__AZURE_STORAGE_KEY__"
  credentials.secretContents.cloud: "__CLOUD_CREDENTIALS_FILE_CONTENT_BASE64__"