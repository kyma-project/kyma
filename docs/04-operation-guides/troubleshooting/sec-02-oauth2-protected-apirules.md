---
title: No access to OAuth2 protected API Rules
---

## Symptom

If you use Kyma 2.0 in the evaluation profile, you could lose access to your OAuth2 procted API Rules. You may get the 401 Unauthorized with "client_id unknown" error when fetching a token for your OAuth2Clients.

## Cause

## Remedy

Restart the ory-hydra-maester Pods to trigger ORY to re-fetch the OAuth2Clients.
