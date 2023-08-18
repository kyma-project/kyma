---
title: Expose workloads in multiple Namespaces with a single APIRule definition
---

This tutorial shows how to expose service endpoints in multiple Namespaces using API Gateway Controller.
   > **CAUTION:** Exposing a workload to the outside world causes a potential security vulnerability, so tread carefully. In a production environment, secure the workload you expose with [OAuth2](../apix-05-expose-and-secure-a-workload/apix-05-01-expose-and-secure-workload-oauth2.md) or [JWT](../apix-05-expose-and-secure-a-workload/apix-05-03-expose-and-secure-workload-jwt.md).


##  Prerequisites

1. Create three Namespaces: one for an instance of the HttpBin service, one for a sample Function, and one for an APIRule custom resource (CR). Deploy an instance of the HttpBin service and a sample Function in their respective Namespaces. To learn how to do it, follow the [Create a workload](../apix-01-create-workload.md) tutorial. 

  >**NOTE:** Remember to [enable the Istio sidecar proxy injection](../../../04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection.md) in each Namespace.

2. Export the Namespaces' names as environment variables:

  ```bash
  export NAMESPACE_HTTPBIN={NAMESPACE_NAME}
  export NAMESPACE_FUNCTION={NAMESPACE_NAME}
  export NAMESPACE_APIRULE={NAMESPACE_NAME}
  ```
  
3. Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:
  
  <div tabs name="export-values">

    <details>
    <summary>
    Custom domain
    </summary>
      
    ```bash
    export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
    export GATEWAY=$NAMESPACE_APIRULE/httpbin-gateway
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
     namespace: $NAMESPACE_APIRULE
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