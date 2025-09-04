# APIRules Are Missing from Kyma Dashboard

If you created APIRules using version `v1beta1`, and you have not yet migrated them to version `v2`,
they are not displayed in the Kyma dashboard's `API Rules` view.
To display the APIRules in Kyma dashboard, you must migrate them to version `v2`.

## Symptom

Kyma dashboard's `API Rules` view does not display APIRules created in version `v1beta1`.
To check if the APIRules are present in the cluster, run the following command:

```bash
kubectl get apirules.v2.gateway.kyma-project.io -A
```
This command lists all APIRules available in your Kyma cluster, regardless of their original version.
To get a specific APIRule and check the version in which it was created, run the following command:
```yaml
kubectl get apirules.v2.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -o yaml
```

If the APIRule was created using version `v1beta1`,
the output contains the annotation `gateway.kyma-project.io/original-version: v1beta1`.

See the following example: 

```yaml
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  annotations:
    gateway.kyma-project.io/original-version: v1beta1
  name: httpbin
  namespace: test
spec:
    gateway: kyma-system/kyma-gateway
    hosts:
        - httpbin.local.kyma.dev
    service:
        name: httpbin
        namespace: test
        port: 8000
status:
    lastProcessedTime: "2025-04-25T11:16:11Z"
    state: Warning
```

## Cause

APIRules that are not displayed in Kyma dashboard were originally created using version `v1beta1`, and you haven't yet migrated them to version `v2`.
The APIRule `v1beta1` API is no longer available via Kyma dashboard.

## Solution

To make sure that support for your APIRules is maintained, you must migrate them to version `v2`.
To learn how to do this, follow the [APIRule migration guide](../../apirule-migration/README.md).
