---
title: Acquiring NATS server system account credentials
---
## Context

Accessing certain resources in NATS requires [`system_account` privileges](https://docs.nats.io/running-a-nats-service/configuration/sys_accounts). Kyma automatically generates a `system account` user using a Secret dubbed `eventing-nats-secret` in the `kyma-system` Namespace.

## Procedure

To acquire the credentials, run the following command:

```bash
kubectl get secrets -n kyma-system eventing-nats-secret -ogo-template='{{index .data "resolver.conf"|base64decode}}'| grep 'user:' | tr -d '{}'
```

### Result: 
You get the credentials for the `system account` user in the following format:
```bash
user: admin, password: <your password>
```
### Next steps:
1. To access the NATS server with the [nats-cli tool](https://github.com/nats-io/natscli), forward its port:
   ```bash
   kubectl port-forward -n kyma-system svc/eventing-nats 4222
2. To send your NATS commands, pass the credentials:
   ```bash
   nats server info --user admin --password <your password>
