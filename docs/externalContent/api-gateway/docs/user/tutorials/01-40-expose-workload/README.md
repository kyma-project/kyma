# Tutorials - Expose a Workload
Browse the API Gateway tutorials to learn how to expose workloads. The tutorials are available for the following versions of the APIRule custom resource (CR): `v2`, v2alpha1`, and `v1beta1`. 

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

Expose a workload with APIRule in version `v2`:
- [Expose a Workload](./01-40-expose-workload-apigateway.md)
- [Expose Multiple Workloads on the Same Host](./01-41-expose-multiple-workloads.md)
- [Expose Workloads in Multiple Namespaces with a Single APIRule Definition](./01-42-expose-workloads-multiple-namespaces.md)

Expose a workload with APIRule in version `v2alpha1`:
- [Expose a Workload](./v2alpha1/01-40-expose-workload-apigateway.md)
- [Expose Multiple Workloads on the Same Host](./v2alpha1/01-41-expose-multiple-workloads.md)
- [Expose Workloads in Multiple Namespaces with a Single APIRule Definition](./v2alpha1/01-42-expose-workloads-multiple-namespaces.md)

Expose a workload with APIRule in version `v1beta1`:
- [Expose a Workload](./v1beta1-deprecated/01-40-expose-workload-apigateway.md)
- [Expose Multiple Workloads on the Same Host](./v1beta1-deprecated/01-41-expose-multiple-workloads.md)
- [Expose Workloads in Multiple Namespaces with a Single APIRule Definition](./v1beta1-deprecated/01-42-expose-workloads-multiple-namespaces.md)
