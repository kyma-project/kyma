# Issues with APIRules and Service Connection - Basic Diagnostics

API Gateway is a Kubernetes controller, which operates on APIRule custom resources (CRs). See [Issues When Creating an APIRule in Version v2](./03-40-api-rule-troubleshooting.md).

To diagnose problems, inspect the status code of the APIRule CR:

   ```bash
   kubectl describe apirules.gateway.kyma-project.io -n {NAMESPACE} {APIRULE_NAME}
   ```

If the status is `ERROR`, edit the APIRule and fix the issues described in the **status.description** field.