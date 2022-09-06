---
title: Acquiring NATS server system account credentials
---
## Context

Accessing certain resources in NATS requires [`system_account` privileges](https://docs.nats.io/running-a-nats-service/configuration/sys_accounts). Kyma automatically generates a `system account` user using a Secret dubbed `eventing-nats-secret` in the `kyma-system` Namespace.

## Prerequisites

None.

## Procedure

To acquire the credentials, run the following command:

```bash
kubectl get secrets -n kyma-system eventing-nats-secret -oyaml | grep -e accountsJson | awk '{print $2}' | base64 -d | grep {user: | awk '{$1=$1};1' | awk '{print substr($0, 2, length($0) - 2)}'
```

### Result: 
You got the credentials for the `system account` user in the following format:
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
