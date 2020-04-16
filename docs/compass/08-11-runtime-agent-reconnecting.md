---
title: Reconnect Runtime Agent with Compass
type: Tutorials
---

This tutorial shows how to reconnect Runtime Agent with Compass after an established connection was lost.

## Prerequisites

- Compass (version matching Runtime Agent)
- [ConfigMap created](#tutorials-configure-runtime-agent-with-compass)

## Steps

To force Runtime Agent to reconnect using the parameters from the ConfigMap, delete the Compass Connection CR:

```
kubectl delete compassconnection compass-connection
```

After the Connection CR is removed, Runtime Agent will try to connect to Compass using the token from the ConfigMap.