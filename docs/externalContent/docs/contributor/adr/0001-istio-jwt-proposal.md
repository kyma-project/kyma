# Istio JWT Proposal

## Status

- Accepted

## Context

During sprint 93, we discussed the future structure of Istio JWT.

## Decision

### Structure

We decided on the following structure:

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: service-secured
spec:
  gateway: kyma-system/kyma-gateway
  host: foo.bar
  service:
    name: foo-service
    port: 8080
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategies:
        - handler: jwt
          config:
            authentications:
            - issuer: "https://example.com"
              jwksUri: "https://example.com/.well-known/jwks.json"
              fromHeaders:
              - name: x-jwt-assertion
                prefix: "Bearer "
              fromParams:
              - "my_token"
            - xxx
            authorizations:
            - requiredScopes: ["read"]
              audiences: ["https://example.com", "kyma-goats"]
```

### Subresources

We decided that the authentications and the authorizations will be array fields.

#### Authentications
For each authentication a separate RequestAuthentication will be created.

Selecting a request header to be used as a JWT token is possible with the **fromHeaders** field. The **prefix** field, which specifies a leading string before the actual token value, is optional. Alternatively, you can specify a request URL parameter in the **fromParams** field. Both **fromHeaders** and **fromParams** fields are optional. They are translated respectively to the [JWTRule](https://istio.io/latest/docs/reference/config/security/request_authentication/#JWTRule)'s **fromHeaders** and **fromParams** in the created RequestAuthentication.

>**NOTE:** For now, requests with multiple tokens (at different locations) are not supported.

#### Authorizations
The **requiredScopes** validates the scope of a JWT. The validation of the scope checks for the **scp**, **scope**, or **scopes** claims in the JWT. We support multiple claims because their usage is not standardized and we want to provide backward compatibility with Ory-based configurations.

A request must have all scopes defined in the **requiredScopes** of an authorization. If there are multiple authorizations, the request must fulfill at least one of them to be considered valid.

With Istio, the **requiredScope** is configured as a [Condition](https://istio.io/latest/docs/reference/config/security/authorization-policy/#Condition) of a [Rule](https://istio.io/latest/docs/reference/config/security/authorization-policy/#Rule)
in an [AuthorizationPolicy](https://istio.io/latest/docs/reference/config/security/authorization-policy).

To support the described behavior, we need to create a separate AuthorizationPolicy for each authorization because a request is valid only when all Conditions of a Rule match the request.

In addition, it must be taken into account that multiple scope claims (**scp**, **scope**, **scopes**) are supported. Therefore, a separate Rule must be created in the AuthorizationPolicy for each scope claim, as a request must match at least one Rule.
Having an array of multiple `audiences` defined in **authorizations** ensures that a request is authorized with a JWT token, which has access to all the `audiences` at the same time (`AND` logic).

## Consequences

This configuration differs from the one we used with the Ory JWT implementation. Therefore, customers will need to reconfigure their APIRules after we switch to Istio JWT.