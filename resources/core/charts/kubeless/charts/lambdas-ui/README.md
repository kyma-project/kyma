# Lambdas

## Overview

**Lambdas** is a web-based UI component which is a part of the Console UI.
It allows you to manage lambda functions, which are the key part of Kyma capabilities.

## Details

**Lambdas** allow you to:
- Create and modify lambda functions.
- Expose a function as API.
- Bind a function to a service.
- Write the code of the function. This option is available only through node.js.
- Manage dependencies.
- Add and remove labels.
- Enable API security.
- Select the function size.
- Direct access to Lambda dashboard.

### Configuration

**Lambdas** UI is deployed with the use of a [ConfigMap](templates/configmap.yaml).
The **ConfigMap**  includes the `config.js` file that is mounted as the asset of the Console application and injected as a configuration file.
Use this mechanism to overwrite the default configuration with custom values resulting from the Helm chart installation.

### Function size

You can modify the function size when you either create or update a function using  **Lambdas** UI.

The [ConfigMap](templates/configmap.yaml) contains a set of configurations for a function size, including:
- **resources** configuration with the following parameters: **requests** and **limits**.  
- **horizontalPodAutoscaler** configuration with the following parameters: **minReplicas**, **maxReplicas** and **targetAverageUtilization**.

- The basic **requests** configuration for a function is as follows:

```
resources:
  requests:
    cpu: '100m'
    memory: '100Mi'
```

- You can configure the sizes using the following parameters:
1. S: `{memory: '128Mi', cpu: '100m', minReplicas: 1, maxReplicas: 2}`
2. M: `{memory: '256Mi', cpu: '500m', minReplicas: 2, maxReplicas: 5}`
3. L: `{memory: '512Mi', cpu: '1.0',  minReplicas: 5, maxReplicas: 10}`
4. XL:`{memory: '1024Mi', cpu: '2.0', minReplicas: 10, maxReplicas: 20}`

>**NOTE:** The size of the function may exhaust the resources of the Namespace. When creating a function, keep in mind the configured resource quotas.

The yaml file shows a sample M size function.

The function includes the container **test** with the `spec.deployment.spec.containers[0].resources` configuration, as well as the configuration for the **horizontalPodAutoscaler**.

```yaml
apiVersion: kubeless.io/v1beta1
kind: Function
metadata:
  annotations:
    function-size: M
  finalizers:
  - kubeless.io/function
  generation: 1
  name: test
  namespace: stage
spec:
  checksum: ""
  deployment:
    spec:
      replicas: 2
      template:
        spec:
          containers:
          - name: test
            resources:
              limits:
                cpu: 500m
                memory: 256Mi
              requests:
                cpu: 100m
                memory: 100Mi
  deps: ""
  function: |-
    module.exports = { main: function (event, context) {
        return "test";
    } }
  function-content-type: text
  handler: handler.main
  horizontalPodAutoscaler:
    metadata:
      labels:
        function: test
      name: test
      namespace: stage
    spec:
      maxReplicas: 5
      metrics:
      - resource:
          name: cpu
          targetAverageUtilization: 40
        type: Resource
      minReplicas: 2
      scaleTargetRef:
        apiVersion: apps/v1beta1
        kind: Deployment
        name: test
  runtime: nodejs8
  service:
    ports:
    - name: http-function-port
      port: 8080
      protocol: TCP
      targetPort: 8080
    selector:
      created-by: kubeless
      function: test
  timeout: ""
```
