---
title: KService revision is missing
type: Troubleshooting
---


If the Function is not running, and the KService is stuck with the `RevisionMissing` reason, try updating the Function. If the update does not solve the problem, follow the steps for a given Kyma release.

## Kyma 1.12

1. Copy the ServiceBindingUsage annotation to an environment variable:
    
    ```bash
    export SBU_ANNOTATION=$(kubectl get services.serving.knative.dev {NAME} -n {NAMESPACE} -o=jsonpath="{.metadata.annotations['servicebindingusages\.servicecatalog\.kyma-project\.io/tracing-information']}")
    ```

2. Delete the KService:
    
    ```bash
    kubectl delete services.serving.knative.dev {NAME} -n {NAMESPACE}
    ```

3. Add the ServiceBindingUsage annotation to the new KService that is automatically created:
    
    ```bash
    if [[ -n "${SBU_ANNOTATION}" ]]; then; kubectl annotate services.serving.knative.dev {NAME} -n {NAMESPACE} "servicebindingusages.servicecatalog.kyma-project.io/tracing-information=${SBU_ANNOTATION}" --overwrite; fi
    ```

## Kyma 1.13 and newer

1. Delete the KService:

    ```bash
    kubectl delete services.serving.knative.dev {NAME} -n {NAMESPACE}
    ```
