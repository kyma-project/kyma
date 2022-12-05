---
title: Access to Secrets mounted as Volume
---

This tutorial shows how to use Secrets mounted as volume with the Serverless Function.
It's based on a simple Function in Python 3.9. The Function reads data from Secret and returns it.

## Prerequisites

Before you start, make sure you have these tools installed:

- [Kyma installed](../../04-operation-guides/operations/02-install-kyma.md) on a cluster

## Steps

Follow these steps:

1. Export these variables:

    ```bash
    export FUNCTION_NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    export DOMAIN={DOMAIN_NAME}

    export SECRET_NAME={SECRET_NAME}
    export SECRET_DATA_KEY={SECRET_DATA_KEY}
    export SECRET_MOUNT_PATH={SECRET_MOUNT_PATH}
    ```

2. Create a Secret:

    ```bash
    kubectl -n $NAMESPACE create secret generic $SECRET_NAME \
        --from-literal=$SECRET_DATA_KEY={SECRET_DATA_VALUE}
    ```
   
3. Create your Function with `secretMounts`:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha2
    kind: Function
    metadata:
      name: $FUNCTION_NAME
      namespace: $NAMESPACE
    spec:
      runtime: python39
      source:
        inline:
          source: |
            from os import path
            BASE_PATH = "$SECRET_MOUNT_PATH"
            DATA_PATH = path.join(BASE_PATH, "$SECRET_DATA_KEY")
            def main(event, context):
              with open(DATA_PATH, "r") as f:
                data = f.read()
                return data
      secretMounts:
        - secretName: $SECRET_NAME
          mountPath: $SECRET_MOUNT_PATH
    EOF
    ```

   >**NOTE:** Read more about [creating Functions](svls-01-create-inline-function.md).

4. Create an APIRule:

    The following steps allow you to test the Function in action.

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: $FUNCTION_NAME
      namespace: $NAMESPACE
    spec:
      gateway: kyma-system/kyma-gateway
      host: $FUNCTION_NAME.$DOMAIN
      rules:
        - path: /.*
          accessStrategies:
            - config: {}
              handler: noop
          methods:
            - GET
            - POST
            - PUT
            - DELETE
      service:
        name: $FUNCTION_NAME
        port: 80
    EOF
    ```

   >**NOTE:** Read more about [exposing Functions](svls-03-expose-function.md).

5. Call Function:

    ```bash
    curl https://$FUNCTION_NAME.$DOMAIN
    ```

    You should get `{SECRET_DATA_VALUE}` as a result.

6. Next steps:

    Now you can edit the Secret and see if the Function returns the new value from the Secret.
    
    To edit your Secret, use:
    ```bash
    kubectl -n $NAMESPACE edit secret $SECRET_NAME
    ```
   
    To encode values used in `data` from the Secret, use `base64`, for example:
    ```bash
    echo -n '{NEW_SECRET_DATA_VALUE}' | base64
    ```

    Calling the Function again (using `curl`) must return `{NEW_SECRET_DATA_VALUE}`. 
    Note that the Secret propagation may take some time, and the call may initially return the old value.

