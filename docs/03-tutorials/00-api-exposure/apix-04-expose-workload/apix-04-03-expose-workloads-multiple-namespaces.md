---
title: Expose workloads in multiple Namespaces with a single APIRule definition
---

This tutorial shows how to expose service endpoints in multiple Namespaces using API Gateway Controller.
   > **CAUTION:** Exposing a workload to the outside world causes a potential security vulnerability, so tread carefully. In a production environment, secure the workload you expose with [OAuth2](../apix-05-expose-and-secure-a-workload/apix-05-01-expose-and-secure-workload-oauth2.md) or [JWT](../apix-05-expose-and-secure-a-workload/apix-05-03-expose-and-secure-workload-jwt.md).


##  Prerequisites

1. Create a sample HttpBin service deployment and a sample Function in the separate Namespaces:

   * Create a Namespace for the HttpBin service and export its value as an environment variable. Run:

     ```bash
     export NAMESPACE_HTTPBIN={NAMESPACE_HTTPBIN}
     kubectl create ns $NAMESPACE_HTTPBIN
     kubectl label namespace $NAMESPACE_HTTPBIN istio-injection=enabled --overwrite
     ```

   * Create a different Namespace for the Function service and export its value as an environment variable. Run:

     ```bash
     export NAMESPACE_FUNCTION={NAMESPACE_FUNCTION}
     kubectl create ns $NAMESPACE_FUNCTION
     kubectl label namespace $NAMESPACE_FUNCTION istio-injection=enabled --overwrite
     ```

   * Deploy an instance of the HttpBin service in its Namespace using the [sample code](https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml):

     ```bash
     kubectl -n $NAMESPACE_HTTPBIN create -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml
     ```

   * Create a Function in its Namespace using the [sample code](../assets/function.yaml):

     ```bash
     kubectl -n $NAMESPACE_FUNCTION apply -f https://raw.githubusercontent.com/kyma-project/kyma/main/docs/03-tutorials/assets/function.yaml
     ```

2. Create a Namespace for the Gateway and APIRule CRs and export its value as the environment variable. Run:

  ```bash
  export NAMESPACE={NAMESPACE}
  kubectl create ns $NAMESPACE
  kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
  ```

3. Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:
  
<div tabs name="export-values">

  <details>
  <summary>
  Custom domain
  </summary>
    
  ```bash
  export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
  export GATEWAY=$NAMESPACE/httpbin-gateway
  ```
  </details>

  <details>
  <summary>
  Kyma domain
  </summary>

  ```bash
  export DOMAIN_TO_EXPOSE_WORKLOADS={KYMA_DOMAIN_NAME}
  export GATEWAY=kyma-system/kyma-gateway
  ```
  </details>
</div> 

## Expose and access your workloads in multiple Namespaces

1. Expose the HttpBin and Function services in their respective Namespaces by creating an APIRule CR in its own Namespace. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: gateway.kyma-project.io/v1beta1
   kind: APIRule
   metadata:
     name: httpbin-and-function
     namespace: $NAMESPACE
   spec:
     host: httpbin-and-function.$DOMAIN_TO_EXPOSE_WORKLOADS
     gateway: $GATEWAY
     rules:
       - path: /headers
         methods: ["GET"]
         service:
           name: httpbin
           namespace: $NAMESPACE_HTTPBIN
           port: 8000
         accessStrategies:
           - handler: noop
         mutators:
           - handler: noop
       - path: /function
         methods: ["GET"]
         service:
           name: function
           namespace: $NAMESPACE_FUNCTION
           port: 80
         accessStrategies:
           - handler: noop
         mutators:
           - handler: noop
   EOF
   ```

   >**NOTE:** If you are running Kyma on k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

2. Call the HttpBin endpoint by sending a `GET` request to the HttpBin service:

   ```bash
   curl -ik -X GET https://httpbin-and-function.$DOMAIN_TO_EXPOSE_WORKLOADS/headers
   ```

  If successful, the call returns the code `200 OK` response.

3. Call the Function endpoint by sending a `GET` request to the Function service:

   ```bash
   curl -ik -X GET https://httpbin-and-function.$DOMAIN_TO_EXPOSE_WORKLOADS/function
   ```
  If successful, the call returns the code `200 OK` response.