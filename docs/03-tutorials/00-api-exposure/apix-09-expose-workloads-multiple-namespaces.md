---
title: Expose workloads in multiple Namespaces with a single APIRule definition
---

This tutorial shows how to expose service endpoints in multiple Namespaces using API Gateway Controller.
   > **CAUTION:** Exposing a workload to the outside world is always a potential security vulnerability, so tread carefully. In a production environment, always secure the workload you expose with [OAuth2](./apix-05-expose-and-secure-workload-oauth2.md) or [JWT](./apix-08-expose-and-secure-workload-jwt.md).


## Expose and access your workloads in multiple Namespaces

Follow the instructions to expose and access your unsecured instance of the HttpBin service and unsecured sample Function.

1. Create a Namespace for the HttpBin service and export its value as an environment variable. Run:

   ```bash
   export NAMESPACE_HTTPBIN={NAMESPACE_HTTPBIN}
   kubectl create ns $NAMESPACE_HTTPBIN
   kubectl label namespace $NAMESPACE_HTTPBIN istio-injection=enabled --overwrite
   ```

2. Create a different Namespace for the Function service and export its value as an environment variable. Run:

   ```bash
   export NAMESPACE_FUNCTION={NAMESPACE_FUNCTION}
   kubectl create ns $NAMESPACE_FUNCTION
   kubectl label namespace $NAMESPACE_FUNCTION istio-injection=enabled --overwrite
   ```

3. Deploy an instance of the HttpBin service in its Namespace:

   ```bash
   kubectl -n $NAMESPACE_HTTPBIN create -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml
   ```

4. Create a Function using the [supplied code](./assets/function.yaml) in its Namespace:

   ```bash
   kubectl -n $NAMESPACE_FUNCTION apply -f https://raw.githubusercontent.com/kyma-project/kyma/main/docs/03-tutorials/assets/function.yaml
   ```

## Next steps

1. Create a Namespace for the Gateway and APIRule CRs. Run:

   ```bash
   export NAMESPACE={NAMESPACE}
   kubectl create ns $NAMESPACE
   kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
   ```

   >**NOTE:** Skip this step if you already have a Namespace.

2. Export the following values as environment variables:

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME}
   export GATEWAY=$NAMESPACE_APIRULE/httpbin-gateway
   ```
   >**NOTE:** `DOMAIN_NAME` is the domain that you own, for example, api.mydomain.com. If you don't want to use your custom domain, replace `DOMAIN_NAME` with a Kyma domain and `$NAMESPACE/httpbin-gateway` with Kyma's default Gateway `kyma-system/kyma-gateway`.

3. Expose the HttpBin and Function services in their respective Namespaces by creating an APIRule CR which is in its own namespace. Run:

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

4. Call the HttpBin endpoint by sending a `GET` request to the HttpBin service:

   ```bash
   curl -ik -X GET https://httpbin-and-function.$DOMAIN_TO_EXPOSE_WORKLOADS/headers
   ```

5. Call the Function endpoint by sending a `GET` request to the Function service:

   ```bash
   curl -ik -X GET https://httpbin-and-function.$DOMAIN_TO_EXPOSE_WORKLOADS/function
   ```

   These calls return the code `200` response.
