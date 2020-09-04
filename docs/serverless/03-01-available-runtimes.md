---
title: Runtimes
---

Functions support multiple languages through the use of runtimes. In order to use chosen runtime supply the value in `spec.runtime` field of [Function CR](#custom-resource-function). If it's not specified, it defaults to `nodejs12`. Node.js function dependencies should be specified using [package.json](https://docs.npmjs.com/creating-a-package-json-file) format. Python Function dependencies should follow requirements format, used by [pip](https://packaging.python.org/key_projects/#pip).

See sample Functions for each available runtime:

> **CAUTION:** When you create a Function, the exported object in the Function's body must have `main` as the handler name.

<div tabs name="available-runtimes" group="available-runtimes">
  <details>
  <summary label="nodejs12">
  Node.js 12
  </summary>

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: test-function-nodejs12
spec:
  runtime: nodejs12
  source: |
    const fetch = require("node-fetch");

    module.exports = {
      main: function (event, context) {
        return fetch("https://swapi.dev/api/people/12").then(res => res.json())
      }
    }
  deps: |
    {
      "name": "test-function-nodejs12",
      "version": "1.0.0",
      "dependencies": {
        "node-fetch": "^2.6.0"
      }
    }
EOF
```

  </details>
  <details>
  <summary label="nodejs10">
  Node.js 10
  </summary>

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: test-function-nodejs10
spec:
  runtime: nodejs10
  source: |
    const _ = require('lodash')

    module.exports = {
      main: function(event, context) {
        return _.kebabCase('Hello World from Node.js 10 Function');
      }
    }
  deps: |
    {
      "name": "test-function-nodejs10",
      "version": "1.0.0",
      "dependencies": {
        "lodash":"^4.17.20"
      }
    }
EOF
```

  </details>
  <details>
  <summary label="python38">
  Python 3.8
  </summary>

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: test-function-python38
spec:
  runtime: python38
  source: |
    import requests

    def main(event, context):
        params = {'format': 'wookiee'}
        r = requests.get('https://swapi.dev/api/people/13', params=params)
        return r.json()
  deps: |
    certifi==2020.6.20
    chardet==3.0.4
    idna==2.10
    requests==2.24.0
    urllib3==1.25.10
EOF
```

</details>
</div>
