# Configure Authorization (OAuth2, JWT)

With the APIGateway Controller, you can secure the Services you expose in the following ways:

- with OAuth2 tokens
- with JWT tokens
- with both OAuth2 and JWT tokens

If you secure a Service with either OAuth2 or JWT tokens, you must include a valid OAuth2 or JWT token in the `Authorization` header of the call to the Service.

If you secure a Service with both OAuth2 and JWT, the Oathkeeper proxy expects the OAuth2 tokens in the `Authorization` header of incoming calls. For endpoints secured with JWT, you must define the header from which the system extracts the JWT token for every **accessStrategy** you define. Set the **token_from.location** parameter to `header:{NAME}` to extract the JWT token from a specific header. You can use any header name different from `Authorization`.

Alternatively, you can set the **token_from.location** parameter to `query_parameter:{NAME}` to extract the token from a specific query parameter.

> [!TIP]
>  You can define the location of the OAuth2 token through the **token_from.location** parameter. However, by default, OAuth2 tokens are extracted from the `Authorization` header.

## Examples

See these sample excerpts from APIRule custom resources that show the **rules** attribute for Services secured with OAuth2, JWT, and an OAuth2 and JWT combination.

<!-- tabs:start -->
#### OAuth2

  ```yaml
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategy:
        - handler: oauth2_introspection
          config:
            required_scope: ["read"]
  ```

#### JWT

  ```yaml
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategy:
        - handler: jwt
          config:
            trusted_issuers:
            - {issuer URL of your custom OpenID Connect-compliant identity provider}
  ```

####  OAuth2 and JWT

  ```yaml
  rules:
     - path: /.*
       methods: ["GET"]
       mutators: []
       accessStrategy:
         - handler: oauth2_introspection
           config:
             required_scope: ["read"]
         - handler: jwt
           config:
             trusted_issuers:
             - {issuer URL of your custom OpenID Connect-compliant identity provider}
             token_from:
               header: ID-Token
  ```
<!-- tabs:end -->