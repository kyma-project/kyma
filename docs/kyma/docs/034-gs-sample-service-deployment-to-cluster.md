---
title: Sample service deployment on a cluster
type: Getting Started
---

## Overview

This Getting Started guide is intended for the developers who want to quickly learn how to deploy a sample service and test it with the Kyma cluster.

This guide uses a standalone sample service written in the [Go](http://golang.org) language.

## Prerequisites

To use the Kyma cluster and install the example, download these tools:

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.10.0
- [curl](https://github.com/curl/curl)

## Steps

### Download configuration for kubectl

Follow these steps to download **kubeconfig** and configure kubectl to access the Kyma cluster:
1. Access the Console UI and download the **kubectl** file from the settings page.
2. Place downloaded file in the following location: `$HOME/.kube/kubeconfig`.
3. Point **kubectl** to the configuration file using the terminal: `export KUBECONFIG=$HOME/.kube/kubeconfig`.
4. Confirm **kubectl** is configured to use your cluster: `kubectl cluster-info`.

### Set the cluster domain variable

The commands throughout this guide use URLs that require you to provide the domain of the cluster which you are using. To complete this configuration, set the variable `yourClusterDomain` to the domain of your cluster.

For example if your cluster's domain is 'demo.cluster.kyma.cx' then run the following command:

   ```bash
   export yourClusterDomain='demo.cluster.kyma.cx'
   ```

### Deploy and expose a sample standalone service

Follow these steps:

1. Deploy the sample service to any of your Environments. Use the `stage` Environment for this guide:

   ```bash
   kubectl create -n stage -f https://minio.$yourClusterDomain/content/root/kyma/assets/deployment.yaml
   ```

2. Create an unsecured API for your service:

   ```bash
   curl -k https://minio.$yourClusterDomain/content/root/kyma/assets/api-without-auth.yaml |  sed "s/.kyma.local/.$yourClusterDomain/" | kubectl apply -n stage -f -
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
   curl -k https://minio.$yourClusterDomain/content/root/kyma/assets/api-with-auth.yaml |  sed "s/.kyma.local/.$yourClusterDomain/" | kubectl apply -n stage -f -
   ```
After you apply this update, you must include a valid bearer ID token in the Authorization header to access the service.

>**NOTE:** The update might take some time.
