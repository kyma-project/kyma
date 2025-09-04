# Tutorials - Expose and Secure a Workload
Browse the API Gateway tutorials to learn how to expose and secure workloads.

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

> [!NOTE] 
> To expose a workload using APIRule in version `v2` or `v2alpha1`, the workload must be a part of the Istio service mesh. See [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection?id=enable-istio-sidecar-proxy-injection).

Expose and secure a workload with APIRule in version `v2`:
- [Get a JSON Web Token (JWT)](./01-51-get-jwt.md)
- [Expose and Secure a Workload with JWT](./01-52-expose-and-secure-workload-jwt.md)
- [Expose and Secure a Workload with extAuth](./01-53-expose-and-secure-workload-ext-auth.md)

Expose and secure a workload with APIRule in version `v2alpha1`:
- [Get a JSON Web Token (JWT)](./01-51-get-jwt.md)
- [Expose and Secure a Workload with JWT](./v2alpha1/01-52-expose-and-secure-workload-jwt.md)
- [Expose and Secure a Workload with extAuth](./v2alpha1/01-53-expose-and-secure-workload-ext-auth.md)

Expose and secure a workload with APIRule in version `v1beta1`:
- [Expose and Secure a Workload with OAuth2](./v1beta1-deprecated/01-50-expose-and-secure-workload-oauth2.md)
- [Get a JSON Web Token (JWT)](./01-51-get-jwt.md)
- [Expose and Secure a Workload with JWT](./v1beta1-deprecated/01-52-expose-and-secure-workload-jwt.md)
- [Expose and Secure a Workload with Istio](./v1beta1-deprecated/01-53-expose-and-secure-workload-istio.md)

[Expose and Secure a Workload with a Certificate](./01-54-expose-and-secure-workload-with-certificate.md)

[Use the XFF Header to Configure IP-Based Access to a Workload](./01-55-ip-based-access-with-xff.md)