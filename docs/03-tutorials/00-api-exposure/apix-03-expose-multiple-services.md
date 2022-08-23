---
title: Expose multiple services on the same host
---

This tutorial shows how to expose a workload with multiple services that share the same host on different paths.

The tutorial may be a follow-up to the [Use a custom domain to expose a workload](./apix-01-own-domain.md) tutorial.

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create those workloads, follow the [Create a workload](./apix-02-create-workload.md) tutorial.


## Multiple services definition on different paths

Follow the instruction to expose your unsecured instance of the HttpBin service and your unsecured sample Function at the same time on different paths.

1. Export the following value as an environment variable:

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME} # This is a Kyma domain or your custom subdomain e.g. api.mydomain.com.
   ```

2. Expose the instance of the HttpBin service and the instance of the sample Function by creating an API Rule CR in your Namespace. If you don't want to use Kyma default gateway (kyma-system/kyma-gateway), replace it with your custom gateway. Run:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: multiple-service
  namespace: $NAMESPACE
  labels:
    app: multiple-service
    example: multiple-service
spec:
  host: multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS
  gateway: kyma-system/kyma-gateway
  rules:
    - path: /headers
      methods: ["GET"]
      accessStrategies:
        - handler: noop
      service:
        name: httpbin
        port: 8000
    - path: /function
      methods: ["GET"]
      accessStrategies:
        - handler: noop
      service:
        name: function
        port: 8
EOF
```


3. Call the endpoints by sending `GET` requests to the HttpBin service and the sample function:

   ```bash
   curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/headers  # Send a GET request to the HttpBin
   curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function  # Send a GET request to the Function

   ```
These calls return the code 200 response.



## Root level service definition and multiple services definition on different paths

Follow the instruction to expose your unsecured instance of the HttpBin service and your sample Function on different paths with a service defined at the root level - HttpBin in the following example. 
  >**NOTE:** The services definition at the **spec.rules** level have higher precedence than the service definition at the **spec.service** level.


1. Export the following value as an environment variable:

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME} # This is a Kyma domain or your custom subdomain e.g. api.mydomain.com
   ```

2. Expose the instance of the HttpBin service and the instance of the sample Function by creating an API Rule CR in your Namespace. If you don't want to use Kyma default gateway (kyma-system/kyma-gateway), replace it with your custom gateway. In the following example, the services definition at the **spec.rules** level overwrites the service definition at the **spec.service** level. Run:

```yaml
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: multiple-service
  namespace: $NAMESPACE
  labels:
    app: multiple-service
    example: multiple-service
spec:
  host: multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS
  gateway: kyma-system/kyma-gateway
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /headers
      methods: ["GET"]
      accessStrategies:
        - handler: noop
    - path: /function
      methods: ["GET"]
      accessStrategies:
        - handler: noop
      service:
        name: function
        port: 80
EOF
```

3. Call the endpoints by sending `GET` requests to the HttpBin service and the sample function:

   ```bash
   curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/headers  # Send a GET request to the HttpBin
   curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function  # Send a GET request to the Function

   ```
These calls return the code 200 response.