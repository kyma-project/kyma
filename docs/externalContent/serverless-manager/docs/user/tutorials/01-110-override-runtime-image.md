# Override Runtime Image

This tutorial shows how to build a custom runtime image and override the Function's base image with it.

## Prerequisites

Before you start, make sure you have these tools installed:

- [Serverless module installed](https://kyma-project.io/docs/kyma/latest/04-operation-guides/operations/08-install-uninstall-upgrade-kyma-module/) in a cluster

## Steps

Follow these steps:

1. Follow [this example](https://github.com/kyma-project/serverless/tree/main/examples/custom-serverless-runtime-image) to build the Python's custom runtime image.

<Tabs>
<Tab name="Kyma CLI">

2. Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    export RUNTIME_IMAGE_URL={RUNTIME_IMAGE_URL} # image pull URL; for example {dockeruser}/foo:0.1.0
    ```

3. Create your local development workspace using the built image:

    ```bash
    mkdir {FOLDER_NAME}
    cd {FOLDER_NAME}
    kyma alpha function init
    ```

4. Deploy your Function:

    ```bash
    kyma alpha function create $NAME \
      --namespace $NAMESPACE --runtime python312 \
      --runtime-image-override $RUNTIME_IMAGE_URL \
      --source handler.py --dependencies requirements.txt
    ```

5. Verify whether your Function is running:

    ```bash
    kyma alpha function get $NAME --namespace $NAMESPACE
    ```
</Tab>
<Tab name="kubectl">

2. Export these variables:

    ```bash
    export NAME={FUNCTION_NAME}
    export NAMESPACE={FUNCTION_NAMESPACE}
    export RUNTIME_IMAGE_URL={RUNTIME_IMAGE_URL} # image pull URL; for example {dockeruser}/foo:0.1.0
    ```

3. Create a Function CR that specifies the Function's logic:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: serverless.kyma-project.io/v1alpha2
   kind: Function
   metadata:
     name: $NAME
     namespace: $NAMESPACE
   spec:
     runtime: python312
     runtimeImageOverride: $RUNTIME_IMAGE_URL
     source:
       inline:
         source: |
           def main(event, context):
             return "hello world"
   EOF
   ```

4. Verify whether your Function is running:

    ```bash
    kubectl get functions $NAME -n $NAMESPACE
    ```
</Tab>
</Tabs>
