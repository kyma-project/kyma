---
title: Expose a service
type: Tutorials
---

This tutorial shows how to expose services using the API Gateway Controller. The controller reacts to an instance of the APIRule custom resource (CR) and creates an Istio Virtual Service according to the details specified in the CR.

The tutorial comes with a sample HttpBin service deployment.

## Deploy and expose a service

Follow the instruction to deploy an instance of the HttpBin service and expose it.

1. Export this value as an environment variable:

```bash
export DOMAIN={CLUSTER_DOMAIN}
```


2. Deploy an instance of the HttpBin service:

  ```bash
  kubectl apply -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml
  ```

2. Expose the service by creating an APIRule CR:

```bash
  cat <<EOF | kubectl apply -f -
  apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: httpbin-unsecured
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
  EOF
  ```

>**NOTE:** If you are running Kyma on Minikube, add `httpbin.kyma.local` to the entry with Minikube IP in your system's `/etc/hosts` file.