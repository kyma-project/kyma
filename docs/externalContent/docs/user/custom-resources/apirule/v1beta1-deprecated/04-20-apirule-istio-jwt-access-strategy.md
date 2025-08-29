# JWT Access Strategy

To enable Istio JWT, run the following command:

``` sh
kubectl patch configmap/api-gateway-config -n kyma-system --type merge -p '{"data":{"api-gateway-config":"jwtHandler: istio"}}'
```

To enable Oathkeeper JWT, run the following command:

``` sh
kubectl patch configmap/api-gateway-config -n kyma-system --type merge -p '{"data":{"api-gateway-config":"jwtHandler: ory"}}'
```

## Istio JWT Configuration

> [!WARNING]
>  Istio JWT is not a production-ready feature, and the API might change.

This table lists all the possible parameters of the Istio JWT access strategy together with their descriptions:

**Spec:**

| Field                                                                  | Mandatory | Description                                                                                                                                                                                 |
|:-----------------------------------------------------------------------|:----------|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **rules.accessStrategies.config**                                      | **YES**   | Access strategy configuration, must contain at least authentication or authorization.                                                                                                       |
| **rules.accessStrategies.config.authentications**                      | **YES**   | List of authentication objects.                                                                                                                                                             |
| **rules.accessStrategies.config.authentications.issuer**               | **YES**   | Identifies the issuer that issued the JWT. <br/>If the issuer contains `:`, it must be a valid URI.                                                                                         |
| **rules.accessStrategies.config.authentications.jwksUri**              | **YES**   | URL of the provider’s public key set to validate the signature of the JWT. <br/>The value must be an URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints.    |
| **rules.accessStrategies.config.authentications.fromHeaders**          | **NO**    | List of headers from which the JWT token is taken.                                                                                                                                          |
| **rules.accessStrategies.config.authentications.fromHeaders.name**     | **YES**   | Name of the header.                                                                                                                                                                         |
| **rules.accessStrategies.config.authentications.fromHeaders.prefix**   | **NO**    | Prefix used before the JWT token. The default is `Bearer `.                                                                                                                                 |
| **rules.accessStrategies.config.authentications.fromParams**           | **NO**    | List of parameters from which the JWT token is taken.                                                                                                                                       |
| **rules.accessStrategies.config.authorizations**                       | **NO**    | List of authorization objects.                                                                                                                                                              |
| **rules.accessStrategies.config.authorizations.requiredScopes**        | **NO**    | List of required scope values for the JWT.                                                                                                                                                  |
| **rules.accessStrategies.config.authorizations.audiences**             | **NO**    | List of audiences required for the JWT.                                                                                                                                                     |

> [!WARNING]
>  You can define multiple JWT issuers, but each of them must be unique.

> [!WARNING]
>  Currently, we support only a single `fromHeader` or a single `fromParameter`. Specifying both of these fields for a JWT issuer is not supported.

### Authentications
Under the hood, an authentications array creates a corresponding [requestPrincipals](https://istio.io/latest/docs/reference/config/security/authorization-policy/#Source) array in the Istio's [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) resource. Every `requestPrincipals` string is formatted as `<ISSUSER>/*`.

### Authorizations
The authorizations field is optional. When not defined, the authorization is satisfied if the JWT is valid. You can define multiple authorizations for an access strategy. The request is allowed if at least one of them is satisfied.

The **requiredScopes** and **audiences** fields are optional. If **requiredScopes** is defined, the JWT must contain all the scopes in the `scp`, `scope`, or `scopes` claims to be authorized. If **audiences** is defined, the JWT has to contain all the audiences in the `aud` claim to be authorized.

### Example

In the following example, the APIRule has two defined Issuers. The first Issuer, called `ISSUER`, uses a JWT token extracted from the HTTP header. The header is named `X-JWT-Assertion` and has a prefix of `Kyma`. The second Issuer, called `ISSUER2`, uses a JWT token extracted from a URL parameter named `jwt-token`.  
**requiredScopes** defined in the **authorizations** field allow only for JWTs that have the claims `scp`, `scope`, or `scopes` with a value of `test`. Additionally, the JWTs must have an audience of either `example.com` or `example.org`. Alternatively, the JWTs can have the same claims with the `read` and `write` values.

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: service-secured
  namespace: $NAMESPACE
spec:
  gateway: kyma-system/kyma-gateway
  host: httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /headers
      methods: ["GET"]
      mutators: []
      accessStrategies:
        - handler: jwt
          config:
            authentications:
            - issuer: $ISSUER
              jwksUri: $JWKS_URI
              fromHeaders:
              - name: X-JWT-Assertion
                prefix: "Kyma "
            - issuer: $ISSUER2
              jwksUri: $JWKS_URI2
              fromParameters:
              - "jwt_token"
            authorizations:
            - requiredScopes: ["test"]
              audiences: ["example.com", "example.org"]
            - requiredScopes: ["read", "write"]
```