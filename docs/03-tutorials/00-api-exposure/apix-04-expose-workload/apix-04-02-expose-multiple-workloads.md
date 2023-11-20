---
title: Expose multiple workloads on the same host
---

This tutorial shows how to expose multiple workloads on different paths by defining a service at the root level and by defining services on each path separately.

   > **CAUTION:** Exposing a workload to the outside world is always a potential security vulnerability, so tread carefully. In a production environment, remember to secure the workload you expose with [OAuth2](../apix-05-expose-and-secure-a-workload/apix-05-01-expose-and-secure-workload-oauth2.md) or [JWT](../apix-05-expose-and-secure-a-workload/apix-05-03-expose-and-secure-workload-jwt.md).

## Prerequisites

* Deploy [a sample HttpBin service and a sample Function](../apix-01-create-workload.md).
* Set up [your custom domain](../apix-02-setup-custom-domain-for-workload.md) or use a Kyma domain instead. 
* Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:

<!-- tabs:start -->

#### **Custom domain**
    
    ```bash
    export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
    export GATEWAY=$NAMESPACE/httpbin-gateway
    ```

#### **Kyma domain**

    ```bash
    export DOMAIN_TO_EXPOSE_WORKLOADS={KYMA_DOMAIN_NAME}
    export GATEWAY=kyma-system/kyma-gateway
    ```

<!-- tabs:end -->   

## Define multiple services on different paths

Follow the instructions to expose the instance of the HttpBin service and the sample Function on different paths at the `spec.rules` level without a root service defined.

1. To expose the instance of the HttpBin service and the instance of the sample Function, create an APIRule CR in your Namespace. Run:

   ```bash
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
     gateway: $GATEWAY
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

2. To call the endpoints, send `GET` requests to the HttpBin service and the sample Function:

    ```bash
    curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/headers

    curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function 
    ```
  If successful, the calls return the code `200 OK` response.

## Define a service at the root level

A service can be also defined at the root level. Such a definition is applied to all the paths specified at the `spec.rules` which do not have their own services defined. 
 
 > **NOTE:** Services definitions at the `spec.rules` level have precedence over service definition at the `spec.service` level.

Follow the instructions to expose the instance of the HttpBin service and the sample Function on different paths with a service defined at the root level.

1. To expose the instance of the HttpBin service and the instance of the sample Function, create an APIRule CR in your Namespace. Run:

   ```bash
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
     gateway: $GATEWAY
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
  In the above APIRule, the HttpBin service on port 8000 is defined at the `spec.service` level. This service definition is applied to the `/headers` path. The `/function` path has the service definition overwritten.

2. To call the endpoints, send `GET` requests to the HttpBin service and the sample Function:

    ```bash
    curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/headers

    curl -ik -X GET https://multiple-service-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function 
    ```
  If successful, the calls return the code `200 OK` response.