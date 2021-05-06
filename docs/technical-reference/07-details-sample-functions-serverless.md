---
title: Sample Functions
type: Details
---

See sample Functions for all available runtimes:

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
        return fetch("https://swapi.dev/api/people/1").then(res => res.json())
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
  <summary label="nodejs14">
  Node.js 14
  </summary>

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: serverless.kyma-project.io/v1alpha1
kind: Function
metadata:
  name: test-function-nodejs14
spec:
  runtime: nodejs14
  source: |
    const _ = require('lodash')

    module.exports = {
      main: function(event, context) {
        return _.kebabCase('Hello World from Node.js 14 Function');
      }
    }
  deps: |
    {
      "name": "test-function-nodejs14",
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
        r = requests.get('https://swapi.dev/api/people/13')
        return r.json()
  deps: |
    requests==2.24.0
EOF
```

</details>
</div>
