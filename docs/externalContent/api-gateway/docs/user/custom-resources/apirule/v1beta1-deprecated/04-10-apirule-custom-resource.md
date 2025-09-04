# APIRule v1beta1 Custom Resource <!-- {docsify-ignore-all} -->

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

## Specification of APIRule v1beta1 Custom Resource

This table lists all parameters of APIRule CRD together with their descriptions:

**Spec:**

| Field                           | Mandatory | Description                                                                                                                                                                                                                                                                                                                                  |
|---------------------------------|:---------:|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **gateway**                     |  **YES**  | Specifies the Istio Gateway.                                                                                                                                                                                                                                                                                                                 |
| **corsPolicy**                  |  **NO**   | Allows configuring CORS headers sent with the response. If **corsPolicy** is not defined, the default values are applied. If **corsPolicy** is configured, only the CORS headers defined in the APIRule are sent with the response. For more information, see the [decision record](https://github.com/kyma-project/api-gateway/issues/752). |
| **corsPolicy.allowHeaders**     |  **NO**   | Specifies headers allowed with the **Access-Control-Allow-Headers** CORS header.                                                                                                                                                                                                                                                             |
| **corsPolicy.allowMethods**     |  **NO**   | Specifies methods allowed with the **Access-Control-Allow-Methods** CORS header.                                                                                                                                                                                                                                                             |
| **corsPolicy.allowOrigins**     |  **NO**   | Specifies origins allowed with the **Access-Control-Allow-Origins** CORS header.                                                                                                                                                                                                                                                             |
| **corsPolicy.allowCredentials** |  **NO**   | Specifies whether credentials are allowed in the **Access-Control-Allow-Credentials** CORS header.                                                                                                                                                                                                                                           |
| **corsPolicy.exposeHeaders**    |  **NO**   | Specifies headers exposed with the **Access-Control-Expose-Headers** CORS header.                                                                                                                                                                                                                                                            |
| **corsPolicy.maxAge**           |  **NO**   | Specifies the maximum age of CORS policy cache. The value is provided in the **Access-Control-Max-Age** CORS header. The value type is `duration`, for example, `200s`.                                                                                                                                                                      |
| **host**                        |  **YES**  | Specifies the Service's communication address for inbound external traffic. If only the leftmost label is provided, the default domain name will be used.                                                                                                                                                                                    |
| **service.name**                |  **NO**   | Specifies the name of the exposed Service.                                                                                                                                                                                                                                                                                                   |
| **service.namespace**           |  **NO**   | Specifies the namespace of the exposed Service.                                                                                                                                                                                                                                                                                              |
| **service.port**                |  **NO**   | Specifies the communication port of the exposed Service.                                                                                                                                                                                                                                                                                     |
| **timeout**                     |  **NO**   | Specifies the timeout for HTTP requests in seconds for all Oathkeeper Access Rules. The value can be overridden for each Access Rule. The maximum timeout is limited to 3900 seconds (65 minutes). </br> If no timeout is specified, the default timeout of 180 seconds applies.                                                             |
| **rules**                       |  **YES**  | Specifies the list of Oathkeeper Access Rules.                                                                                                                                                                                                                                                                                               |
| **rules.service**               |  **NO**   | Services definitions at this level have higher precedence than the Service definition at the **spec.service** level.                                                                                                                                                                                                                         |
| **rules.path**                  |  **YES**  | Specifies the path of the exposed Service.                                                                                                                                                                                                                                                                                                   |
| **rules.methods**               |  **NO**   | Specifies the list of HTTP request methods available for **spec.rules.path**. The list of supported methods is defined in [RFC 9910: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html) and [RFC 5789: PATCH Method for HTTP](https://www.rfc-editor.org/rfc/rfc5789.html).                                                        |
| **rules.mutators**              |  **NO**   | Specifies the list of the [Oathkeeper](https://www.ory.sh/docs/oathkeeper/pipeline/mutator) or Istio mutators.                                                                                                                                                                                                                          |
| **rules.accessStrategies**      |  **YES**  | Specifies the list of access strategies. Supported are the [Oathkeeper's](https://www.ory.sh/docs/oathkeeper/pipeline/authn/) `oauth2_introspection`, `jwt`, `noop`, `allow`, and `no_auth`. We also support `jwt` as [Istio](https://istio.io/latest/docs/tasks/security/authorization/authz-jwt/) access strategy.                     |
| **rules.timeout**               |  **NO**   | Specifies the timeout, in seconds, for HTTP requests made to **spec.rules.path**. The maximum timeout is limited to 3900 seconds (65 minutes). Timeout definitions set at this level take precedence over any timeout defined at the **spec.timeout** level.                                                                                 |

> [!WARNING]
>  If `service` is not defined at the **spec.service** level, all defined Access Rules must have `service` defined at the **spec.rules.service** level. Otherwise, the validation fails.

> [!WARNING]
> When you use the Ory handler, do not define the access strategies `noop`, `allow`, or `no_auth` with any other access strategy on the same **spec.rules.path**.
> When you use the Istio handler, do not define the access strategies `jwt`, `noop`, `allow`, or `no_auth` with any other access strategy on the same **spec.rules.path**.
> Additionally, do not use secured access strategies (such as `jwt`, `oauth2_introspection`, `oauth2_client_credentials`, or `cookie_session`) with unsecured access strategies (for example, `allow`, `no_auth`, `noop`, `unauthorized`, or `anonymous`).

**Status:**

When you fetch an existing APIRule CR, the system adds the **status** section which describes the status of the
VirtualService and the Oathkeeper Access Rule created for this CR. The following table lists the fields of the **status** section.

| Field                                | Description                                        |
|:-------------------------------------|:---------------------------------------------------|
| **status.apiRuleStatus**             | Status code describing the APIRule CR.             |
| **status.virtualServiceStatus.code** | Status code describing the VirtualService.         |
| **status.virtualService.desc**       | Current state of the VirtualService.               |
| **status.accessRuleStatus.code**     | Status code describing the Oathkeeper Access Rule. |
| **status.accessRuleStatus.desc**     | Current state of the Oathkeeper Access Rule.       |

**Status codes:**

The following status codes describe VirtualServices and Oathkeeper Access Rules:

| Code        | Description                  |
|-------------|------------------------------|
| **OK**      | Resource created.            |
| **SKIPPED** | Skipped creating a resource. |
| **ERROR**   | Resource not created.        |

## Sample Custom Resource

This is a sample custom resource (CR) that the APIGateway Controller listens for to expose a Service. The following
example has the **rules** section specified, which makes APIGateway Controller create an Oathkeeper Access Rule for the
Service.

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
  timeout: 360
  rules:
    - path: /.*
      methods: [ "GET" ]
      mutators: [ ]
      accessStrategies:
        - handler: oauth2_introspection
          config:
            required_scope: [ "read" ]
```
