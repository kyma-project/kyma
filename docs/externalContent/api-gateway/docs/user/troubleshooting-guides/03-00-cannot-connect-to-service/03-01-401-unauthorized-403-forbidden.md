# 401 Unauthorized or 403 Forbidden

## Symptom

When you try to reach your service, you get `401 Unauthorized` or `403 Forbidden` in response.

## Cause 

The error `401 Unauthorized` occurs when you try to access a Service that requires authentication, but you have either not provided appropriate credentials or have not provided any credentials at all. You get the error `403 Forbidden` when you try to access a Service or perform an action for which you lack permission.

## Solution

Make sure that you are using an active JSON Web Token (JWT) with proper scopes.

### Remedy

1. Decode the JWT.

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

3. [Generate a new JWT](../../tutorials/01-50-expose-and-secure-a-workload/01-51-get-jwt.md) if needed.