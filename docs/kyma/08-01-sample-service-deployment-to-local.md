---
title: Sample service deployment on local
type: Tutorials
---

This tutorial is intended for the developers who want to quickly learn how to deploy a sample service and test it with Kyma installed locally on Mac.

This tutorial uses a standalone sample service written in the [Go](http://golang.org) language .

## Prerequisites

To use the Kyma cluster and install the example, download these tools:

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.10.0
- [curl](https://github.com/curl/curl)

## Steps

### Deploy and expose a sample standalone service

Follow these steps:

1. Deploy the sample service to any of your Namespaces. Use the `stage` Namespace for this guide:

   ```bash
   kubectl create -n stage -f https://raw.githubusercontent.com/kyma-project/examples/master/http-db-service/deployment/deployment.yaml
   ```

2. Create an unsecured API for your example service:

   ```bash
   kubectl apply -n stage -f https://raw.githubusercontent.com/kyma-project/examples/master/gateway/service/api-without-auth.yaml
   ```

3. Add the IP address of Minikube to the `hosts` file on your local machine for your APIs:

   ```bash
   $ echo "$(minikube ip) http-db-service.kyma.local" | sudo tee -a /etc/hosts
   ```

4. Access the service using the following call:
   ```bash
   curl -ik https://http-db-service.kyma.local/orders
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
   kubectl apply -n stage -f https://raw.githubusercontent.com/kyma-project/examples/master/gateway/service/api-with-auth.yaml
   ```
After you apply this update, you must include a valid bearer ID token in the Authorization header to access the service.

>**NOTE:** The update might take some time.
