---
title: Expose a service
type: Tutorials
---

This tutorial shows how to expose service endpoints and configure different allowed HTTP methods for them using API Gateway Controller.

The tutorial comes with a sample HttpBin service deployment.

## Deploy and expose a service

Follow the instruction to deploy an unsecured instance of the HttpBin service and expose it.

1. Export this value as an environment variable:

  ```bash
  export DOMAIN={CLUSTER_DOMAIN}
  ```

2. Deploy an instance of the HttpBin service:

  ```bash
  kubectl apply -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml
  ```

3. Expose the service by creating an APIRule CR:

  ```bash
  cat <<EOF | kubectl apply -f -
  apiVersion: gateway.kyma-project.io/v1alpha1
  kind: APIRule
  metadata:
    name: httpbin
  spec:
    service:
      host: httpbin.$DOMAIN
      name: httpbin
      port: 8000
    gateway: kyma-gateway.kyma-system.svc.cluster.local
    rules:
      - path: /.*
        methods: ["GET"]
        accessStrategies:
          - handler: noop
        mutators:
          - handler: noop
      - path: /post
        methods: ["POST"]
        accessStrategies:
          - handler: noop
        mutators:
          - handler: noop
  EOF
  ```

  >**NOTE:** If you are running Kyma on Minikube, add `httpbin.kyma.local` to the entry with Minikube IP in your system's `/etc/hosts` file.

## Access the exposed resources

1. Call the endpoint by sending a `GET` request to the HttpBin service:

  ```bash
  curl -ik -X GET https://httpbin.$DOMAIN/ip
  ```

2. Send a `POST` request to the HttpBin's `/post` endpoint:

  ```bash
  curl -ik -X POST https://httpbin.$DOMAIN/post -d "test data"
  ```

These calls return the code `200` response.
