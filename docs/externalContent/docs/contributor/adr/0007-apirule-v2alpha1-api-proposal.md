# APIRule `v2alpha1` API Proposal

## Status

Accepted

## Context

Due to the deprecation of Ory and the introduction of new features in API Gateway, the next version of APIRule resource needs to be defined.

## Decision

- **accessStrategies** field is replaced with **extAuths**, **jwt**, and **noAuth**
- multiple hosts are allowed

**Spec:**

| Field                                            | Mandatory | Description                                                                                                                                                                                                                                                                                                                                  | Validation                                                                                                                                                                                                                       |
|:-------------------------------------------------|:---------:|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------| 
| **gateway**                                      |  **YES**  | Specifies the Istio Gateway.                                                                                                                                                                                                                                                                                                                 |                                                                                                                                                                                                                                  |
| **corsPolicy**                                   |  **NO**   | Allows configuring CORS headers sent with the response. If **corsPolicy** is not defined, the default values are applied. If **corsPolicy** is configured, only the CORS headers defined in the APIRule are sent with the response. For more information, see the [decision record](https://github.com/kyma-project/api-gateway/issues/752). |                                                                                                                                                                                                                                  |
| **corsPolicy.allowHeaders**                      |  **NO**   | Specifies headers allowed with the **Access-Control-Allow-Headers** CORS header.                                                                                                                                                                                                                                                             |                                                                                                                                                                                                                                  |
| **corsPolicy.allowMethods**                      |  **NO**   | Specifies methods allowed with the **Access-Control-Allow-Methods** CORS header.                                                                                                                                                                                                                                                             |                                                                                                                                                                                                                                  |
| **corsPolicy.allowOrigins**                      |  **NO**   | Specifies origins allowed with the **Access-Control-Allow-Origins** CORS header.                                                                                                                                                                                                                                                             |                                                                                                                                                                                                                                  |
| **corsPolicy.allowCredentials**                  |  **NO**   | Specifies whether credentials are allowed in the **Access-Control-Allow-Credentials** CORS header.                                                                                                                                                                                                                                           |                                                                                                                                                                                                                                  |
| **corsPolicy.exposeHeaders**                     |  **NO**   | Specifies headers exposed with the **Access-Control-Expose-Headers** CORS header.                                                                                                                                                                                                                                                            |                                                                                                                                                                                                                                  |
| **corsPolicy.maxAge**                            |  **NO**   | Specifies the maximum age of CORS policy cache. The value is provided in the **Access-Control-Max-Age** CORS header.                                                                                                                                                                                                                         |                                                                                                                                                                                                                                  |
| **hosts**                                        |  **YES**  | Specifies the Service's communication address for inbound external traffic. If only the leftmost label is provided, the domain name from the referenced Gateway is used, expanding the host to `<label>.<gateway domain>`.                                                                                                                                                                                         | The full domain name or the leftmost label cannot contain the wildcard character `*`.                                                                                                                                            |
| **service.name**                                 |  **NO**   | Specifies the name of the exposed Service.                                                                                                                                                                                                                                                                                                   |                                                                                                                                                                                                                                  |
| **service.namespace**                            |  **NO**   | Specifies the namespace of the exposed Service.                                                                                                                                                                                                                                                                                              |                                                                                                                                                                                                                                  |
| **service.port**                                 |  **NO**   | Specifies the communication port of the exposed Service.                                                                                                                                                                                                                                                                                     |                                                                                                                                                                                                                                  |
| **timeout**                                      |  **NO**   | Specifies the timeout for HTTP requests in seconds for all Access Rules. The value can be overridden for each Access Rule. </br> If no timeout is specified, the default timeout of 180 seconds applies.                                                                                                                                     | The maximum timeout is limited to 3900 seconds (65 minutes).                                                                                                                                                                     | 
| **rules**                                        |  **YES**  | Specifies the list of Access Rules.                                                                                                                                                                                                                                                                                                          |                                                                                                                                                                                                                                  |
| **rules.service**                                |  **NO**   | Services definitions at this level have higher precedence than the Service definition at the **spec.service** level.                                                                                                                                                                                                                         |                                                                                                                                                                                                                                  |
| **rules.path**                                   |  **YES**  | Specifies the path of the exposed Service. If the configured path of a rule overlaps with the path of another rule, e.g. if a rule defines `/*` in the path, the configuration of both rules is applied.                                                                                                                                      | Value can be either exact path or the path wildcard `/*`.                                                                                                                                                                        |
| **rules.methods**                                |  **NO**   | Specifies the list of HTTP request methods available for **spec.rules.path**. The list of supported methods is defined in [RFC 9910: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html) and [RFC 5789: PATCH Method for HTTP](https://www.rfc-editor.org/rfc/rfc5789.html).                                                        |                                                                                                                                                                                                                                  |
| **rules.mutators**                               |  **NO**   | Specifies the list of the request mutators.                                                                                                                                                                                                                                                                                                  | Currently, the `Headers` and `Cookie` mutators are supported. For more information, see the [documentation](../../user/custom-resources/apirule/v1beta1-deprecated/04-40-apirule-mutators.md). |
| **rules.noAuth**                                 |  **NO**   | Setting `noAuth` to `true` disables authorization.                                                                                                                                                                                                                                                                                           | When `noAuth` is set to true, it is not allowed to define `jwt` or `extAuth` on the same path.                                                                                                                                   |
| **rules.extAuths**                               |  **NO**   | Specifies the list of external authorizers. For more information see below and the [External Authorizer ADR](https://github.com/kyma-project/api-gateway/issues/938).                                                                                                                                                                        |                                                                                                                                                                                                                                  |
| **rules.extAuth.name**                           |  **NO**   | Specifies the name of the external authorizer.                                                                                                                                                                                                                                                                                               |                                                                                                                                                                                                                                  |
| **rules.jwt**                                    |  **NO**   | Specifies the Istio JWT access strategy. For more information see [JWT Access Strategy](../../user/custom-resources/apirule/v1beta1-deprecated/04-20-apirule-istio-jwt-access-strategy.md) and the bellow table.                                                                                           |                                                                                                                                                                                                                                  |
| **rules.jwt.authentications**                    |  **YES**  | Specifies the list of authentication objects.                                                                                                                                                                                                                                                                                                |                                                                                                                                                                                                                                  |
| **rules.jwt.authentications.issuer**             |  **YES**  | Identifies the issuer that issued the JWT. <br/>The value must be a URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints.                                                                                                                                                                                      |                                                                                                                                                                                                                                  |
| **rules.jwt.authentications.jwksUri**            |  **YES**  | Contains the URL of the provider’s public key set to validate the signature of the JWT. <br/>The value must be a URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints.                                                                                                                                         |                                                                                                                                                                                                                                  |
| **rules.jwt.authentications.fromHeaders**        |  **NO**   | Specifies the list of headers from which the JWT token is extracted.                                                                                                                                                                                                                                                                         |                                                                                                                                                                                                                                  |
| **rules.jwt.authentications.fromHeaders.name**   |  **YES**  | Specifies the name of the header.                                                                                                                                                                                                                                                                                                            |                                                                                                                                                                                                                                  |
| **rules.jwt.authentications.fromHeaders.prefix** |  **NO**   | Specifies the prefix used before the JWT token. The default is `Bearer `.                                                                                                                                                                                                                                                                    |                                                                                                                                                                                                                                  |
| **rules.jwt.authentications.fromParams**         |  **NO**   | Specifies the list of parameters from which the JWT token is extracted.                                                                                                                                                                                                                                                                      |                                                                                                                                                                                                                                  |
| **rules.jwt.authorizations**                     |  **NO**   | Specifies the list of authorization objects.                                                                                                                                                                                                                                                                                                 |                                                                                                                                                                                                                                  |
| **rules.jwt.authorizations.requiredScopes**      |  **NO**   | Specifies the list of required scope values for the JWT.                                                                                                                                                                                                                                                                                     |                                                                                                                                                                                                                                  |
| **rules.jwt.authorizations.audiences**           |  **NO**   | Specifies the list of audiences required for the JWT.                                                                                                                                                                                                                                                                                        |                                                                                                                                                                                                                                  |
| **rules.timeout**                                |  **NO**   | Specifies the timeout, in seconds, for HTTP requests made to **spec.rules.path**. Timeout definitions set at this level take precedence over any timeout defined at the **spec.timeout** level.                                                                                                                                              | The maximum timeout is limited to 3900 seconds (65 minutes).                                                                                                                                                                     |

### Examples

- Multiple hosts with external authorizers and JWT:
```yaml
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: service-config
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - api1.example.com
    - api2.example.com
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /headers
      methods: ["GET"]
      extAuths:
        - name: oauth2-proxy
        - name: geo-blocker
      jwt:
        authentications:
          - issuer: https://example.com
            jwksUri: https://example.com/.well-known/jwks.json
        authorizations:
          - audiences: ["app1"]
```

- One host with JWT:
```yaml
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: service-config
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - app1.example.com
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /* # Should be warning user that it is not recommended, as it applies to all paths
      methods: ["GET"]
      jwt:
        authentications:
          - issuer: https://example.com
            jwksUri: https://example.com/.well-known/jwks.json
        authorizations:
          - audiences: ["app1"]
```

- One host with `noAuth`:
```yaml
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: service-config
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - app1.example.com
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /headers
      methods: ["GET"]
      noAuth: true
```

- Istio mutators:
```yaml
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: service-config
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - app1.example.com
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /headers
      methods: ["GET"]
      mutators:
        - handler: header
          config:
            headers:
              X-Custom-Auth: "%REQ(Authorization)%"
              X-Some-Data: "some-data"
        - handler: cookie
          config:
            cookies:
              user: "test"
      jwt:
        authentications:
          - issuer: https://example.com
            jwksUri: https://example.com/.well-known/jwks.json
        authorizations:
          - audiences: ["app1"]
```

- Multiple paths with different configurations:
```yaml
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: service-config
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - api1.example.com
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /test
      noAuth: true
    - path: /login
      extAuths:
        - name: geoBlocking
    - path: /headers
      jwt:
        authentications:
          - issuer: https://example.com
            jwksUri: https://example.com/.well-known/jwks.json
        authorizations:
          - audiences: ["app1"]
    - path: /image
      extAuths:
        - name: oauth2-proxy
```
