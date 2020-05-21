---
title: Serverless troubleshooting
type: Troubleshooting
---

## Function stuck in Error and Knative Service is not ready with RevisionMissing reason

In case Knative Service get stuck with RevisionMissing reason and updates on Function doesn't solve that problem, try these steps:

### Kyma 1.12

1. Copy Service Binding Usage annotation to environment variable:
    
    ```bash
    export SBU_ANNOTATION=$(kubectl get services.serving.knative.dev {NAME} -n {NAMESPACE} -o=jsonpath="{.metadata.annotations['servicebindingusages\.servicecatalog\.kyma-project\.io/tracing-information']}")
    ```

2. Delete Knative Service:
    
    ```bash
    kubectl delete services.serving.knative.dev {NAME} -n {NAMESPACE}
    ```
3. Add Service Binding Usage annotation to the new Knative Service that has been created:
    
    ```bash
    if [[ -n "${SBU_ANNOTATION}" ]]; then; kubectl annotate services.serving.knative.dev {NAME} -n {NAMESPACE} "servicebindingusages.servicecatalog.kyma-project.io/tracing-information=${SBU_ANNOTATION}" --overwrite; fi
    ```

### Kyma 1.13 and newer

1. Delete Knative Service:

```bash
kubectl delete services.serving.knative.dev {NAME} -n {NAMESPACE}
```
