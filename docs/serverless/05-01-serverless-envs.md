---
title: Environment variables
type: Configuration
---

To configure the Serverless Function, override the default values of these environment variables:

| Environment variable | Description                                                                   | Unit   | Default value |
| ---------------------- | ----------------------------------------------------------------------------- | ------ | ------------- |
| **FUNC_TIMEOUT**       | Specifies the number of seconds in which a runtime must execute the code. | Number | `180`           |
| **REQ_MB_LIMIT**       | Payload body size limit in megabytes                                          | Number |  `1`            |

See [`kubeless.js`](https://github.com/kubeless/runtimes/blob/master/stable/nodejs/kubeless.js) to get a deeper understanding of how the Express server, that acts as a runtime, uses these values internally to run Functions.

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
