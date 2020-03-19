---
title: Create a lambda
type: Tutorials
---

This tutorial shows how you can create a simple "Hello World!" lambda.

## Steps

Follows these steps:

1. Export these variables:

    ```bash
    export DOMAIN={DOMAIN_NAME}
    export NAME={LAMBDA_NAME}
    export NAMESPACE={LAMBDA_NAMESPACE}
    ```

2. Create a Function CR that specifies the lambda's logic and defines a runtime on which it should run:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha1
    kind: Function
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      functionContentType: plaintext
      runtime: nodejs8
      function: |
        module.exports = {
          main: function(event, context) {
            return 'Hello World!'
          }
        }
    EOF    
    ```

3. Check if your lambda was created successfully and has the `Running` status:

    ```bash
    kubectl get functions $NAME -n $NAMESPACE -o=jsonpath='{.status.condition}'
    ```
