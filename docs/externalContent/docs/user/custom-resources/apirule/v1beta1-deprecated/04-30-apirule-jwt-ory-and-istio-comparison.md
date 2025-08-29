# Differences Between Ory Oathkeeper and Istio JWT Access Strategies

We are in the process of transitioning from Ory Oathkeeper to Istio JWT access strategy. This document explains the differences between those two strategies and compares their configuration.

## Corresponding JWT Configuration Properties in Ory Oathkeeper and Istio

This table lists all possible configuration properties of the Ory Oathkeeper JWT access strategy and their corresponding properties in Istio:

| Ory Oathkeeper                 | Required  |        | Istio                                                                           | Required |
|--------------------------------|:---------:|--------|---------------------------------------------------------------------------------|:--------:|
| **jwks_urls**                  | **YES**   | &rarr; | **authentications.jwksUri**                                                     | **YES**  |
| **trusted_issuers**            |  **NO**   | &rarr; | **authentications.issuer**                                                      | **YES**  |
| **token_from.header**          |  **NO**   | &rarr; | **authentications.fromHeaders.name**<br/>**authentications.fromHeaders.prefix** |  **NO**  |
| **token_from.query_parameter** |  **NO**   | &rarr; | **authentications.fromParams**                                                  |  **NO**  |
| **token_from.cookie**          |  **NO**   | &rarr; | *Not Supported*                                                                 |  **-**   |
| **target_audience**            |  **NO**   | &rarr; | **authorizations.audiences**                                                    |  **NO**  |
| **required_scope**             |  **NO**   | &rarr; | **authorizations.requiredScopes**                                               |  **NO**  |
| **scope_strategy**             |  **NO**   | &rarr; | *Not Supported*                                                                 |  **-**   |
| **jwks_max_wait**              |  **NO**   | &rarr; | *Not Supported*                                                                 |  **-**   |
| **jwks_ttl**                   |  **NO**   | &rarr; | *Not Supported*                                                                 |  **-**   |
| **allowed_algorithms**         |  **NO**   | &rarr; | *Not Supported*                                                                 |  **-**   |

## Examplary APIRule Custom Resources

> [!WARNING]
>  Istio JWT is not a production-ready feature, and the API might change.

These are sample APIRule custom resources (CRs) of both Ory Oathkeeper and Istio JWT access strategy configuration for a Service.

<!-- tabs:start -->
#### Ory Oathkeeper

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
    namespace: foo-namespace
    port: 8080
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategies:
        - handler: jwt
          config:
            trusted_issuers:
              - $ISSUER1
              - $ISSUER2
            jwks_urls:
              - $JWKS_URI1
              - $JWKS_URI2
            required_scope:
              - "test"
            target_audience:
              - "example.com"
              - "example.org"
            token_from:
              header: X-JWT-Assertion
```
#### Istio

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: service-secured
  namespace: $NAMESPACE
spec:
  gateway: kyma-system/kyma-gateway
  host: foo.bar
  service:
    name: foo-service
    namespace: foo-namespace
    port: 8080
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategies:
        - handler: jwt
          config:
            authentications:
            - issuer: $ISSUER1
              jwksUri: $JWKS_URI1
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
<!-- tabs:end -->

> [!WARNING]
>  Both `jwks_urls` and `trusted_issuers` must be valid URIs. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints. 

> [!WARNING]
>  You can define multiple JWT issuers, but each of them must be unique.

> [!WARNING]
> We support only a single `fromHeader` or a single `fromParameter` for a JWT issuer.

## How Istio JWT Access Strategy Differs from Ory Oathkeeper JWT Access Strategy

### Configuration of Properties Handling in Ory Oathkeeper and Istio Resources

When you use Ory Oathkeeper, the APIRule JWT access strategy configuration is translated directly as [authenticator configuration](https://www.ory.sh/docs/oathkeeper/api-access-rules#handler-configuration) in the [Ory Oathkeeper Access Rule CR](https://www.ory.sh/docs/oathkeeper/api-access-rules). See the official Ory Oathkeeper [JWT authenticator documentation](https://www.ory.sh/docs/oathkeeper/pipeline/authn#jwt) to learn more.

With Istio JWT access strategy, for each `authentications` entry, an Istio's [Request Authentication](https://istio.io/latest/docs/reference/config/security/request_authentication/) resource is created, and for each `authorizations` entry, an [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) resource is created.

### Header Support
Istio JWT access strategy only supports `header` and `cookie` mutators. Learn more about [supported mutators](./04-40-apirule-mutators.md).

### Regex Type of Path Matching
Istio doesn't support regex type of path matching in Authorization Policies. Ory Oathkeeper Access Rules and Virtual Service do support this feature.

### Configuring a JWT Token from `cookie`
Istio doesn't support configuring a JWT token from `cookie`, and Ory Oathkeeper does. Istio supports only `fromHeaders` and `fromParams` configurations.

### Workload in the Service Mesh
Using Istio as JWT access strategy requires the workload behind the Service to be in the service mesh, for example, to have the Istio proxy injected. Learn how to [add workloads to the Istio service mesh](https://istio.io/latest/docs/ops/common-problems/injection/).

### Change of Status `401` to `403` When Calling an Endpoint Without the `Authorization` Header
Previously, when using ORY Oathkeeper, if you called a secured workload without a JWT token, it resulted in the `401` error. This behavior has changed with the implementation of Istio-based JWT. Now, calls made without a token result in the `403` error. To learn more, read the Istio documentation on [RequestAuthentication](https://istio.io/latest/docs/concepts/security/#request-authentication) and [AuthorizationPolicy](https://istio.io/latest/docs/reference/config/security/authorization-policy).

### Blocking of In-Cluster Connectivity to an Endpoint
Istio JWT uses the `istio-sidecar` container to validate requests in the context of the target Pod. Previously, in-cluster requests were allowed in the `ory-oathkeeper` context because request validation happened within the `ory-oathkeeper` Pod. Now, these requests fail unless they are explicitly permitted.


