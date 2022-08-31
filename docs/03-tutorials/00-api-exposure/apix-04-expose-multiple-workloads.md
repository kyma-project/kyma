---
title: Expose multiple workloads on the same host
---

This tutorial shows how to expose multiple workloads that share the same host on different paths.

You can either define a service at the root level, which is applied to all paths except the ones you've explicitly set service for at the rules level, or you can just define different services on each path separately without the need to define a root service.
   > **CAUTION:** Exposing a workload to the outside world is always a potential security vulnerability, so tread carefully. In a production environment, always secure the workload you expose with [OAuth2](./apix-05-expose-and-secure-workload-oauth2.md) or [JWT](./apix-08-expose-and-secure-workload-jwt.md).

The tutorial may be a follow-up to the [Set up a custom domain for a workload](./apix-02-setup-custom-domain-for-workload.md) tutorial.

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create those workloads, follow the [Create a workload](./apix-01-create-workload.md) tutorial.

## Root level service definition and multiple services definition on different paths

Follow the instructions to expose your instance of the HttpBin service and your sample Function on different paths with a service defined at the root level - HttpBin in the following example. 
  >**NOTE:** The service definitions at the **spec.rules** level have higher precedence than the service definition at the **spec.service** level.


1. Export the following value as an environment variable:

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME} # This is a Kyma domain or your custom subdomain, for example, api.mydomain.com
   ```

2. To expose the instance of the HttpBin service and the instance of the sample Function, create an API Rule CR in your Namespace. If you don't want to use Kyma's default gateway, replace `kyma-system/kyma-gateway` with your custom gateway.
In the following example, the services definition at the **spec.rules** level overwrites the service definition at the **spec.service** level. Run:

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

3. To call the endpoints, send `GET` requests to the HttpBin service and the sample Function:

   ```bash
   curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/headers  # Send a GET request to the HttpBin
   curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function  # Send a GET request to the Function

   ```
   These calls return the code 200 response.

## Multiple services definition on different paths

Follow the instruction to expose your instance of the HttpBin service and your sample Function at the same time on different paths.

1. Export the following value as an environment variable:

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME} # This is a Kyma domain or your custom subdomain, for example, api.mydomain.com
   ```

2. To expose the instance of the HttpBin service and the instance of the sample Function, create an API Rule CR in your Namespace. If you don't want to use Kyma's default gateway, replace `kyma-system/kyma-gateway` with your custom gateway. Run:

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
           port: 80
   EOF
   ```


3. To call the endpoints, send `GET` requests to the HttpBin service and the sample Function:

   ```bash
   curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/headers  # Send a GET request to the HttpBin
   curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function  # Send a GET request to the Function

   ```
   These calls return the code 200 response.

