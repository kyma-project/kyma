---
title: Configure Runtime Agent with Compass
type: Tutorials
---

This tutorial shows how to configure the Runtime Agent with Compass.

## Prerequisites

- [Compass](https://github.com/kyma-incubator/compass)
- Runtime connected to Compass and the Runtime ID
- [Connector URL](#tutorials-establish-a-secure-connection-with-compass)
- One-time token from the Connector
- Tenant ID

> **NOTE:** Learn also about the [parameters required](#details-runtime-agent-initializing-connection) to initialize the connection between the Runtime Agent and Compass.

## Steps

To configure the Runtime Agent with Compass, you need to create a Secret in the Runtime Agent Namespace and specify it in the Runtime Agent Deployment. The default Secret is `compass-agent-configuration`. To create the Secret, run:

```bash
cat <<EOF | kubectl -n compass-system apply -f -
apiVersion: v1
data:
  CONNECTOR_URL: $({CONNECTOR_URL})
  RUNTIME_ID: $({RUNTIME_ID})
  TENANT: $({TENANT_ID})
  TOKEN: $({ONE_TIME_TOKEN})
kind: Secret
metadata:
  name: compass-agent-configuration
  namespace: compass-system
EOF
```
