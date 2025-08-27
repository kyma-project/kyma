# APIRule v2alpha1 Custom Resource <!-- {docsify-ignore-all} -->

> [!WARNING]
> APIRule CRDs in versions `v1beta1` and `v2alpha1` have been deprecated and will be removed in upcoming releases.
>
> After careful consideration, we have decided that the deletion of `v1beta1` planned for end of May will be postponed. A new target date will be announced in the future.
> 
> **Required action**: Migrate all your APIRule custom resources (CRs) to version `v2`.
> 
> To migrate your APIRule CRs from version `v2alpha1` to version `v2`, you must update the version in APIRule CRs’ metadata.
> 
> To learn how to migrate your APIRule CRs from version `v1beta1` to version `v2`, see [APIRule Migration](../../../apirule-migration/README.md).
> 
> Since the APIRule CRD `v2alpha1` is identical to `v2`, the migration procedure from version `v1beta1` to version `v2` is the same as from version `v1beta1` to version `v2alpha1`.

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) describes the kind and the format of data the
APIGateway Controller listens for. To get the up-to-date CRD in the `yaml` format, run the following command:

```shell
kubectl get crd apirules.gateway.kyma-project.io -o yaml
```

## Specification of APIRule v2alpha1 Custom Resource

This table lists all parameters of APIRule `v2alpha1` CRD together with their descriptions:

**Spec:**

| Field                                            | Required | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | Validation                                                                                                            |
|:-------------------------------------------------|:--------:|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:----------------------------------------------------------------------------------------------------------------------|
| **gateway**                                      | **YES**  | Specifies the Istio Gateway. The value must reference an actual Gateway in the cluster.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | It must be in the `namespace/gateway` format. The namespace and the Gateway cannot be longer than 63 characters each. |
| **corsPolicy**                                   |  **NO**  | Allows configuring CORS headers sent with the response. If **corsPolicy** is not defined, the CORS headers are enforced to be empty.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | None                                                                                                                  |
| **corsPolicy.allowHeaders**                      |  **NO**  | Specifies headers allowed with the **Access-Control-Allow-Headers** CORS header.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | None                                                                                                                  |
| **corsPolicy.allowMethods**                      |  **NO**  | Specifies methods allowed with the **Access-Control-Allow-Methods** CORS header.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | None                                                                                                                  |
| **corsPolicy.allowOrigins**                      |  **NO**  | Specifies origins allowed with the **Access-Control-Allow-Origins** CORS header.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | None                                                                                                                  |
| **corsPolicy.allowCredentials**                  |  **NO**  | Specifies whether credentials are allowed in the **Access-Control-Allow-Credentials** CORS header.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | None                                                                                                                  |
| **corsPolicy.exposeHeaders**                     |  **NO**  | Specifies headers exposed with the **Access-Control-Expose-Headers** CORS header.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | None                                                                                                                  |
| **corsPolicy.maxAge**                            |  **NO**  | Specifies the maximum age of CORS policy cache. The value is provided in the **Access-Control-Max-Age** CORS header.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | None                                                                                                                  |
| **hosts**                                        | **YES**  | Specifies the Service's communication address for inbound external traffic. It must be a RFC 1123 label (short host) or a valid, fully qualified domain name (FQDN) in the following format: at least two domain labels with characters, numbers, or hyphens.                                                                                                                                                                                                                                                                                                                                                                        | Lowercase RFC 1123 label or FQDN format.                                                                              |
| **service.name**                                 |  **NO**  | Specifies the name of the exposed Service.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              | None                                                                                                                  |
| **service.namespace**                            |  **NO**  | Specifies the namespace of the exposed Service.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | None                                                                                                                  |
| **service.port**                                 |  **NO**  | Specifies the communication port of the exposed Service.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | None                                                                                                                  |
| **timeout**                                      |  **NO**  | Specifies the timeout for HTTP requests in seconds for all Access Rules. The value can be overridden for each Access Rule. </br> If no timeout is specified, the default timeout of 180 seconds applies.                                                                                                                                                                                                                                                                                                                                                                                                                | The maximum timeout is limited to 3900 seconds (65 minutes).                                                          |
| **rules**                                        | **YES**  | Specifies the list of Access Rules.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     | None                                                                                                                  |
| **rules.service**                                |  **NO**  | Services definitions at this level have higher precedence than the Service definition at the **spec.service** level.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | None                                                                                                                  |
| **rules.path**                                   | **YES**  | Specifies the path on which the service is exposed. The supported configurations are:<ul><li>Exact path (e.g. `/abc`) - matches the specified path exactly.</li><li>Usage of the `{*}` operator (e.g. `/foo/{*}` or `/foo/{*}/bar`) - matches any request that matches the pattern with exactly one path segment in the operator's place.</li><li>Usage of the `{**}` operator (e.g. `/foo/{**}` or `/foo/{**}/bar`) - matches any request that matches the pattern with zero or more path segments in the operator's place. `{**}` must be the last operator in the path.</li><li>Wildcard path `/*` - matches all paths. It's equivalent to the `/{**}` path.</li></ul>| The value might contain operators `{*}` and/or `{**}`. It can also be a wildcard match `/*`. The order of rules in the APIRule CR is important. Rules defined earlier in the list have a higher priority than those defined later.                                    |
| **rules.methods**                                |  **NO**  | Specifies the list of HTTP request methods available for **spec.rules.path**. The list of supported methods is defined in [RFC 9910: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html) and [RFC 5789: PATCH Method for HTTP](https://www.rfc-editor.org/rfc/rfc5789.html).                                                                                                                                                                                                                                                                                                                                   | None                                                                                                                  |
| **rules.noAuth**                                 |  **NO**  | Setting `noAuth` to `true` disables authorization.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      | Must be set to true if jwt and extAuth are not specified.                                                             |
| **rules.request**                                |  **NO**  | Defines request modification rules, which are applied before forwarding the request to the target workload.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | None                                                                                                                  |
| **rules.request.cookies**                        |  **NO**  | Specifies a list of cookie key-value pairs, that are forwarded inside the **Cookie** header.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | None                                                                                                                  |
| **rules.request.headers**                        |  **NO**  | Specifies a list of header key-value pairs that are forwarded as header=value to the target workload.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | None                                                                                                                  |
| **rules.jwt**                                    |  **NO**  | Specifies the Istio JWT access strategy.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | Must exists if noAuth and extAuth are not specified.                                                                  |
| **rules.jwt.authentications**                    | **YES**  | Specifies the list of authentication objects.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | None                                                                                                                  |
| **rules.jwt.authentications.issuer**             | **YES**  | Identifies the issuer that issued the JWT. <br/>The value must be a URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints.                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | None                                                                                                                  |
| **rules.jwt.authentications.jwksUri**            | **YES**  | Contains the URL of the provider’s public key set to validate the signature of the JWT. <br/>The value must be a URL. Although HTTP is allowed, it is recommended that you use only HTTPS endpoints.                                                                                                                                                                                                                                                                                                                                                                                                                    | None                                                                                                                  |
| **rules.jwt.authentications.fromHeaders**        |  **NO**  | Specifies the list of headers from which the JWT token is extracted.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | None                                                                                                                  |
| **rules.jwt.authentications.fromHeaders.name**   | **YES**  | Specifies the name of the header.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       | None                                                                                                                  |
| **rules.jwt.authentications.fromHeaders.prefix** |  **NO**  | Specifies the prefix used before the JWT token. The default is `Bearer`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | None                                                                                                                  |
| **rules.jwt.authentications.fromParams**         |  **NO**  | Specifies the list of parameters from which the JWT token is extracted.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | None                                                                                                                  |
| **rules.jwt.authorizations**                     |  **NO**  | Specifies the list of authorization objects.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | None                                                                                                                  |
| **rules.jwt.authorizations.requiredScopes**      |  **NO**  | Specifies the list of required scope values for the JWT.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                | None                                                                                                                  |
| **rules.jwt.authorizations.audiences**           |  **NO**  | Specifies the list of audiences required for the JWT.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   | None                                                                                                                  |
| **rules.extAuth**                                |  **NO**  | Specifies the Istio External Authorization access strategy.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             | Must exists if noAuth and jwt are not specified.                                                                      |
| **rules.extAuth.authorizers**                    | **YES**  | Specifies the Istio External Authorization authorizers. In case extAuth is configured, at least one must be present.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    | Validated that the provider exists in Istio external authorization providers.                                         |
| **rules.extAuth.restrictions**                   |  **NO**  | Specifies the Istio External Authorization JWT restrictions. Field configuration is the same as for `rules.jwt`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | None                                                                                                                  |
| **rules.timeout**                                |  **NO**  | Specifies the timeout, in seconds, for HTTP requests made to **spec.rules.path**. Timeout definitions set at this level take precedence over any timeout defined at the **spec.timeout** level.                                                                                                                                                                                                                                                                                                                                                                                                                         | The maximum timeout is limited to 3900 seconds (65 minutes).                                                          |

> [!WARNING]
> When you use an unsupported `v1beta1` configuration in version `v2alpha1` of the APIRule CR, you get an empty **spec**. See [supported access strategies](04-15-api-rule-access-strategies.md).

> [!WARNING]
> The Ory handler is not supported in version `v2alpha1` of the APIRule. When **noAuth** is set to true, **jwt** cannot be defined on the same path.

> [!WARNING]
>  If a service is not defined at the **spec.service** level, all defined Access Rules must have it defined at the **spec.rules.service** level. Otherwise, the validation fails.

> [!WARNING]
>  If a short host name is defined at the **spec.hosts** level, the referenced Gateway must provide the same single host for all [Server](https://istio.io/latest/docs/reference/config/networking/gateway/#Server) definitions and it must be prefixed with `*.`. Otherwise, the validation fails.

**Status:**

The following table lists the fields of the **status** section.

| Field                  | Description                                                                                                                       |
|:-----------------------|:----------------------------------------------------------------------------------------------------------------------------------|
| **status.state**       | Defines the reconciliation state of the APIRule. The possible states are `Ready`, `Warning`, `Error`, `Processing`, or `Deleting`. |
| **status.description** | Contains a detailed description of **status.state**.                                                                                         |

### Significance of Rules Path Order
Operators `{*}` and `{**}` allow you to define a single APIRule that matches multiple request paths.
However, this also introduces the possibility of path conflicts.
A path conflict occurs when two or more APIRule resources match the same path and share at least one common HTTP method. This is why the order of rules is important.

Rules defined earlier in the list have a higher priority than those defined later. Therefore, we recommend defining rules from the most specific path to the most general.

See an example of a valid **rules.path** order, listed from the most specific to the most general:
- `/anything/one`
- `/anything/one/two`
- `/anything/{*}/one`
- `/anything/{*}/one/{**}/two`
- `/anything/{*}/{*}/two`
- `/anything/{**}/two`
- `/anything/`
- `/anything/{**}`
- `/{**}`

Understanding the relationship between paths and methods in a rule is crucial to avoid unexpected behavior. For example, the following APIRule configuration excludes the `POST` and `GET` methods for the path `/anything/one` with `noAuth`. This happens because the rule with the path `/anything/{**}` shares at least one common method (`GET`) with a preceding rule.

```yaml
...
rules:
  - methods:
    - GET
    jwt:
      authentications:
        - issuer: https://example.com
          jwksUri: https://example.com/.well-known/jwks.json
    path: /anything/one
  - methods:
    - GET
    - POST
    noAuth: true
    path: /anything/{**}
```
To use the `POST` method on the path `/anything/one`, you must define separate rules for overlapping methods and paths. See the following example:
```yaml
...
rules:
  - methods:
      - GET
    jwt:
      authentications:
        - issuer: https://example.com
          jwksUri: https://example.com/.well-known/jwks.json
    path: /anything/one
  - methods:
      - GET
    noAuth: true
    path: /anything/{**}
  - methods:
      - POST
    noAuth: true
    path: /anything/{**}
```

## APIRule CR's State

|     Code     | Description                                                                                                                                                                                                         |
|:------------:|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|   `Ready`    | APIRule Controller finished reconciliation.                                                                                                                                                                                 |
| `Processing` | APIRule Controller is reconciling resources.                                                                                                                                                                                |
|  `Deleting`  | APIRule Controller is deleting resources.                                                                                                                                                                                   |
|   `Error`    | An error occurred during the reconciliation. The error is rather related to the API Gateway module than the configuration of your resources.                                                                        |
|  `Warning`   | An issue occurred during the reconciliation that requires your attention. Check the status.description message to identify the issue and make the necessary corrections to the APIRule CR or any related resources. |

## Sample Custom Resource

See an exemplary APIRule custom resource:

```yaml
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: service-exposed
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - foo.bar
  service:
    name: foo-service
    namespace: foo-namespace
    port: 8080
  timeout: 360
  rules:
    - path: /*
      methods: [ "GET" ]
      noAuth: true
```

This sample APIRule illustrates the usage of a short host name. It uses the domain from the referenced Gateway `kyma-system/kyma-gateway`:

```yaml
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: service-exposed
spec:
  gateway: kyma-system/kyma-gateway
  hosts:
    - foo
  service:
    name: foo-service
    namespace: foo-namespace
    port: 8080
  timeout: 360
  rules:
    - path: /*
      methods: [ "GET" ]
      noAuth: true
```
