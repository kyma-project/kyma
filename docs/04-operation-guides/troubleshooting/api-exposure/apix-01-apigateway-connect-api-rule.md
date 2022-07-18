---
title: Cannot connect to a service exposed by an API Rule - basic diagnostics
---

API Gateway is a Kubernetes controller, which operates on API Rule custom resources (CRs). To diagnose problems, inspect the [`status`](../../../05-technical-reference/00-custom-resources/apix-01-apirule.md#status-codes) field of the API Rule CR:

   ```bash
   kubectl describe apirules.gateway.kyma-project.io -n {NAMESPACE} {APIRULE_NAME}
   ```

If the status is `Error`, edit the API Rule and fix issues described in `.Status.APIRuleStatus.Desc`. If you still encounter issues, make sure the API Gateway, Hydra, and Oathkeeper are running or take a look at one of the more specific troubleshooting guides:

- [Cannot connect to a service exposed by an API Rule - `401 Unauthorized` or `403 Forbidden`](apix-02-401-unauthorized-403-forbidden.md)
- [Cannot connect to a service exposed by an API Rule - `404 Not Found`](apix-03-404-not-found.md)
