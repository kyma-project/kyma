---
title: Basic diagnostics
---

API Gateway is a Kubernetes controller, which operates on APIRule custom resources (CRs). To diagnose problems, inspect the [`status` code](../../../../05-technical-reference/00-custom-resources/apix-01-apirule.md#status-codes) of the APIRule CR:

   ```bash
   kubectl describe apirules.gateway.kyma-project.io -n {NAMESPACE} {APIRULE_NAME}
   ```

If the status is `Error`, edit the APIRule and fix the issues described in the **.Status.APIRuleStatus.desc** field. If you still encounter issues, make sure that API Gateway and Oathkeeper are running, or take a look at one of the more specific troubleshooting guides:

- [Cannot connect to a service exposed by an APIRule - `404 Not Found`](./apix-01-03-404-not-found.md)
- [Cannot connect to a service exposed by an APIRule - `500 Internal Server Error`](./apix-01-04-500-server-error.md)