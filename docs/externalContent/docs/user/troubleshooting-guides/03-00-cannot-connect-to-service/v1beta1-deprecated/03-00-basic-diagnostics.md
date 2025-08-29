# Issues with APIRules and Service Connection - Basic Diagnostics

API Gateway is a Kubernetes controller, which operates on APIRule custom resources (CRs). See [Issues When Creating an APIRule in version v1beta1](./03-00-basic-diagnostics.md).

To diagnose problems, inspect the status code of the APIRule CR:

   ```bash
   kubectl describe apirules.gateway.kyma-project.io -n {NAMESPACE} {APIRULE_NAME}
   ```

If the status is `ERROR`, edit the APIRule and fix the issues described in the **.Status.APIRuleStatus.desc** field. If you still encounter issues, make sure that API Gateway and Oathkeeper are running, or take a look at one of the more specific troubleshooting guides:

- [Cannot connect to a Service exposed by an APIRule - `404 Not Found`](./03-02-404-not-found.md)
- [Cannot connect to a Service exposed by an APIRule - `401 Unathorized or 403 Forbidden`](./03-01-401-unauthorized-403-forbidden.md)
- [Cannot connect to a Service exposed by an APIRule - `500 Internal Server Error`](./03-03-500-server-error.md)