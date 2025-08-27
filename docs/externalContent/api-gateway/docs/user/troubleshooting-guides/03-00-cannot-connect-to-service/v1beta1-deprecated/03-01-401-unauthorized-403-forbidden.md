# 401 Unauthorized or 403 Forbidden

## Symptom

When you try to reach your service, you get `401 Unauthorized` or `403 Forbidden` in response.

## Cause 

The error `401 Unauthorized` occurs when you try to access a Service that requires authentication, but you have either not provided appropriate credentials or have not provided any credentials at all. You get the error `403 Forbidden` when you try to access a Service or perform an action for which you lack permission.

## Remedy

Make sure that you are using an access token with proper scopes, and it is active. Depending on the type of your access token, follow the relevant steps.

### JSON Web Token

1. Decode the JSON Web Token (JWT).

2. Check the validity and scopes of the JWT:

      ```bash
      {
         "sub": ********,
         "scp": "test",
         "aud": ********,
         "iss": ******,
         "exp": 1697120462,
         "iat": ******,
         "jti": ******,
      }
      ```

3. Generate a [new JWT](../../../tutorials/01-50-expose-and-secure-a-workload/01-51-get-jwt.md) if needed.

### Opaque Access Token

1. Export the credentials of your Oauth2 client as environment variables:

      ```bash
      export CLIENT_ID={CLIENT_ID}
      export CLIENT_SECRET={CLIENT_SECRET}
      ```

2. Encode your client credentials and export them as an environment variable:

      ```bash
      export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
      ```

3. Export the introspection URL. You can find it in the well-known OPENID configuration.

      ```bash
      export INTROSPECTION_URL={INTROSPECTION_URL}
      ```

4. Check the access token's status:

      ```bash
      curl -X POST "$INTROSPECTION_URL" -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "token={ACCESS_TOKEN}"
      ```

5. Generate a [new access token](../../../tutorials/01-50-expose-and-secure-a-workload/v1beta1-deprecated/01-50-expose-and-secure-workload-oauth2.md) if needed.