---
title: Kyma Gateway - not reachable
---

## Symptom

The user cannot access services or functions using the APIRules created. The gateway refuses connection.

## Cause

When multiple gateways are created that point to the same host, only the first gateway takes precedence. The issue could be caused when the default `kyma-gateway` is either renamed or duplicated.

## Remedy

It is not recommended having two gateways pointing to the same host. 

- Make sure the default `kyma-gateway` exists and is not renamed or duplicated.

- If there are multiple gateways pointing to the same host, then delete the duplicated gateway and retain only one. You can delete the gateway as follows:

```bash
kubectl -n kyma-system delete gateway $DUPLICATED_GATEWAY_NAME`
```