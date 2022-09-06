---
title: Acquiring NATS server system account credentials
---
## Context

Accessing certain resources in NATS requires [`system_account` privileges](https://docs.nats.io/running-a-nats-service/configuration/sys_accounts). Kyma will automatically generate a `system account` user via a Secret dubbed `eventing-nats-secret` in the `kyma-system` Namespace.

## Prerequisites

None.

## Procedure

To acquire the cretentials run the following command:

```bash
kubectl get secrets -n kyma-system eventing-nats-secret -oyaml | grep -e accountsJson | awk '{print $2}' | base64 -d | grep {user: | awk '{$1=$1};1' | awk '{print substr($0, 2, length($0) - 2)}'
```

This will return the credentials for the `system account` user.
```bash
user: admin, password: <your password>
```
Result: 
To access the NATS server with the [nats-cli tool](https://github.com/nats-io/natscli) you need to first forward its port:
```bash
kubectl port-forward -n kyma-system svc/eventing-nats 4222
```
Next steps:
Now you can send your nats commands by passing the credentials:
```bash
nats server info --user admin --password <your password>
```
