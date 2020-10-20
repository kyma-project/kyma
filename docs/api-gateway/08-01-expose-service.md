---
title: Expose a service
type: Tutorials
---

This tutorial shows how to expose services using the API Gateway Controller for the `GET` method.

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
  name: httpbin-unsecured
spec:
  service:
    host: httpbin-un.$DOMAIN
    name: httpbin-unsecured
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

## Access the exposed resources

Follow the instruction to call the service.

1. Send a `GET` request to the HttpBin service:

```bash
curl -ik -X GET https://httpbinallow.$DOMAIN/headers
```
