apiVersion: v1
kind: Secret
metadata:
  name: azure-broker
  namespace: kyma-installer
type: Opaque
data:
  azure_broker_subscription_id: "__AZURE_BROKER_SUBSCRIPTION_ID__"
  azure_broker_tenant_id: "__AZURE_BROKER_TENANT_ID__"
  azure_broker_client_id: "__AZURE_BROKER_CLIENT_ID__"
  azure_broker_client_secret: "__AZURE_BROKER_CLIENT_SECRET__"
