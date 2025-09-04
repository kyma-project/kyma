# Reconnect Runtime Agent with UCL

This tutorial shows how to reconnect Runtime Agent with UCL after the established connection was lost.

## Prerequisites

- [UCL](https://github.com/kyma-incubator/compass) (previously called Compass)
- [Runtime Agent Configuration using Kubernetes Secret](../tutorials/01-90-configure-runtime-agent-with-compass.md)

## Steps

To force Runtime Agent to reconnect using the parameters from the Secret containing the Runtime Agent Configuration, delete the Connection CR:

```bash
kubectl delete compassconnection compass-connection
```

After the Connection CR is removed, Runtime Agent tries to connect to UCL using the token from the Runtime Agent configuration Secret.
