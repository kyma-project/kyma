---
title: Kyma Gateway - not reachable
---

## Symptom

You cannot access services or Functions using the APIRules created. The Kyma Gateway refuses the connection.

## Cause

The issue can come up if you either rename or duplicate the default `kyma-gateway`. Once you have multiple Gateway custom resources (CRs) pointing to the same host, the first Gateway CR created takes precedence over the other ones.

## Remedy

It is not recommended having two Gateway CRs pointing to the same host. To solve the issue, choose one of the proposed solutions:

- Make sure the default `kyma-gateway` exists and is not renamed or duplicated.

- If there are multiple Gateway CRs pointing to the same host, delete the duplicated Gateway CR. To delete the Gateway CR, run:

   ```bash
   kubectl -n kyma-system delete gateway $DUPLICATED_GATEWAY_NAME
   ```
