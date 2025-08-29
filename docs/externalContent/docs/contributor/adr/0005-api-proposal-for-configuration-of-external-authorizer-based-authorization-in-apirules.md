# API Proposal for Configuration of External Authorizer-Based Authorization in APIRules

## Status
Accepted

## Context

We are going to introduce a new handler for APIRules that allows configuring External Authorizer for services exposed by the user.

The new access strategy will be called `extAuth`,
which follows the camel case naming convention used for previous strategies.
The `no_auth` access strategy will also change its name to `noAuth` in the future.

## Decision

### Should We Support Combining the `extAuth` and `jwt` Access Strategies?

The Authorization Policy that enables External Authorizer uses `action: CUSTOM`.
This allows the External Authorizer handler to be used alongside other handlers,
especially Istio-based JWT. The reason such a configuration is possible is that the `CUSTOM` actions are evaluated independently,
as described in the [Istio Authorization Policy documentation](https://istio.io/latest/docs/reference/config/security/authorization-policy).
By taking advantage of this, the customer could have a setup that performs both authentication with 
OAuth2 Authorization code flow and based on the claims of the presented JWT bearer token.

Additionally, the evaluation is performed based on the matched host and path, meaning that the user can have a configuration
where multiple APIRules with different hosts/gateways are used to reach the same service.

For example, it is possible to use the following configuration:

- An `AuthorizationPolicy` enabling External Authorizer:

```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: ext-authz
spec:
  action: CUSTOM
  provider:
    name: oauth2-proxy
  rules:
  - to:
    - operation:
        hosts:                                                                                         
        - httpbin.local
        paths:
        - /headers
```

- and an `AuthorizationPolicy` restricting the access on a claim-based strategy:

```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: require-claim
spec:
  action: ALLOW
  rules:
  - to:
      - operation:
          hosts:
            - httpbin.local
          paths:
            - /headers
    when:
      - key: request.auth.claims[some_claim]
        values:
          - some_value
```

- and an additional `RequestAuthentication` ensuring that Istio recognizes the issuer:
```yaml
apiVersion: security.istio.io/v1beta1
kind: RequestAuthentication
metadata:
  name: httpbin
spec:
  jwtRules:
  - issuer: https://example.com
    jwksUri: https://example.com/.well-known/jwks.json
```

**Decision**
We decided to enforce that the user must use one and only one **accessStrategy** per every entry in **spec.rules**.
Therefore, we don't support combining both `extAuth` and `jwt` access strategies.
Instead, we would like to enable configuration within the `extAuth` strategy scope and allow the creation of an `ALLOW`
AuthorizationPolicy based on that configuration.
As a result, the proposed API would look as follows:

```yaml
accessStrategy: # Validation: there needs to be one access strategy and only one
  extAuth: # There can be multiple external authorizers configured
    - name: oauth2-proxy
      # Will most likely configure the same fields as in `jwt` access strategy
      authentications:
        - issuer: https://example.com
          jwksUri: https://example.com/.well-known/jwks.json            
      authorizations:
        - audiences: ["app1"]
```

### Should We Support Multiple External Authorizers?

We must consider whether a configuration that uses multiple external authorizers on one path is valuable.
Technically, such configuration would be possible, as all `CUSTOM` policies must result in the `allow`
response for the request to be allowed.

**Decision**
We decided to support multiple External Authorizers. Since there are some useful configurations that our users might want to have.

### API Proposal

We have discussed how the API would be structured to support future versions,
specifically `v2alpha1/v1`. We decided that **accessStrategy** will hold a single entry,
either `extAuth`, `jwt`, or `noAuth`. The user must define only one access strategy in every **rule** of the **spec.rules** field.

See the following sample, which uses the proposed API:

```yaml
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: service-secured
spec:
  gateway: kyma-system/kyma-gateway
  host: httpbin
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /headers
      methods: ["GET"]
      accessStrategy: # Validation: there needs to be one access strategy, and only one
        extAuth: # There can be multiple external authorizers configured
          - name: oauth2-proxy # Validation: Check if there is that authorizer in Istio mesh config
            restrictions:
              authentications:
                - issuer: https://example.com
                  jwksUri: https://example.com/.well-known/jwks.json            
              authorizations:
                 - audiences: ["app1"]
        ## OR
        jwt:
          authentications:
            - issuer: https://example.com
              jwksUri: https://example.com/.well-known/jwks.json            
          authorizations:
            - audiences: ["app1"]
        ## OR
        noAuth: true
```
