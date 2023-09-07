---
title: Configure Authorization (OAuth2, JWT)
---

With the API Gateway Controller, you can secure the services you expose in the following ways:

- with OAuth2 tokens
- with JWT tokens
- with both OAuth2 and JWT tokens

If you secure a service with either OAuth2 or JWT tokens, by default, you must include a valid OAuth2 or JWT token in the `Authorization` header of the call to the service.

If you secure a service with both OAuth2 and JWT, by default, the Oathkeeper proxy expects OAuth2 tokens in the `Authorization` header of incoming calls. For endpoints secured with JWT, you must define the header from which the system extracts the JWT token for every **accessStrategy** you define. Set the **token_from.location** parameter to `header:{NAME}` to extract the JWT token from a specific header. You can use any header name different from `Authorization`.

Alternatively, you can set the **token_from.location** parameter to `query_parameter:{NAME}` to extract the token from a specific query parameter.

>**TIP:** You can define the location of OAuth2 token through the **token_from.location** parameter. However, by default, OAuth2 tokens are extracted from the `Authorization` header.

## Examples

See these sample excerpts from APIRule custom resources that show the **rules** attribute for services secured with OAuth2, JWT, and an OAuth2 and JWT combination.


<div tabs>
  <details>
  <summary>
  OAuth2
  </summary>

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


  </details>
  <details>
  <summary>
  JWT
  </summary>

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

  </details>
  <details>
  <summary>
  OAuth2 and JWT
  </summary>

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

  </details>

</div>

>**TIP:** To learn more, read about the [APIRule custom resource](./00-custom-resources/apix-01-apirule.md). You can also follow the [tutorials](../03-tutorials/00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-01-expose-and-secure-workload-oauth2.md) to learn how to expose and secure services.
