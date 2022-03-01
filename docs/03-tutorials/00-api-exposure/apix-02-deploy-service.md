---
title: Deploy a service
---

Follow these steps to deploy an instance of the HttpBin service or a sample Function.

The tutorial comes with a sample HttpBin service deployment and a sample Function. It may be a follow-up to the [Use a custom domain to expose a service](./apix-01-own-domain.md) tutorial.

<div tabs>

  <details>
  <summary>
  HttpBin
  </summary>

1. Deploy an instance of the HttpBin service in your Namespace:

   ```bash
   kubectl -n ${NAMESPACE_NAME} create -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml
   ```

  </details>

  <details>
  <summary>
 Function
  </summary>

1. Create a Function in your Namespace using the [supplied code](./assets/function.yaml):

   ```shell
   kubectl -n ${NAMESPACE_NAME} apply -f https://raw.githubusercontent.com/kyma-project/kyma/main/docs/03-tutorials/assets/function.yaml
   ```

  </details>
</div>

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

## Next steps

Once you have your service deployed, you can continue by choosing one of the following tutorials to:

- [Expose a service](../../../03-tutorials/00-api-exposure/apix-02-expose-service-apigateway.md)
- [Expose and secure a service with OAuth2](../../../03-tutorials/00-api-exposure/apix-03-expose-and-secure-service-oauth2.md)
- [Expose and secure a service with JTW](../../../03-tutorials/00-api-exposure/apix-04-expose-and-secure-service-jwt.md)
