---
title: Runtime Agent
type: Details
---

Runtime Agent is a Kyma component that connects to Compass. 

The main responsibilities of the component are:
- [Establishing a trusted connection](08-07-establish-secure-connection-with-compass.md) between the Kyma Runtime and Compass
- [Renewing a trusted connection](08-08-maintain-secure-connection-with-compass.md) between the Kyma Runtime and Compass
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

Runtime Agent reads this configuration from the Config Map specified in the Runtime Agent Deployment (`compass-agent-configuration` by default).

To create the Config Map, run:
```
cat <<EOF | kubectl -n compass-system apply -f -
apiVersion: v1
data:
  CONNECTOR_URL: {COMPASS_CONNECTOR_URL}
  RUNTIME_ID: {RUNTIME_ID}
  TENANT: {TENANT_ID}
  TOKEN: {ONE_TIME_TOKEN}
kind: ConfigMap
metadata:
  name: compass-agent-configuration
  namespace: compass-system
EOF
```

### Connection status

Connection status is preserved in the CompassConnection Custom Resource (CR).


The CompassConnection CR contains the following Compass URLs:
- `connectorUrl` - the URL of the Connector used for maintaining secure connection.
- `directorUrl` - the URL of the Director used for fetching Applications.


The CompassConnection CR contains two statuses, `connectionStatus` and `synchronizationStatus`, with their respective parameters.
 
 
 `connectionStatus` contains the following parameters:

| **Parameter**         | **Description**                                           |
|-----------------------|-----------------------------------------------------------|
| `certificateStatus`   | Date of when the certificate was issued and when it expires |
| `established`         | Date of when the connection was established               |
| `lastSuccess`         | Last successful synchronization with the Connector        |
| `lastSync`            | Last synchronization attempt                              |
| `renewed`             | Last time the certificate was renewed                     |


 `synchronizationStatus` contains the following parameters:

| **Parameter**         | **Description**                                           |
|-----------------------|-----------------------------------------------------------|
| `lastAttempt`         | Last synchronization attempt with the Director            |
| `lastSuccessfulFetch` | Last successful fetch of resources from the Director      |
| `lastSuccessfulApplication` | Last successful application of resources fetched from Compass |


### Reconnecting Runtime Agent

If the connection with Compass fails, Runtime Agent will keep trying to connect with the token from the Config Map. If the connection is established successfully, Runtime Agent will ignore the Config Map until the connection is lost. 

To force Runtime Agent to reconnect using the parameters from the Config Map, delete the Compass Connection CR:

```
kubectl delete compassconnection compass-connection
```
