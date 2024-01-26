# Configure Runtime Agent with Compass

This tutorial shows how to configure Runtime Agent with Compass.

## Prerequisites

- [Compass](https://github.com/kyma-incubator/compass)
- Runtime connected to Compass and the Runtime ID
- [Connector URL](ra-01-establish-secure-connection-with-compass.md)
- One-time token from the Connector
- Tenant ID

> [!NOTE]
> Learn also about the [parameters required](../../05-technical-reference/00-configuration-parameters/ra-01-connection-with-compass.md) to initialize the connection between Runtime Agent and Compass.

## Steps

To configure Runtime Agent with Compass, you need to create a Secret in the Runtime Agent namespace and specify it in the Runtime Agent Deployment. The default Secret is `compass-agent-configuration`. To create the Secret, run:

```bash
cat <<EOF | kubectl -n kyma-system apply -f -
apiVersion: v1
data:
  CONNECTOR_URL: $({CONNECTOR_URL})
  RUNTIME_ID: $({RUNTIME_ID})
  TENANT: $({TENANT_ID})
  TOKEN: $({ONE_TIME_TOKEN})
kind: Secret
metadata:
  name: compass-agent-configuration
  namespace: kyma-system
EOF
```
