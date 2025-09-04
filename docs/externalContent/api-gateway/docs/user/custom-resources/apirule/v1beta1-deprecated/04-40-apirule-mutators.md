# APIRule Mutators

You can use mutators to enrich an incoming request with information. Different types of mutators are supported depending on the access strategy you use:

| Access Strategy        | Mutator support                                                           |
|:-----------------------|:--------------------------------------------------------------------------|
| `jwt`                  | Istio `cookie` and `header` mutators                                      |
| `oauth2_introspection` | [Oathkeeper](https://www.ory.sh/docs/oathkeeper/pipeline/mutator) mutator |
| `noop`                 | [Oathkeeper](https://www.ory.sh/docs/oathkeeper/pipeline/mutator) mutator |
| `allow`                | No mutators supported                                                     |
| `no_auth`              | No mutators supported                                                     |

This document explains and provides examples of Istio mutators compatible with the JWT access strategy. Additionally, it explores the possibility of using Oathkeeper mutators with Istio and provides guidance on configuring them.

## Istio Mutators
The `cookie` and `header` mutators are supported in combination with the JWT access strategy. You are allowed to configure multiple mutators for one APIRule, but only one mutator of each type is allowed.

### Header Mutator
The headers are defined in the **headers** field of the header mutator configuration. The keys represent the names of the headers, and each value is a string. You can use [Envoy command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) in the header value to perform operations such as copying an incoming header to a new one. The configured headers are applied to the request. They overwrite any existing headers with the same name.

#### Example

In the following example, two different headers are configured: **X-Custom-Auth**, which uses the incoming Authorization header as a value, and **X-Some-Data** with the value `some-data`.

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
              X-Custom-Auth: "%REQ(Authorization)%"
              X-Some-Data: "some-data"
      accessStrategies:
        - handler: jwt
          config:
            authentications:
              - issuer: $ISSUER
                jwksUri: $JWKS_URI
```

### Cookie Mutator
To configure cookies, use the **cookies** mutator configuration field. The keys represent the names of the cookies, and each value is a string. As the cookie value, you can use [Envoy command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators). The configured cookies are set as the `cookie` header in the request and overwrite any existing cookies.

#### Example

The following APIRule example has a new cookie added with the name `some-data` and the value `data`.

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
              some-data: "data"
      accessStrategies:
        - handler: jwt
          config:
            authentications:
              - issuer: $ISSUER
                jwksUri: $JWKS_URI
```

## Support for Ory Oathkeeper Mutators with Istio

### Templating Support

The Ory Access Rules have support for templating in mutators. Simple cases can be implemented with the help of EnvoyFilter, such as the one presented in [id-token-envoyfilter](../id-token-envoyfilter/) directory.

### Header Mutator

To handle the `header` type mutator in Istio, you can use the VirtualService [HeaderOperations](https://istio.io/latest/docs/reference/config/networking/virtual-service/#Headers-HeaderOperations). HeaderOperations only allow you to add static data. However, [Ory Headers Mutator](https://www.ory.sh/docs/oathkeeper/pipeline/mutator#headers) supports templating, which receives the current AuthenticationSession. To support similar capabilities in Istio, you must use [EnvoyFilter](https://istio.io/latest/docs/reference/config/networking/envoy-filter/).

Ory configuration:

```yaml
...
mutators:
  - config:
      headers:
        X-Some-Arbitrary-Data: "test"
    handler: header
...
```

Corresponding Istio Virtual Service configuration:

```yaml
spec:
  http:
    - headers:
        request:
          set:
            X-Some-Arbitrary-Data: "test"
```

### Cookie Mutator

The mutator of type `cookie` can be handled the same as the `header` mutator with Istio using Virtual Service HeaderOperations. Like with the `header` mutator, Istio has the limitation of only allowing static data. However, the [Ory cookie mutator](https://www.ory.sh/docs/oathkeeper/pipeline/mutator#cookie) supports templating.

Ory configuration:

```yaml
...
mutators:
  - config:
      cookies:
        user: "test"
    handler: cookie
...
```

Corresponding Istio Virtual Service configuration:

```yaml
spec:
  http:
    - headers:
        request:
          set:
            Cookie: "user=test"
```

### Id_token Mutator

The functionality of the `id_token` mutator cannot be supported because it would require a mechanism for encoding and signing the response from the OAuth2 server, such as Ory Hydra, into a JWT. 
Additionally, the JWKS used for signing this JWT is deployed as the `ory-oathkeeper-jwks-secret` Secret, which would have to be fetched in the implementation context or mounted into the component responsible for the encoding.

### Hydrator Mutator

Support for the Hydrator token would require to call external APIs in the context of the Istio proxy. This mutator also influences other mutators as it runs before them and supplies them with the outcome of its execution.
