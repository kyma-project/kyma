---
title: Expose and secure a service with JWT
---

This tutorial shows how to expose and secure services or Functions using API Gateway Controller. The controller reacts to an instance of the API Rule custom resource (CR) and creates an Istio Virtual Service and [Oathkeeper Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules) according to the details specified in the CR. To interact with the secured services, the tutorial uses a JWT client.

The tutorial may be a follow-up to the [Use a custom domain to expose a service](./apix-01-own-domain.md) tutorial.

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Deploy a service](./apix-02-deploy-service.md) tutorial.

## Get a JWT access token

1. In your OpenID Connect (OIDC) compliant identity provider, create an application to get your client credentials such as Client ID and Client Secret. Export your client credentials as environment variables. Run:

```bash
export CLIENT_ID={YOUR_CLIENT_ID}
export CLIENT_SECRET={YOUR_CLIENT_SECRET}
```

>**TIP:** For testing purposes, you can use the client credentials from `https://demo.c2id.com/oidc-client/`. We **don't** recommend the solution for production environments.

2. Encode your client credentials and export them as an environment variable:

```bash
export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
```

3. In your browser, go to `https://{YOUR_OIDC_COMPLIENT_IDENTITY_PROVIDER_INSTANCE}/.well-known/openid-configuration`, save values of the **token_endpoint** and **jwks_uri** parameters, and export them as environment variables:

```bash
export TOKEN_ENDPOINT={YOUR_TOKEN_ENDPOINT}
export JWKS_URI={YOUR_JWKS_URI}
```

>**TIP:** For testing purposes, you can use values of the **token_endpoint** and **jwks_uri** from `https://accounts.google.com/.well-known/openid-configuration`. We **don't** recommend the solution for production environments.

4. Get the JWT access token:

```bash
curl -X POST "$TOKEN_ENDPOINT" -d "grant_type=client_credentials" -d "client_id=$CLIENT_ID" -H "Content-Type: application/x-www-form-urlencoded" -H "Authorization: Basic $ENCODED_CREDENTIALS"
```

## Expose, secure, and access your resources

<div tabs>

  <details>
  <summary>
  HttpBin
  </summary>

1. Expose the service and secure it by creating an API Rule CR in your Namespace. If you don't want to use your custom domain but a Kyma domain, use the following Kyma Gateway: `kyma-system/kyma-gateway`. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: gateway.kyma-project.io/v1alpha1
   kind: APIRule
   metadata:
     name: httpbin
     namespace: $NAMESPACE
   spec:
     gateway: kyma-gateway.kyma-system.svc.cluster.local
     rules:
       - accessStrategies:
           - config:
               jwks_urls:
               - $JWKS_URI
             handler: jwt
         methods:
           - GET
         path: /.*
     service:
       name: httpbin
       port: 8000
       host: httpbin.$DOMAIN
   ```

   >**NOTE:** If you are running Kyma on k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

  </details>

  <details>
  <summary>
  Function
  </summary>

1. Expose the Function and secure it by creating an API Rule CR in your Namespace. If you don't want to use your custom domain but a Kyma domain, use the following Kyma Gateway: `kyma-system/kyma-gateway`. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: gateway.kyma-project.io/v1alpha1
   kind: APIRule
   metadata:
     name: function
     namespace: $NAMESPACE
   spec:
     gateway: namespace-name/httpbin-gateway #The value corresponds to the Gateway CR you created.
     service:
       name: function
       port: 80
       host: function-example.$DOMAIN
     rules:
     - accessStrategies:
         - config:
             jwks_urls:
             - $JWKS_URI
           handler: jwt
       methods:
         - GET
       path: /.*
   EOF
   ```

  </details>
</div>

   2. To access the secured service, call it using a JWT access token:

   ```bash
   curl -ik https://mst.dt-test.goatz.shoot.canary.k8s-hana.ondemand.com/headers -H "Authorization: Bearer $ACCESS_TOKEN"
   ```
