---
title: Sample service deployment on a cluster
type: Tutorials
---

This tutorial is intended for the developers who want to quickly learn how to deploy a sample service and test it with the Kyma cluster.

This tutorial uses a standalone sample service written in the [Go](http://golang.org) language.

## Prerequisites

To use the Kyma cluster and install the example, download these tools:

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.10.0
- [curl](https://github.com/curl/curl)

## Steps

### Get the kubeconfig file and configure the CLI

Follow these steps to get the `kubeconfig` file and configure the CLI to connect to the cluster:

1. Access the Console UI of your Kyma cluster.
2. Click **Administration**.
3. Click the **Download config** button to download the `kubeconfig` file to a selected location on your machine.
4. Open a terminal window.
5. Export the **KUBECONFIG** environment variable to point to the downloaded `kubeconfig`. Run this command:
  ```
  export KUBECONFIG={KUBECONFIG_FILE_PATH}
  ```
  >**NOTE:** Drag and drop the `kubeconfig` file in the terminal to easily add the path of the file to the `export KUBECONFIG` command you run.

6. Run `kubectl cluster-info` to check if the CLI is connected to the correct cluster.

### Set the cluster domain as an environment variable

The commands in this guide use URLs in which you must provide the domain of the cluster that you use.
Export the domain of your cluster as an environment variable. Run:  
  ```
  export yourClusterDomain='{YOUR_CLUSTER_DOMAIN}'
  ```

### Deploy and expose a sample standalone service

Follow these steps:

1. Deploy the sample service to any of your Namespaces. Use the `stage` Namespace for this guide:

   ```bash
   kubectl create -n stage -f https://raw.githubusercontent.com/kyma-project/examples/master/http-db-service/deployment/deployment.yaml
   ```

2. Create an unsecured API for your service:

   ```bash
   curl -k https://raw.githubusercontent.com/kyma-project/examples/master/gateway/service/api-without-auth.yaml |  sed "s/.kyma.local/.$yourClusterDomain/" | kubectl apply -n stage -f -
   ```

3. Access the service using the following call:
   ```bash
   curl -ik https://http-db-service.$yourClusterDomain/orders
   ```

   The system returns a response similar to the following:
   ```
   HTTP/2 200
   content-type: application/json;charset=UTF-8
   vary: Origin
   date: Mon, 01 Jun 2018 00:00:00 GMT
   content-length: 2
   x-envoy-upstream-service-time: 131
   server: envoy

   []
   ```

### Update your service's API to secure it

Run the following command:

   ```bash
   curl -k https://raw.githubusercontent.com/kyma-project/examples/master/gateway/service/api-with-auth.yaml |  sed "s/.kyma.local/.$yourClusterDomain/" | kubectl apply -n stage -f -
   ```
After you apply this update, you must include a valid bearer ID token in the Authorization header to access the service.

>**NOTE:** The update might take some time.
