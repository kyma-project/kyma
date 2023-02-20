---
title: Kyma Gateway - not reachable
---

## Symptom

You cannot access services or Functions using the created APIRules. Kyma Gateway refuses the connection.

## Cause

The issue comes up if you either rename or duplicate the default `kyma-gateway`. Once you have multiple Gateway custom resources (CRs) pointing to the same host, the first Gateway CR created takes precedence over the others.

## Remedy

Having two Gateway CRs pointing to the same host is not recommended. To resolve the issue, choose one of the following solutions:

- Make sure the default `kyma-gateway` exists and is not renamed or duplicated.

- If there are multiple Gateway CRs pointing to the same host, delete the duplicated Gateway CR. To delete the Gateway CR, run:

   ```bash
   kubectl -n kyma-system delete gateway $DUPLICATED_GATEWAY_NAME
   ```
