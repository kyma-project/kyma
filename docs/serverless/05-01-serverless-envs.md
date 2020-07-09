---
title: Environmental variables
type: Configuration
---

To configure the Serverless Function, override the default values of enviromental variables:

| Environmental variable | Description                                                                   | Unit   | Default value |
| ---------------------- | ----------------------------------------------------------------------------- | ------ | ------------- |
| **FUNC_TIMEOUT**       | Specifies the number of seconds to execute code before terminating execution. | Number | 180           |
| **REQ_MB_LIMIT**       | Payload body size limit in megabytes                                          | Number | 1             |

See [this file](https://github.com/kubeless/runtimes/blob/master/stable/nodejs/kubeless.js) to get deeper understanding of how exactly those values are used. That express app is internally used to run Functions.

See the example of a function with such environmental variables set:

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
