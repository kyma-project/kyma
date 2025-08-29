# APIRule `v2` Contains a Changed Status Schema

## Symptom
There is a changed schema of **status** in an APIRule custom resource (CR), for example:


  ```bash
kubectl get apirules.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -oyaml
  ```
```yaml
  ...
  status:
    lastProcessedTime: "2025-04-25T11:16:11Z"
    state: Ready
```

## Cause
The schema of the **status.state** field in the `v2` APIRule CR introduces a unified approach, similar to the one used in the API Gateway CR.
The possible states of the **status.state** field are  `Ready`, `Warning`, `Error`, `Processing`, or `Deleting`.

## Solution

Get the APIRule in its original version:
  ```bash
  kubectl get apirules.v1beta1.gateway.kyma-project.io -n $NAMESPACE $APIRULE_NAME -oyaml
  ```
```yaml
  ...
  status:
  APIRuleStatus:
    code: OK
  accessRuleStatus:
    code: OK
  lastProcessedTime: "2025-04-25T11:16:11Z"
  observedGeneration: 1
  virtualServiceStatus:
    code: OK  
```