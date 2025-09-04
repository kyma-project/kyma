# APIRule v2 Access Strategies

APIRule allows you to define the security configuration for an exposed endpoint using the concept of access strategies. The supported access strategies for APIRule `v2` are **noAuth** and **jwt**.

## Configuration of the **noAuth** Access Strategy

The intended functionality of this access strategy is to provide a simple configuration for exposing workloads.
It only allows access to the specified HTTP methods of the exposed workload.

```yaml
...
rules:
  - path: /headers
    methods: ["GET"]
    noAuth: true
```

## Configuration of the **jwt** Access Strategy

In version `v2` of the APIRule CR, you can use this access strategy only with the Istio JWT configuration. Additionally, defining only one issuer is supported.

```yaml
...
rules:
  - path: /headers
    methods: ["GET"]
    jwt:
      authentications:
        - issuer: https://example.com
          jwksUri: https://example.com/.well-known/jwks.json
      authorizations:
        - audiences: ["app1"]
```

### Authentications
Under the hood, an authentications array creates a corresponding **requestPrincipals** array in the Istio’s Authorization Policy resource. Every **requestPrincipals** string is formatted as `<ISSUSER>/*`.

### Authorizations
The authorizations field is optional. When not defined, the authorization is satisfied if the JWT is valid. You can define multiple authorizations for an access strategy. The request is allowed if at least one of them is satisfied.

The **requiredScopes** and **audiences** fields are optional. If the **requiredScopes** field is defined, the JWT must contain all the scopes in the scp, scope, or scopes claims to be authorized. If the **audiences** field is defined, the JWT has to contain all the audiences in the aud claim to be authorized.

In the following example, the APIRule has two defined Issuers. The first Issuer, called `ISSUER`, uses a JWT token extracted from the HTTP header. The header is named `X-JWT-Assertion` and has a prefix of `Kyma`. The second Issuer, called `ISSUER2`, uses a JWT token extracted from a URL parameter named `jwt-token`.
**requiredScopes** defined in the **authorizations** field allow only for JWTs that have the claims **scp**, **scope**, or **scopes** with a value of `test` and an audience of either `example.com` or `example.org`. Alternatively, the JWTs can have the same claims with the `read` and `write` values.

```yaml
apiVersion: gateway.kyma-project.io/v2
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
      jwt:
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

## Configuration of the **extAuth** Access Strategy

**extAuth** is an access strategy allowing for providing custom authentication and authorization logic. To use it, you must first define the authorization provider in the Istio configuration, most commonly in the Istio custom resource (CR). For example, the following Istio CR defines a provider named `ext-auth-provider`.

```yaml
apiVersion: operator.kyma-project.io/v1alpha2
kind: Istio
metadata:
  name: default
  namespace: kyma-system
spec:
  config:
    authorizers:
    - headers:
        inCheck:
          include:
          - x-ext-authz
      name: ext-auth-provider
      port: 8000
      service: ext-auth-provider.provider-system.svc.cluster.local
```

Once you define the provider in the Istio configuration, you can reference it in the APIRule CR.

### Securing an Endpoint with **extAuth**
This configuration allows access to the `/get` path of the `user-service` service. Based on this APIRule, a `CUSTOM` Istio AuthorizationPolicy is created with the `ext-auth-provider` provider, securing access.

```yaml
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: ext-authz
  namespace: user-system
spec:
  hosts:
    - na.example.com
  service:
    name: user-service
    port: 8000
  gateway: kyma-system/kyma-gateway
  rules:
    - path: "/get"
      methods: ["GET"]
      extAuth:
        authorizers:
        - x-ext-authz
```

### Securing an Endpoint with **extAuth** and JWT Restrictions

This configuration allows access to the `/get` path of the `user-service` service, as in the example in the previous section. The access is further restricted by the JWT configuration, specified in the `restrictions` field.

```yaml
apiVersion: gateway.kyma-project.io/v2
kind: APIRule
metadata:
  name: ext-authz
  namespace: user-system
spec:
  hosts:
    - na.example.com
  service:
    name: user-service
    port: 8000
  gateway: kyma-system/kyma-gateway
  rules:
    - path: "/get"
      methods: ["GET"]
      extAuth:
        authorizers:
          - x-ext-authz
        restrictions:
          authentications:
            - issuer: https://example.com
              jwksUri: https://example.com/.well-known/jwks.json
          authorizations:
            - audiences: ["app1"]
```