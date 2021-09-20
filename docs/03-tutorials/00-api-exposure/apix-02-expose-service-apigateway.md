---
title: Expose a service
---

This tutorial shows how to expose service endpoints and configure different allowed HTTP methods for them using API Gateway Controller.

The tutorial comes with a sample HttpBin service deployment and is a follow-up to the [Use a custom domain to expose a service](./apix-03-own-domain.md) tutorial.

## Deploy and expose a service

Follow the instruction to deploy an unsecured instance of the HttpBin service and expose it.

1. Deploy an instance of the HttpBin service in your Namespace:

  ```bash
  kubectl -n ${NAMESPACE_NAME} create -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml
  ```

2. Export these values as environment variables:

  ```bash
  export NAMESPACE={NAMESPACE_NAME} #If you don't have a Namspeace yet, create one.
  export TLS_SECRET={SECRET_NAME} #e.g. use the TLS_SECRET from your Certificate CR i.e. httpbin-tls-credentials.
  export WILDCARD={WILDCRAD_SUBDOMAIN} #e.g. *.api.mydomain.com
  export DOMAIN={CLUSTER_DOMAIN} #This is a Kyma domain or your custom subdomain e.g. api.mydomain.com.
  ```

3. Create a Gateway CR. Skip this step if you use a Kyma domain instead of your custom domain. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: networking.istio.io/v1alpha3
   kind: Gateway
   metadata:
     name: httpbin-gateway
     namespace: $NAMESPACE
   spec:
     selector:
       istio: ingressgateway # Use Istio Ingress Gateway as default
     servers:
       - port:
           number: 443
           name: https
           protocol: HTTPS
         tls:
           mode: SIMPLE
           credentialName: $TLS_SECRET
         hosts:
           - "$WILDCARD"
   EOF
   ```

4. Expose the service by creating an APIRule CR in your Namespace. If you don't want to use your custom domain but a Kyma domain, use the following Kyma Gateway: `kyma-gateway.kyma-system.svc.cluster.local`. Run:

  ```bash
  cat <<EOF | kubectl apply -f -
  apiVersion: gateway.kyma-project.io/v1alpha1
  kind: APIRule
  metadata:
    name: httpbin
    namespace: $NAMESPACE
  spec:
    service:
      host: httpbin.$DOMAIN
      name: httpbin
      port: 8000
    gateway: httpbin-gateway.namespace-name.svc.cluster.local #The value corresponds to the Gateway CR you created.
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

  >**NOTE:** If you are running Kyma on k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

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
