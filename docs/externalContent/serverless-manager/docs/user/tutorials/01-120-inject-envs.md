# Inject Environment Variables

This tutorial shows how to inject environment variables into Function.

You can specify environment variables in the Function definition, or define references to the Kubernetes Secrets or ConfigMaps.

## Prerequisites

Before you start, make sure you have these tools installed:

- [Serverless module installed](https://kyma-project.io/docs/kyma/latest/04-operation-guides/operations/08-install-uninstall-upgrade-kyma-module/) in a cluster

## Steps

Follow these steps:

1. Create your ConfigMap

```bash
kubectl create configmap my-config --from-literal config-env="I come from config map"
```

2. Create your Secret

```bash
kubectl create secret generic my-secret --from-literal secret-env="I come from secret"
```

<Tabs>
<Tab name="Kyma CLI">

> [!WARNING]
> This section is not yet compliant with Kyma CLI v3.
</Tab>
<Tab name="kubectl">

3. Create a Function CR that specifies the Function's logic:

   ```bash
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
     runtime: nodejs20
     source:
       inline:
         source: |-
           module.exports = {
               main: function (event, context) {
                   envs = ["env1", "env2", "env3"]
                   envs.forEach(function(key){
                       console.log(key+":"+readEnv(key))
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
</Tab>
</Tabs>

## Redis based example

For more use cases, see [this example](https://github.com/kyma-project/serverless/tree/main/examples/redis-rest)
