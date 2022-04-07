---
title: Get a JWT
---

This tutorial shows how to get a JWT token, which can be used to access secured endpoints created in [Expose and secure a workload with Istio](./apix-04-expose-and-secure-workload-istio.md) and [Expose and secure a workload with JWT](./apix-04-expose-and-secure-workload-jwt.md) tutorials.

## Prerequisites

To follow this tutorial, you need to use OpenID Connect-compliant (OIDC-compliant) identity provider.

## Get a JWT

1. In your OpenID Connect-compliant (OIDC-compliant) identity provider, create an application to get your client credentials such as Client ID and Client Secret. Export your client credentials as environment variables. Run:

   ```bash
   export CLIENT_ID={YOUR_CLIENT_ID}
   export CLIENT_SECRET={YOUR_CLIENT_SECRET}
   ```

2. Encode your client credentials and export them as an environment variable:

   ```bash
   export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
   ```

3. In your browser, go to `https://YOUR_OIDC_COMPLIANT_IDENTITY_PROVIDER_INSTANCE/.well-known/openid-configuration`, save values of the **token_endpoint**, **jwks_uri** and **issuer** parameters, and export them as environment variables:

   ```bash
   export TOKEN_ENDPOINT={YOUR_TOKEN_ENDPOINT}
   export JWKS_URI={YOUR_JWKS_URI}
   export ISSUER={YOUR_ISSUER}
   ```

4. Get the JWT:

   ```bash
   curl -X POST "$TOKEN_ENDPOINT" -d "grant_type=client_credentials" -d "client_id=$CLIENT_ID" -H "Content-Type: application/x-www-form-urlencoded" -H "Authorization: Basic $ENCODED_CREDENTIALS"
   ```

5. Save the token and export it as an environment variable:

   ```bash
   export ACCESS_TOKEN={YOUR_ACCESSS_TOKEN}
   ```

## Next steps

Once you have your **ACCESS_TOKEN**, you can use it in the following tutorials:

- [Expose and secure a workload with Istio](./apix-04-expose-and-secure-workload-istio.md)
- [Expose and secure a workload with JWT](./apix-04-expose-and-secure-workload-jwt.md)
