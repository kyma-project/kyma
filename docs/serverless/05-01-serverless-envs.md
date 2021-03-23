---
title: Environment variables
type: Configuration
---

You can configure environment variables either separately for a given runtime or make them runtime-agnostic using a ConfigMap.

## Define environment variables in a ConfigMap

ConfigMaps allow you to define Function's environment variables for any runtime through key-value pairs. After you define the values in a ConfigMap, simply reference it in the Function custom resource (CR) through the **valueFrom** parameter. See an example of such a Function CR that specifies the `my-var` value as a reference to the key stored in the `my-vars-cm` ConfigMap as the `MY_VAR` environment variable.

```yaml
apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: sample-cm-env-values
  namespace: default
spec:
  env:
    - name: MY_VAR
      valueFrom:
        configMapKeyRef:
          name: my-vars-cm
          key: my-var
  source: |
    module.exports = {
      main: function (event, context) {
        return process.env["MY_VAR"];
      }
    }
```

## NodeJS runtime-specific environment variables

To configure the Function with the Node.js runtime, override the default values of these environment variables:

| Environment variable             | Description                                                                                                                  | Type    | Default value |
| -------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | ------- | ------------- |
| **FUNC_TIMEOUT**                 | Specifies the number of seconds in which a runtime must execute the code.                                                    | Number  | `180`         |
| **REQ_MB_LIMIT**                 | Specifies the payload body size limit in megabytes.                                                                          | Number  | `1`           |
| **KYMA_INTERNAL_LOGGER_ENABLED** | Enables the default HTTP request logger which uses the standard Apache combined log output. To enable it, set its value to `true`.  | Boolean | `false`       |

See [`kubeless.js`](https://github.com/kubeless/runtimes/blob/master/stable/nodejs/kubeless.js) to get a deeper understanding of how the Express server, that acts as a runtime, uses these values internally to run Node.js Functions.

See the example of a Function with these environment variables set:

```yaml
apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: sample-fn-with-envs
  namespace: default
spec:
  env:
    - name: FUNC_TIMEOUT
      value: "2"
    - name: REQ_MB_LIMIT
      value: "10"
  source: |
    module.exports = {
      main: function (event, context) {
        return "Hello World!";
      }
    }
```

## Python runtime-specific environment variables

To configure a Function with the Python runtime, override the default values of these environment variables:

| Environment variable             | Description                                                                                                                  | Unit    | Default value   |
| -------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | ------- | --------------- | ------------------------------------------------------------------------------ |
| **FUNC_MEMFILE_MAX**             | Specifies the maximum size of the memory buffer for the HTTP request body in bytes.                                                            | Number  | `100*1024*1024` | <!-- https://bottlepy.org/docs/dev/api.html#bottle.BaseRequest.MEMFILE_MAX --> |
| **CHERRYPY_NUMTHREADS**          | Specifies the number of requests that can be handled in parallel                                                                           | Number  | 10              |
| **KYMA_INTERNAL_LOGGER_ENABLED** | Enables the default HTTP request logger which uses the standard Apache combined log output. To enable it, set its value to `true`. | Boolean | `false`         |

See [`kubeless.py`](https://github.com/kubeless/runtimes/blob/master/stable/python/_kubeless.py) to get a deeper understanding of how the Bottle server, that acts as a runtime, uses these values internally to run Python Functions.

```yaml
apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: sample-fn-with-envs
  namespace: default
spec:
  env:
    - name: FUNC_MEMFILE_MAX
      value: "1048576" # 1MiB
  source: |
    def main(event. context):
      return "Hello World!"
```
