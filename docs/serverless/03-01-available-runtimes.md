---
title: Runtimes
---

<div tabs name="available-runtimes" group="available-runtimes">
  <details>
  <summary label="nodejs12">
  Node.js
  </summary>

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha1
    kind: Function
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
    runtime: nodejs12
    source: |
      module.exports = {
        main: function(event, context) {
          return 'Hello World!'
        }
      }
    EOF
    ```

  </details>
  <details>
  <summary label="python38">
  Python38
  </summary>

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha1
    kind: Function
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
    runtime: nodejs12
    source: |
      def main(event, context):
        return 'Hello world!'
    EOF
    ```

</div>
