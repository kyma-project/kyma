---
title: No access to OAuth2 protected APIRules
---

## Symptom

If you upgraded to Kyma 2.0 and use the evaluation profile, you could lose access to your OAuth2 protected APIRules. You may get `401 Unauthorized` with the `client_id unknown` error when fetching a token for your OAuth2 Client resources.

## Cause

Kyma 2.0 comes with a bumped Ory Hydra version. The update enforced a restart of Ory Hydra and Ory Hydra Maester Pods. As Ory Hydra in the evaluation profile uses an in-memory database (IMBD), the previously created OAuth2 client resources might be no longer available in the Hydra database.

## Remedy

Restart the Ory Hydra Maester Pods to trigger Ory to recreate the OAuth2 client resources. Use the following command:

```bash
kubectl rollout restart deployment ory-hydra-maester -n kyma-system
```
