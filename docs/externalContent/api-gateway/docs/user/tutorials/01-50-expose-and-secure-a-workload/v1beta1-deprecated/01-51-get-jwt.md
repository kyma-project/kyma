# Get a JSON Web Token (JWT)

Learn how to get a JSON Web Token (JWT), which you can use to access secured endpoints.

## Prerequisites

You use an OpenID Connect-compliant (OIDC-compliant) identity provider.

## Steps

1. Create an application in your OIDC-compliant identity provider. Save the client credentials: Client ID and Client Secret. 

2. In the URL `https://{YOUR_IDENTITY_PROVIDER_INSTANCE}/.well-known/openid-configuration` replace `{YOUR_IDENTITY_PROVIDER_INSTANCE}` with the name of your OIDC-compliant identity provider instance. Then, open the link in your browser. Save the values of the **token_endpoint**, **jwks_uri**, and **issuer** parameters.

3. Export the saved values as environment variables:
      
      ```bash
      export CLIENT_ID={YOUR_CLIENT_ID}
      export CLIENT_SECRET={YOUR_CLIENT_SECRET}
      export TOKEN_ENDPOINT={YOUR_TOKEN_ENDPOINT}
      ```
      
4. Encode your client credentials and export them as an environment variable:

      ```bash
      export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
      ```

5. To obtain a JWT, run the following command:

      ```bash
      curl -X POST "$TOKEN_ENDPOINT" -d "grant_type=client_credentials" -d "client_id=$CLIENT_ID" -H "Content-Type: application/x-www-form-urlencoded" -H "Authorization: Basic $ENCODED_CREDENTIALS"
      ```