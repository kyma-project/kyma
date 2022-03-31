---
title: Expose and secure a workload with JWT
---

This tutorial shows how to expose and secure services or Functions using API Gateway Controller. The Controller reacts to an instance of the API Rule custom resource (CR) and creates an Istio Virtual Service and [Oathkeeper Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules) according to the details specified in the CR. To interact with the secured workloads, the tutorial uses a JWT token.

You can use it as a follow-up to the [Use a custom domain to expose a workload](./apix-01-own-domain.md) tutorial.

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create them, follow the [Create a workload](./apix-02-create-workload.md) tutorial.

## Get a JWT access token

1. In your OpenID Connect-compliant (OIDC-compliant) identity provider, create an application to get your client credentials such as Client ID and Client Secret. Export your client credentials as environment variables. Run:

   ```bash
   export CLIENT_ID={YOUR_CLIENT_ID}
   export CLIENT_SECRET={YOUR_CLIENT_SECRET}
   ```

2. Encode your client credentials and export them as an environment variable:

   ```bash
   export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
   ```

3. In your browser, go to `https://{YOUR_OIDC_COMPLIANT_IDENTITY_PROVIDER_INSTANCE}/.well-known/openid-configuration`, save values of the **token_endpoint** and **jwks_uri** parameters, and export them as environment variables:

   ```bash
   export TOKEN_ENDPOINT={YOUR_TOKEN_ENDPOINT}
   export JWKS_URI={YOUR_JWKS_URI}
   ```

4. Get the JWT access token:

   ```bash
   curl -X POST "$TOKEN_ENDPOINT" -d "grant_type=client_credentials" -d "client_id=$CLIENT_ID" -H "Content-Type: application/x-www-form-urlencoded" -H "Authorization: Basic $ENCODED_CREDENTIALS"
   ```

   >**NOTE:** This is just example and may differ for IDP providers.

5. Save the result, and export it as an environment variable:

   ```bash
   export ACCESS_TOKEN={YOUR_ACCESSS_TOKEN}
   ```

## Expose, secure, and access your workload

<div tabs>

  <details>
  <summary>
  HttpBin
  </summary>

1. Export the following value as an environment variable:

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME} #This is a Kyma domain or your custom subdomain e.g. api.mydomain.com.
   export GATEWAY=$NAMESPACE/httpbin-gateway #If you don't want to use your custom domain but a Kyma domain, use the following Kyma Gateway: `kyma-system/kyma-gateway`.
   ```

2. Expose the service and secure it by creating an API Rule CR in your Namespace. If you don't want to use your custom domain but a Kyma domain, use the following Kyma Gateway: `kyma-system/kyma-gateway`. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: gateway.kyma-project.io/v1alpha1
   kind: APIRule
   metadata:
     name: httpbin
     namespace: $NAMESPACE
   spec:
     service:
       name: httpbin
       port: 8000
       host: httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS   
     gateway: $GATEWAY
     rules:
       - accessStrategies:
         - handler: jwt
           config:
             jwks_urls:
             - $JWKS_URI
         methods:
           - GET
         path: /.*
   EOF
   ```

   >**NOTE:** If you are running Kyma on k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

3. To access the secured service, call it using the JWT access token:

   ```bash
   curl -ik https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers -H "Authorization: Bearer $ACCESS_TOKEN"
   ```

   This call returns the code `200` response.
   
  </details>

  <details>
  <summary>
  Function
  </summary>

1. Export the following value as an environment variable:

   ```bash
   export DOMAIN_TO_EXPOSE_WORKLOADS={DOMAIN_NAME} #This is a Kyma domain or your custom subdomain e.g. api.mydomain.com.
   export GATEWAY=$NAMESPACE/httpbin-gateway #If you don't want to use your custom domain but a Kyma domain, use the following Kyma Gateway: `kyma-system/kyma-gateway`.
   ```

2. Expose the Function and secure it by creating an API Rule CR in your Namespace. If you don't want to use your custom domain but a Kyma domain, use the following Kyma Gateway: `kyma-system/kyma-gateway`. Run:

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: gateway.kyma-project.io/v1alpha1
   kind: APIRule
   metadata:
     name: function
     namespace: $NAMESPACE
   spec:
     service:
       name: function
       port: 80
       host: function-example.$DOMAIN_TO_EXPOSE_WORKLOADS   
     gateway: $GATEWAY
     rules:
       - accessStrategies:
         - handler: jwt
           config:
             jwks_urls:
             - $JWKS_URI
         methods:
           - GET
         path: /.*
   EOF
   ```

3. To access the secured Function, call it using the JWT access token:

   ```bash
   curl -ik https://function-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function -H "Authorization: Bearer $ACCESS_TOKEN"
   ```

   This call returns the code `200` response.

  </details>
</div>
