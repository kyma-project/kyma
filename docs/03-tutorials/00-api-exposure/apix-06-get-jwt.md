---
title: Get a JSON Web Token (JWT)
---

This tutorial shows how to get a JSON Web Token (JWT), which can be used to access secured endpoints created in the tutorials [Expose and secure a workload with Istio](./apix-07-expose-and-secure-workload-istio.md) and [Expose and secure a workload with JWT](./apix-08-expose-and-secure-workload-jwt.md).

## Prerequisites

You are using an OpenID Connect-compliant (OIDC-compliant) identity provider.

## Get a JWT

1. In your OIDC-compliant identity provider, create an application to get your client credentials such as Client ID and Client Secret. 

2. Export your client credentials as environment variables. Run:

   ```bash
   export CLIENT_ID={YOUR_CLIENT_ID}
   export CLIENT_SECRET={YOUR_CLIENT_SECRET}
   ```

2. Encode your client credentials and export them as an environment variable:

   ```bash
   export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
   ```

3. In your browser, go to `https://YOUR_OIDC_COMPLIANT_IDENTITY_PROVIDER_INSTANCE/.well-known/openid-configuration`, save the values of the **token_endpoint**, **jwks_uri** and **issuer** parameters, and export them as environment variables:

   ```bash
   export TOKEN_ENDPOINT={YOUR_TOKEN_ENDPOINT}
   export JWKS_URI={YOUR_JWKS_URI}
   export ISSUER={YOUR_ISSUER}
   ```

4. Get the JWT:

   ```bash
   curl -X POST "$TOKEN_ENDPOINT" -d "grant_type=client_credentials" -d "client_id=$CLIENT_ID" -H "Content-Type: application/x-www-form-urlencoded" -H "Authorization: Basic $ENCODED_CREDENTIALS"
   ```

5. Save the JWT and export it as an environment variable:

   ```bash
   export ACCESS_TOKEN={YOUR_ACCESSS_TOKEN}
   ```
## Result

You have created your access token.

## Next steps

You can use your access token in the following tutorials:

- [Expose and secure a workload with Istio](./apix-07-expose-and-secure-workload-istio.md)
- [Expose and secure a workload with JWT](./apix-08-expose-and-secure-workload-jwt.md)
