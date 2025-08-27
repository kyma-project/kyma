# Changes Introduced in APIRule v2alpha1 and v2

Learn about the changes that APIRule v2 introduces and the actions you must take to adjust your `v1beta1` resources. Since version `v2alpha1` is identical to the stable version `v2`, you must consider these changes when migrating either to version `v2` or `v2alpha1`.

See the changes introduced in the new versions:
- [A Workload Must Be in the Istio Service Mesh](#a-workload-must-be-in-the-istio-service-mesh)
- [Internal Traffic to Workloads Is Blocked by Default](#internal-traffic-to-workloads-is-blocked-by-default)
- [CORS Policy Is Not Applied by Default](#cors-policy-is-not-applied-by-default)
- [Path Specification Must Not Contain Regexp](#path-specification-must-not-contain-regexp)
- [JWT Configuration Requires Explicit Issuer URL](#jwt-configuration-requires-explicit-issuer-url)
- [Removed Support for Oathkeeper OAuth2 Handlers](#removed-support-for-oathkeeper-oauth2-handlers)
- [Removed Support for Oathkeeper Mutators](#removed-support-for-oathkeeper-mutators)
- [Removed Support for Opaque Tokens](#removed-support-for-opaque-tokens)

> [!WARNING]
> APIRule CRDs in versions `v1beta1` and `v2alpha1` have been deprecated and will be removed in upcoming releases.
>
> After careful consideration, we have decided that the deletion of `v1beta1` planned for end of May will be postponed. A new target date will be announced in the future.
> 
> **Required action**: Migrate all your APIRule custom resources (CRs) to version `v2`.
> 
> To migrate your APIRule CRs from version `v2alpha1` to version `v2`, you must update the version in APIRule CRs’ metadata.
> 
> To learn how to migrate your APIRule CRs from version `v1beta1` to version `v2`, see [APIRule Migration](../../apirule-migration/README.md). 
> 
> Since the APIRule CRD `v2alpha1` is identical to `v2`, the migration procedure from version `v1beta1` to version `v2` is the same as from version `v1beta1` to version `v2alpha1`.

## A Workload Must Be in the Istio Service Mesh

To use APIRules in versions `v2` or `v2alpha1`, the workload that an APIRule exposes must be in the Istio service mesh. If the workload is not inside the Istio service mesh, the APIRule does not work as expected.

**Required action**: To add a workload to the Istio service mesh, [enable Istio sidecar proxy injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection).

## Internal Traffic to Workloads Is Blocked by Default

By default, access to the workload from internal traffic is blocked if APIRule CR in versions `v2` or `v2alpha1` is applied. This approach aligns with Kyma's "secure by default" principle. 
## CORS Policy Is Not Applied by Default

Version `v1beta1` applied the following CORS configuration by default:
```yaml
corsPolicy:
  allowOrigins: ["*"]
  allowMethods: ["GET", "POST", "PUT", "DELETE", "PATCH"]
  allowHeaders: ["Authorization", "Content-Type", "*"]
```

Versions `v2` and `v2alpha1` do not apply these default values. If the **corsPolicy** field is empty, the CORS configuration is not applied. For more information, see [architecture decision record #752](https://github.com/kyma-project/api-gateway/issues/752).

**Required action**: Configure CORS policy in the **corsPolicy** field.
> [!NOTE]
> Since not all APIs require the same level of access, adjust your CORS policy configuration according to your application's specific needs and security requirements.

If you decide to use the default CORS values defined in the APIRule `v1beta1`, you must explicitly define them in your APIRule `v2`. For preflight requests to work, you must explicitly allow the `"OPTIONS"` method in the **rules.methods** field of your APIRule.

## Path Specification Must Not Contain Regexp

APIRule in versions `v2` and `v2alpha1` does not support regexp in the **spec.rules.path** field of APIRule CR. Instead, it supports the use of the `{*}` and `{**}` operators. See the supported configurations:
- Use the exact path (for example, `/abc`). It matches the specified path exactly.
- Use the `{*}` operator (for example, `/foo/{*}` or `/foo/{*}/bar`).  This operator represents any request that matches the given pattern, with exactly one path segment replacing the operator.
- Use the `{**}` operator (for example, `/foo/{**}` or `/foo/{**}/bar`). This operator represents any request that matches the pattern with zero or more path segments in the operator’s place. It must be the last operator in the path.
- Use the wildcard path `/*`, which matches all paths. It’s equivalent to the `/{**}` path. If your configuration in APIRule `v1beta1` used such a path as `/foo(.*)`, when migrating to the new versions, you must define configurations for two separate paths: `/foo` and `/foo/{**}`.



> [!NOTE] The order of rules in the APIRule CR is important. Rules defined earlier in the list have a higher priority than those defined later. Therefore, we recommend defining rules from the most specific path to the most general.
> 
> Operators allow you to define a single APIRule that matches multiple request paths. However, this also introduces the possibility of path conflicts. A path conflict occurs when two or more APIRule resources match the same path and share at least one common HTTP method. This is why the order of rules is important.


For more information on the APIRule specification, see [APIRule v2alpha1 Custom Resource](https://kyma-project.io/#/api-gateway/user/custom-resources/apirule/v2alpha1/04-10-apirule-custom-resource) and [APIRule v2 Custom Resource](https://kyma-project.io/#/api-gateway/user/custom-resources/apirule/04-10-apirule-custom-resource).

**Required action**: Replace regexp expressions in the **spec.rules.path** field of your APIRule CRs with the `{*}` and `{**}` operators.

## JWT Configuration Requires Explicit Issuer URL

Versions `v2` and `v2alpha1` of APIRule introduce an additional mandatory configuration filed for JWT-based authorization - **issuer**. You must provide an explicit issuer URL in the APIRule CR. See an example configuration:

```yaml
rules:
- jwt:
    authentications:
        -   issuer: {YOUR_ISSUER_URL}
            jwksUri: {YOUR_JWKS_URI}
```
If you use Cloud Identity Services, you can find the issuer URL in the OIDC well-known configuration at `https://{YOUR_TENANT}.accounts.ondemand.com/.well-known/openid-configuration`.

**Required action**: Add the **issuer** field to your APIRule specification. For more information, see [Migrating APIRule `v1beta1` of Type **jwt** to Version `v2`](../../apirule-migration/01-83-migrate-jwt-v1beta1-to-v2.md).

## Removed Support for Oathkeeper OAuth2 Handlers
The APIRule CR in versions `v2` and `v2alpha1` does not support Oathkeeper OAuth2 handlers. Instead, it introduces the **extAuth** field, which you can use to configure an external authorizer.

**Required action**: Migrate your Oathkeeper-based OAuth2 handlers to use an external authorizer. To learn how to do this, see [SAP BTP, Kyma runtime: APIRule migration - Ory Oathkeeper-based OAuth2 handlers](https://community.sap.com/t5/technology-blogs-by-sap/sap-btp-kyma-runtime-apirule-migration-ory-oathkeeper-based-oauth2-handlers/ba-p/13896184) and [Configuration of the extAuth Access Strategy](https://kyma-project.io/#/api-gateway/user/custom-resources/apirule/v2alpha1/04-15-api-rule-access-strategies).

## Removed Support for Oathkeeper Mutators
The APIRule CR in versions `v2` and `v2alpha1` does not support Oathkeeper mutators. Request mutators are replaced with request modifiers defined in the **spec.rule.request** section of the APIRule CR. This section contains the request modification rules applied before the request is forwarded to the target workload. Token mutators are not supported in APIRules `v2` and `v2alpha1`. For that, you must define your own **extAuth** configuration.

**Required action**: Migrate your rules that rely on Oathkeeper mutators to use request modifiers or an external authorizer. For more information, see [Configuration of the extAuth Access Strategy](https://kyma-project.io/#/api-gateway/user/custom-resources/apirule/v2alpha1/04-15-api-rule-access-strategies) and [APIRule v2alpha1 Custom Resource](https://kyma-project.io/#/api-gateway/user/custom-resources/apirule/v2alpha1/04-10-apirule-custom-resource).

## Removed Support for Opaque Tokens

The APIRule CR in versions `v2` and `v2alpha1` does not support the usage of Opaque tokens. Instead, it introduces the **extAuth** field, which you can use to configure an external authorizer.

**Required action**: Migrate your rules that use Opaque tokens to use an external authorizer. For more information, see [Configuration of the extAuth Access Strategy](https://kyma-project.io/#/api-gateway/user/custom-resources/apirule/v2alpha1/04-15-api-rule-access-strategies).
