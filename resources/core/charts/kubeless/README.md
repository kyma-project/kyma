
```
  _  __     _          _
 | |/ /    | |        | |
 | ' /_   _| |__   ___| | ___  ___ ___
 |  <| | | | '_ \ / _ \ |/ _ \/ __/ __|
 | . \ |_| | |_) |  __/ |  __/\__ \__ \
 |_|\_\__,_|_.__/ \___|_|\___||___/___/

```

## Overview

This document explains how Kyma installs Kubeless in the `kyma-system` namespace, and how the Kubeless CLI works with functions.

The installation of Kubeless in the `kyma-system` namespace enables these components:

* CustomResourceDefinition
* Controller
* ClusterRole with ClusterRoleBinding (valid for a global installation in a cluster)
* Namespace
* ServiceAccount
* Lambdas UI

## Prerequisites

The Kubeless CLI is already installed inside the Kubeless CLI container. It is required only outside of the CLI container. For more details on how to install the Kubeless CLI, go to [this](https://github.com/kubeless/kubeless#installation) URL.

## Details

This section explains how to manually create, deploy, call, and delete functions from the Kubeless CLI which is already installed inside the Kubeless Container.

### Create a function in node.js

Create a sample function in `node.js`. Save the function in a file.


```bash
$ echo "module.exports = { \
  foo: function (req, res) { \
        res.end('hello world') \
  } \
}" > hello.js
```

### Deploy a function using Kubeless CLI

For `node.js`, run this command:

```bash
$ kubeless function deploy testjs --runtime nodejs8 --handler hello.foo --from-file hello.js --trigger-http
```

### Call a function using CLI

To call a function using CLI, run this command:

```bash
$ kubeless function call <function-name>
E.g
$ kubeless function call testjs
```

### Delete a function

To delete a function, run this command:

```bash
$ kubeless function delete <function-name>
E.g
$ kubeless function delete testjs
```

For more examples, see [this](https://github.com/kyma-project/examples/tree/master/serverless-lambda) document.

## Troubleshooting

To debug function pods, run this command:

```bash
$ kubectl logs <function-pod-name>
```

To debug the `kube-controller`, run this command:

```bash
$ kubectl logs <kube-controller-pod-name> -n kyma-system
```

## References

* [Kubeless Installation](https://github.com/kubeless/kubeless#installation)
* [Expose and secure Kubeless functions](https://github.com/kubeless/kubeless/blob/master/docs/http-triggers.md#expose-and-secure-kubeless-functions)