# Sample Functions

Functions support multiple languages through the use of runtimes. To use a chosen runtime, add its name and version as a value in the **spec.runtime** field of the [Function custom resource (CR)](../resources/06-10-function-cr.md). If this value is not specified, it defaults to `nodejs20`. Dependencies for a Node.js Function must be specified using the [`package.json`](https://docs.npmjs.com/creating-a-package-json-file) file format. Dependencies for a Python Function must follow the format used by [pip](https://packaging.python.org/key_projects/#pip).

> [!TIP]
> If you are interested in the Function's signature, `event` and `context` objects, and custom HTTP responses the Function returns, read about [Function’s specification](07-70-function-specification.md).

See sample Functions for all available runtimes:

<!-- tabs:start -->

#### Node.js

Serverless supports both CommonJS (cjs) and ECMAScript (ESM) syntax supported by Node.js.
You can switch between them using the `type` property in the Function dependencies.

CommonJS example:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: test-function-nodejs-cjs
spec:
  runtime: nodejs20
  source:
    inline:
      dependencies: |
        {
          "name": "test-function-nodejs",
          "version": "1.0.0",
          "dependencies": {
            "lodash":"^4.17.20"
          }
        }
      source: |
        const _ = require('lodash')
        module.exports = {
          main: function(event, context) {
            return _.kebabCase('Hello World from Node.js 20 Function');
          }
        }
EOF
```

ECMAScript example:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: test-function-nodejs-esm
spec:
  runtime: nodejs20
  source:
    inline:
      dependencies: |
        {
          "name": "test-function-nodejs-esm",
          "version": "1.0.0",
          "type": "module",
          "dependencies": {
            "lodash":"^4.17.20"
          }
        }
      source: |
        import _ from 'lodash'
        export function main (event, context) {
            return _.kebabCase('Hello World from Node.js 20 Function');
        }
EOF
```

#### Python

```bash
cat <<EOF | kubectl apply -f -
apiVersion: serverless.kyma-project.io/v1alpha2
kind: Function
metadata:
  name: test-function-python312
spec:
  runtime: python312
  source:
    inline:
      dependencies: |
        requests==2.31.0
      source: |
        import requests
        def main(event, context):
            r = requests.get('https://swapi.dev/api/people/13')
            return r.json()
EOF
```

<!-- tabs:end -->
