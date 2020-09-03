---
title: Runtimes
---

Functions support multiple languages through the use of runtimes. In order to use chosen runtime supply the value in `spec.runtime` field of Function CR. If it's not specified, it defaults to `nodejs12`. Node.js function dependencies should be specified using [package.json](https://docs.npmjs.com/creating-a-package-json-file) format. Python Function dependencies should follow requirements format, used by [pip](https://packaging.python.org/key_projects/#pip).

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
    module.exports = {
      main: function(event, context) {
        return 'Hello World!'
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
    import pydash

    def main(event, context):
        return pydash.omit({'name': 'moe', 'age': 40}, 'age')
  deps: |
    pydash==4.8.0
EOF
```

</details>
</div>
