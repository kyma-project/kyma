---
title: Override runtime image
---

This tutorial show how to build custom runtime image and override functions base image with thee built one

## Prerequisites

Before you start, make sure you have these tools installed:

- Kyma installed on a cluster

## Steps

Follows these steps:

1. Follow this example to build python's custom runtime image.

<div tabs name="steps" group="create-function">
  <details>
  <summary label="cli">
  Kyma CLI
  </summary>

2. Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    export RUNTIME_IMAGE={RUNTIME_IMAGE_WITH_TAG}
    ```

3. Create your local development workspace using built image:

    ```bash
    mkdir {FOLDER_NAME}
    cd {FOLDER_NAME}
    kyma init function --name $NAME --namespace $NAMESPACE --runtime-image-override $RUNTIME_IMAGE --runtime python39
    ```

4. Deploy function:

    ```bash
    kyma apply function
    ```

5. Verify if the function is running:

    ```bash
    kubectl get functions $NAME -n $NAMESPACE
    ```

  </details>
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

2. Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    export RUNTIME_IMAGE={RUNTIME_IMAGE_WITH_TAG}
    ```

3. Create a Function CR that specifies the Function's logic:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha1
    kind: Function
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      runtime: python39
      runtimeImageOverride: $RUNTIME_IMAGE
      source: |
        module.exports = {
          main: function(event, context) {
            return 'Hello World!'
          }
        }
    EOF
    ```

4. Verify if the function is running:

    ```bash
    kubectl get functions $NAME -n $NAMESPACE
    ```

</details>
</div>
