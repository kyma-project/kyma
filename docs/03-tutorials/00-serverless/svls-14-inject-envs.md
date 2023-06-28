---
title: Inject environment variables
---

This tutorial shows how to inject environment variables into Function.

You can specify environment variables in the Function definition, or define references to the Kubernetes Secrets or ConfigMaps.

## Prerequisites

Before you start, make sure you have these tools installed:

- [Kyma installed](../../04-operation-guides/operations/02-install-kyma.md) on a cluster

## Steps

Follow these steps:

1. Create your ConfigMap

```bash
kubectl create configmap my-config --from-literal config-env="I come from config map"
```

2. Create your Secret

```bash
kubectl create secret generic  my-secret  --from-literal secret-env="I come from secret"
```


<div tabs name="steps" group="create-function">
  <details>
  <summary label="cli">
  Kyma CLI
  </summary>

3. Generate the Function's configuration and sources:

    ```bash
    kyma init function --name my-function
    ```

4. Define environment variables as part of the Function configuration file. Modify `config.yaml` with the following:
    ```yaml
    name: my-function
    namespace: default
    runtime: nodejs18
    source:
        sourceType: inline
    env:
      - name: env1
        value: "I come from function definition"
      - name: env2
        valueFrom:
          configMapKeyRef:
            name: my-config
            key: config-env
      - name: env3
        valueFrom:
          secretKeyRef:
            name: my-secret
            key: secret-env
    ```
5. Use injected environment variables in the handler file. Modify `handler.js` with the following:
    ```js
    module.exports = {
        main: function (event, context) {
            envs = ["env1", "env2", "env3"]
            envs.forEach(function(key){
                console.log(`${key}:${readEnv(key)}`)
            });
            return 'Hello Serverless'
        }
    }

    readEnv=(envKey) => {
        if(envKey){
            return process.env[envKey];
        }
        return
    }
    ```

6. Deploy your Function:

    ```bash
    kyma apply function
    ```

7. Verify whether your Function is running:

    ```bash
    kubectl get functions my-function
    ```

  </details>
  <details>
  <summary label="kubectl">
  kubectl
  </summary>


3. Create a Function CR that specifies the Function's logic:

   ```yaml
   cat <<EOF | kubectl apply -f -
   apiVersion: serverless.kyma-project.io/v1alpha2
   kind: Function
   metadata:
     name: my-function
   spec:
     env:
       - name: env1
         value: I come from function definition
       - name: env2
         valueFrom:
           configMapKeyRef:
             key: config-env
             name: my-config
       - name: env3
         valueFrom:
           secretKeyRef:
             key: secret-env
             name: my-secret
     runtime: nodejs18
     source:
       inline:
         source: |-
           module.exports = {
               main: function (event, context) {
                   envs = ["env1", "env2", "env3"]
                   envs.forEach(function(key){
                       console.log(\`${key}:${readEnv(key)}\`)
                   });
                   return 'Hello Serverless'
               }
           }
           readEnv=(envKey) => {
               if(envKey){
                   return process.env[envKey];
               }
               return
           }
   EOF
   ```

4. Verify whether your Function is running:

    ```bash
    kubectl get functions my-function
    ```

</details>
</div>

