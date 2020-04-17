---
title: Configure Runtime Agent with Compass
type: Tutorials
---

This tutorial shows how to configure the Runtime Agent with Compass. 

## Prerequisites

- Compass (version matching the Runtime Agent)
- Runtime connected to Compass and the Runtime ID
- Connector URL
- One-time token from the Connector
- Tenant ID

> **NOTE:** To read more about the required parameteres, see [this](#details-runtime-agent-initializing-connection) document.

## Steps

To configure the Runtime Agent with Compass, you need to create a ConfigMap in the Runtime Agent Deployment. The default deployment is `compass-agent-configuration`. To create the ConfigMap, run:

```
cat <<EOF | kubectl -n compass-system apply -f -
apiVersion: v1
data:
  CONNECTOR_URL: {CONNECTOR_URL}
  RUNTIME_ID: {RUNTIME_ID}
  TENANT: {TENANT_ID}
  TOKEN: {ONE_TIME_TOKEN}
kind: ConfigMap
metadata:
  name: compass-agent-configuration
  namespace: compass-system
EOF
```