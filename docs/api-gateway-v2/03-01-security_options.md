---
title: Available security options
type: Details
---

The API Gateway Controller allows you to secure the services you expose in 3 different ways:

- with OAuth2 tokens
- with JWT tokens
- with both OAuth2 and JWT tokens

If you secure a service with either OAuth2 or JWT tokens, you must include a valid OAuth2 or JWT token in the `Authorization` header of the call to the service.

If you secure a service with both OAuth2 and JWT, the Oathkeeper proxy expects OAuth2 tokens in the `Authorization` header of incoming calls. For endpoints secured with JWT, you must define the header from which the system extracts the JWT token for every **accessStrategy** you define. You can use any header name different from `Authorization`.

See these sample excerpts from APIRule custom resources that show the **rules** attribute for services secured with OAuth2, JWT, and an OAuth2 and JWT combination.

>**TIP:** To learn more about the APIRule custom resource, read [this](/#custom-resource-api-rule) document. To learn how to expose and secure services, follow [this](#tutorials-expose-and-secure-a-service) tutorial.

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
            - https://dex.{CLUSTER_DOMAIN}
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
             - https://dex.$DOMAIN
             token_from:
               header: ID-Token
  ```

  </details>

</div>
