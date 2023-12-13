# Reconnect Runtime Agent With Compass

This tutorial shows how to reconnect Runtime Agent with Compass after the established connection was lost.

## Prerequisites

- [Compass](https://github.com/kyma-incubator/compass)
- [ConfigMap created](ra-04-configure-runtime-agent-with-compass.md)

## Steps

To force Runtime Agent to reconnect using the parameters from the Secret, delete the CompassConnection CR:

```bash
kubectl delete compassconnection compass-connection
```

After the Connection CR is removed, Runtime Agent will try to connect to Compass using the token from the Secret.
