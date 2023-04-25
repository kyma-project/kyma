---
title: API Rule
---

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format the API Gateway Controller listens for. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```shell
kubectl get crd apirules.gateway.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample custom resource (CR) that the API Gateway Controller listens for to expose a service. This example has the **rules** section specified which makes the API Gateway Controller create an Oathkeeper Access Rule for the service.

<div tabs name="api-rule" group="sample-cr">
  <details>
  <summary label="v1beta1">
  v1beta1
  </summary>

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
        - handler: oauth2_introspection
          config:
            required_scope: ["read"]
```

  </details>
  <details>
  <summary label="v1alpha1">
  v1alpha1
  </summary>

>**NOTE:** Since Kyma 2.5 the `v1alpha1` resource has been deprecated. However, you can still create it. It is stored as `v1beta1`.

```yaml
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: service-secured
spec:
  gateway: kyma-system/kyma-gateway
  service:
    name: foo-service
    port: 8080
    host: foo.bar
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategies:
        - handler: oauth2_introspection
          config:
            required_scope: ["read"]
```

  </details>
</div>

## Specification

This table lists all the possible parameters of a given resource together with their descriptions:

| Field   |      Mandatory      |  Description |
|---|:---:|---|
| **metadata.name** | **YES** | Specifies the name of the exposed API. |
| **spec.gateway** | **YES** | Specifies the Istio Gateway. |
| **spec.host** | **YES** | Specifies the service's communication address for inbound external traffic. If only the leftmost label is provided, the default domain name will be used. |
| **spec.service.name** | **NO** | Specifies the name of the exposed service. |
| **spec.service.namespace** | **NO** | Specifies the Namespace of the exposed service. |
| **spec.service.port** | **NO** | Specifies the communication port of the exposed service. |
| **spec.rules** | **YES** | Specifies the list of Oathkeeper access rules. |
| **spec.rules.service** | **NO** | Services definitions at this level have higher precedence than the service definition at the **spec.service** level.|
| **spec.rules.service.name** | **NO** | Specifies the name of the exposed service. |
| **spec.rules.service.namespace** | **NO** | Specifies the Namespace of the exposed service. |
| **spec.rules.service.port** | **NO** | Specifies the communication port of the exposed service. |
| **spec.rules.path** | **YES** | Specifies the path of the exposed service. |
| **spec.rules.methods** | **NO** | Specifies the list of HTTP request methods available for **spec.rules.path**. |
| **spec.rules.mutators** | **NO** | Specifies the list of [Oathkeeper mutators](https://www.ory.sh/docs/next/oathkeeper/pipeline/mutator). |
| **spec.rules.accessStrategies** | **YES** | Specifies the list of [Oathkeeper](https://www.ory.sh/docs/next/oathkeeper/pipeline/authn) access strategies. Supported are `oauth2_introspection`, `noop` and `allow`. We also support `jwt` as [Istio JWT](https://istio.io/latest/docs/tasks/security/authorization/authz-jwt/) access strategy. |

>**CAUTION:** If `service` is not defined at **spec.service** level, all defined rules must have `service` defined at **spec.rules.service** level, otherwise the validation fails.
>**CAUTION:** We currently support only one access strategy per `rule`.

### Istio JWT access strategy configuration

This table lists all the possible parameters of the Istio JWT access strategy together with their descriptions:

| Field                                                                     | Description                                                                |
|:--------------------------------------------------------------------------|:---------------------------------------------------------------------------|
| **spec.rules.accessStrategies.config.authentications**                    | List of authentication objects.                                            |
| **spec.rules.accessStrategies.config.authentications.issuer**             | Identifies the issuer that issued the JWT.                                 |
| **spec.rules.accessStrategies.config.authentications.jwksUri**            | URL of the provider’s public key set to validate the signature of the JWT. |
| **spec.rules.accessStrategies.config.authentications.fromHeaders**        | List of headers from which the JWT token is taken.                         |
| **spec.rules.accessStrategies.config.authentications.fromHeaders.name**   | Name of the header.                                                        |
| **spec.rules.accessStrategies.config.authentications.fromHeaders.prefix** | Prefix used before the JWT token. The default is `Bearer `.                |
| **spec.rules.accessStrategies.config.authentications.fromParams**         | List of parameters from which the JWT token is taken.                      |
| **spec.rules.accessStrategies.config.authorizations**                     | List of authorization objects.                                             |
| **spec.rules.accessStrategies.config.authorizations.requiredScopes**      | List of required scope values for the JWT.                                 |
| **spec.rules.accessStrategies.config.authorizations.audiences**           | List of audiences required for the JWT.                                    |

>**CAUTION:** Currently, we support only a single `fromHeader` **or** a single `fromParameter`. Specifying both of these fields for a JWT issuer is not supported.

Example:

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
              # JWT token must be taken from a HTTP header called X-JWT-Assertion and it will have a "Kyma " prefix
              - name: X-JWT-Assertion
                prefix: "Kyma "
            - issuer: $ISSUER2
              jwksUri: $JWKS_URI2
              fromParameters:
              # JWT token must be taken from URL parameter called jwt_token
              - "jwt_token"
            authorizations:
            # Allow only JWTs with the claim "scp", "scope" or "scopes" with the value "test" and the audience "example.com" and "example.org"
            # or JWTs with the claim "scp", "scope" or "scopes" with the values "read" and "write"
            - requiredScopes: ["test"]
              audiences: ["example.com", "example.org"]
            - requiredScopes: ["read", "write"]
```

#### Authentications
Under the hood, an authentications array creates a corresponding [requestPrincipals](https://istio.io/latest/docs/reference/config/security/authorization-policy/#Source) array in the Istio's [Authorization Policy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) resource. Every `requestPrincipals` string is formatted as `<ISSUSER>/*`.

#### Authorizations
The authorizations field is optional. When not defined, the authorization is satisfied if the JWT is valid. You can define multiple authorizations for an access strategy. When multiple authorizations are defined, the request is allowed if at least one of them is satisfied.

The `requiredScopes` and `audiences` fields are optional. If `requiredScopes` is defined, the JWT has to contain all the scopes in the `scp`, `scope` or `scopes` claims as in the `requiredScopes` field in order to be authorized. If `audiences` is defined, the JWT has to contain all the audiences in the `aud` claim as in the `audiences` field in order to be authorized.

### Mutators
Different types of mutators are supported depending on the access strategy.

| Access Strategy      | Mutator support                                                     |
|:---------------------|:--------------------------------------------------------------------|
| jwt                  | Istio-based cookie and header mutator                               |
| oauth2_introspection | [Ory mutators](https://www.ory.sh/docs/oathkeeper/pipeline/mutator) |
| noop                 | [Ory mutators](https://www.ory.sh/docs/oathkeeper/pipeline/mutator) |
| allow                | No mutators supported                                               |

#### Istio-based Mutators
Mutators can be used to enrich an incoming request with information. The following mutators are supported in combination with the `jwt` access strategy and can be defined for each rule in an `ApiRule`: `header`,`cookie`. It's possible to configure multiple mutators for one rule, but only one mutator of each type is allowed.

#### Header Mutator
The headers are specified via the `headers` field of the header mutator configuration field. The keys are the names of the headers and the values are a string. In the header value it is possible to use [Envoy command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators), e.g. to write an incoming header to a new header. The configured headers are set to the request and overwrite all existing headers with the same name.

Example:
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
      mutators:
        - handler: header
          config:
            headers:
              # Add a new header called X-Custom-Auth with the value of the incoming Authorization header
              X-Custom-Auth: "%REQ(Authorization)%"
              # Add a new header called X-Some-Data with the value "some-data"
              X-Some-Data: "some-data"
      accessStrategies:
        - handler: jwt
          config:
            authentications:
              - issuer: $ISSUER
                jwksUri: $JWKS_URI
```

#### Cookie Mutator
The cookies are specified via the `cookies` field of the cookie mutator configuration field. The keys are the names of the cookies and the values are a string. In the cookie value it is possible to use [Envoy command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators). The configured cookies are set as `Cookie`-header in the request and overwrite an existing `Cookie`-header.

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
      mutators:
        - handler: cookie
          config:
            cookies:
              # Add a new cookie called some-data with the value "data"
              some-data: "data"
      accessStrategies:
        - handler: jwt
          config:
            authentications:
              - issuer: $ISSUER
                jwksUri: $JWKS_URI
```

## Additional information

When you fetch an existing APIRule CR, the system adds the **status** section which describes the status of the VirtualService and the Oathkeeper Access Rule created for this CR. This table lists the fields of the **status** section.

| Field   |  Description |
|:---|:---|
| **status.apiRuleStatus** | Status code describing the APIRule CR. |
| **status.virtualServiceStatus.code** | Status code describing the VirtualService. |
| **status.virtualService.desc** | Current state of the VirtualService. |
| **status.accessRuleStatus.code** | Status code describing the Oathkeeper Rule. |
| **status.accessRuleStatus.desc** | Current state of the Oathkeeper Rule. |

### Status codes

These are the status codes used to describe the VirtualServices and Oathkeeper Access Rules:

| Code   |  Description |
|---|---|
| **OK** | Resource created. |
| **SKIPPED** | Skipped creating a resource. |
| **ERROR** | Resource not created. |
