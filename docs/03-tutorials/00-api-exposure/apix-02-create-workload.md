---
title: Create a workload
---

The tutorial comes with a sample HttpBin service deployment and a sample Function.

You can use it as a follow-up to the [Use a custom domain to expose a workload](./apix-01-own-domain.md) tutorial.

## Steps

Follow these steps to deploy an instance of the HttpBin service or a sample Function.

1. Create a Namespace and export its value as an environment variable. Skip the step if you already have a Namespace. Run:

   ```bash
   export NAMESPACE={NAMESPACE_NAME}
   kubectl create ns $NAMESPACE
   kubectl label namespace default istio-injection=enabled --overwrite
   ```

<div tabs>

  <details>
  <summary>
  HttpBin
  </summary>

2. Deploy an instance of the HttpBin service in your Namespace:

   ```bash
   kubectl -n $NAMESPACE create -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml
   ```

  </details>

  <details>
  <summary>
 Function
  </summary>

2. Create a Function in your Namespace using the [supplied code](./assets/function.yaml):

   ```shell
   kubectl -n $NAMESPACE apply -f https://raw.githubusercontent.com/kyma-project/kyma/main/docs/03-tutorials/assets/function.yaml
   ```

  </details>
</div>

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

Once you have your workload deployed, you can continue by choosing one of the following tutorials:

- [Expose a workload](./apix-02-expose-workload-apigateway.md)
- [Expose and secure a workload with OAuth2](./apix-03-expose-and-secure-workload-oauth2.md)
- [Expose and secure a workload with Istio](./apix-05-expose-and-secure-workload-istio.md)
- [Expose and secure a workload with JWT](./apix-05-expose-and-secure-workload-jwt.md)
