---
title: CLI reference
type: CLI reference
---

## Overview

This section provides you with useful command line examples used in Kyma.

## Prerequisites

* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.10.0

To develop, deploy, or run functions directly download these tools additionally:

* [Kubeless CLI](https://github.com/kubeless/kubeless/releases)
* [Node.js, version 6 or 8](https://nodejs.org/en/download/)

### Set the cluster domain variable

The commands throughout this guide use URLs that require you to provide the domain of the cluster which you are using. To complete this configuration, set the variable `yourClusterDomain` to the domain of your cluster.

For example if your cluster's domain is 'demo.cluster.kyma.cx' then run the following command:

   ```bash
   export yourClusterDomain='demo.cluster.kyma.cx'
   ```

## Details

Use the command line to create, call, deploy, expose, and bind a function.

### Deploy a function using a yaml file and kubectl

You can use the Kubeless CLI to deploy functions in Kyma.

```bash
$ kubectl apply -f https://minio.$yourClusterDomain/content/components/serverless/assets/deployment.yaml
```

Check if the function is available:
```bash
$ kubeless function list hello
```
### Deploy a function using a JS file and the Kubeless CLI

You can deploy a function using the Kubernetes and Kubeless CLI. See the following example:

```bash
$ kubeless function deploy hello --runtime nodejs8 --handler hello.main --from-file https://minio.$yourClusterDomain/content/components/serverless/assets/hello.js --trigger-http
```

### Call a function using the CLI

Use the CLI to call a function:

```bash
$ kubeless function call hello
```

### Expose a function without authentication

Use the CLI to create an API for your function:

```bash
$ kubectl apply -f https://minio.$yourClusterDomain/content/components/serverless/assets/api-without-auth.yaml
```

### Expose a function with authentication enabled

If your function is deployed to a cluster run:

```bash
 curl -k https://minio.$yourClusterDomain/content/components/serverless/assets/api-with-auth.yaml | sed "s/.kyma.local/.$yourClusterDomain/" | kubectl apply -f -
```


If Kyma is running locally, add `hello.kyma.local` mapped to `minikube ip` to `/etc/hosts`

```bash
$ echo "$(minikube ip) hello.kyma.local" | sudo tee -a /etc/hosts
```

Create the API for your function:

```bash
kubectl apply -f https://minio.$yourClusterDomain/content/components/serverless/assets/api-with-auth.yaml
```

### Bind a function to events
You can bind the function to Kyma and to third-party services. For details, refer to the [Service Catalog](../../service-catalog/docs/001-overview-service-catalog.md) documentation.
