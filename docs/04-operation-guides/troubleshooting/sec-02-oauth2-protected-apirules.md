---
title: No access to OAuth2 protected API Rules
---

## Symptom

If you upgraded to Kyma 2.0 and use the evaluation profile, you could lose access to your OAuth2 procted API Rules. You may get `401 Unauthorized` with the `client_id unknown` error when fetching a token for your OAuth2 clients.

## Cause

## Remedy

Restart the Ory Hydra Maester Pods to trigger Ory to re-fetch the OAuth2 clients.
