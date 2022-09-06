---
title: Acquiring NATS server system account credentials
---

To access the NATS server with the [nats-cli tool](https://github.com/nats-io/natscli) you need to first forward its port:
```bash
kubectl port-forward -n kyma-system svc/eventing-nats 4222
```
As you may know accessing certain resources in NATS requires [`system_account` privileges](https://docs.nats.io/running-a-nats-service/configuration/sys_accounts). Kyma will automatically generate a `system account` user via a Secret dubbed `eventing-nats-secret` in the `kyma-system` Namespace. To acquire the username and password use:
```bash
kubectl get secrets -n kyma-system eventing-nats-secret -oyaml | grep -e accountsJson | awk '{print $2}' | base64 -d | grep {user: | awk '{$1=$1};1' | awk '{print substr($0, 2, length($0) - 2)}'
```
These credentials can be used to access the aforementioned NATS resources with the cli tool by passing the `--user admin` and `--password <your password>` options, e. g.:
```bash
nats server info --user admin --password <your password>
```

