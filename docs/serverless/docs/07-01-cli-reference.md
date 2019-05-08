---
title: CLI reference
---

This section provides you with helpful command line examples used in Kyma.

## Prerequisites

* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.10.0

To develop, deploy, or run functions directly, download these tools additionally:

* [Kubeless CLI](https://github.com/kubeless/kubeless/releases)
* [Node.js, version 6 or 8](https://nodejs.org/en/download/)

### Set the cluster domain variable

The commands throughout this guide use URLs that require you to provide the domain of the cluster which you are using. To complete this configuration, set the variable `yourClusterDomain` to the domain of your cluster.

For example, if your cluster's domain is `demo.cluster.kyma.cx`, run the following command:

```bash
export yourClusterDomain='demo.cluster.kyma.cx'
```

## Details

Use the command line to create, call, deploy, expose, and bind a function.

### Deploy a function

#### Using a yaml file and kubectl

You can use the Kubeless CLI to deploy functions in Kyma.

```bash
kubectl apply -f https://raw.githubusercontent.com/kyma-project/examples/master/gateway/lambda/deployment.yaml
```

#### Using a JS file and the Kubeless CLI

You can deploy a function using the Kubernetes and Kubeless CLI. See the following example:

```bash
kubeless function deploy hello --runtime nodejs8 --handler hello.main --from-file https://raw.githubusercontent.com/kyma-project/examples/master/event-subscription/lambda/js/hello-with-data.js
```

### List functions

Check if the function is available:

```bash
kubeless function list hello
```

### Expose a function

#### Without authentication

Use the CLI to create an API for your function:

```bash
kubectl apply -f https://raw.githubusercontent.com/kyma-project/examples/master/gateway/lambda/api-without-auth.yaml
```

#### With authentication enabled

If your function is deployed to a cluster, run:

```bash
curl https://raw.githubusercontent.com/kyma-project/examples/master/gateway/lambda/api-with-auth.yaml | sed "s/.kyma.local/.$yourClusterDomain/" | kubectl apply -f -
```

### Call the function using curl

If Kyma is running locally, add `hello.kyma.local` mapped to `minikube ip` to `/etc/hosts`.

```bash
echo "$(minikube ip) hello.kyma.local" | sudo tee -a /etc/hosts
```

Use curl to call the function:

```bash
curl -L -k hello.kyma.local
```

You should receive the following response:

```
hello world
```

### Bind a function to events
You can bind the function to Kyma and to third-party services. For details, refer to the Service Catalog-related documentation.
