---
title: Runtime Agent
type: Details
---

Runtime Agent is a Kyma component that connects to Compass. 

The main responsibilities of the component are:
- Establishing a trusted connection between the Kyma Runtime and Compass
- Renewing a trusted connection between the Kyma Runtime and Compass
- Synchronizing with the Director by fetching new Applications from the Director and creating them in the Runtime, and removing from the Runtime Applications that no longer exist in the Director.

### Initializing connection 

Runtime Agent connects to Compass using a one-time token from the Connector and exchanges it for a certificate, which is later used to fetch Applications from the Director. 

The initial connection requires the following parameters:

| **Parameter** | **Description** | **Example value** |
|---------------|-----------------|-------------------|
| **CONNECTOR_URL** | Connector URL | `https://compass-gateway.kyma.local/connector/graphql` |
| **RUNTIME_ID** | ID of the Runtime registered in the Director | `1ae04041-17e5-478f-91f8-3a2ddc7700de` |
| **TENANT** | Tenant ID  | - |
| **TOKEN** | One-time token generated for the Runtime | - |

Runtime Agent reads this configuration from the ConfigMap specified in the Runtime Agent Deployment (`compass-agent-configuration` by default).

To see how to create the ConfigMap, see [this](#tutorials-create-a-configmap) tutorial. 

### Connection status

The connection status is preserved in the [CompassConnection Custom Resource](#custom-resource-compassconnection) (CR). This CR also stores the Connector URL and the Director URL.

### Reconnecting Runtime Agent

If the connection with Compass fails, Runtime Agent keeps trying to connect with the token from the ConfigMap. If the connection is established successfully, Runtime Agent ignores the ConfigMap until the connection is lost. 

To force Runtime Agent to reconnect using the parameters from the ConfigMap, delete the Compass Connection CR:

```
kubectl delete compassconnection compass-connection
```
