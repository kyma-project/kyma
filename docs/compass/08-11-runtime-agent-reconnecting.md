---
title: Reconnect Runtime Agent with Compass
type: Tutorials
---

This tutorial shows how to reconnect the Runtime Agent with Compass after the established connection was lost.

## Prerequisites

- Compass (version matching the Runtime Agent)
- [ConfigMap created](#tutorials-configure-runtime-agent-with-compass)

## Steps

To force the Runtime Agent to reconnect using the parameters from the ConfigMap, delete the Compass Connection CR:

```bash
kubectl delete compassconnection compass-connection
```

After the Connection CR is removed, the Runtime Agent will try to connect to Compass using the token from the ConfigMap.