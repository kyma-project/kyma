# Issues with APIRules and Service Connection

If you have issues creating an APIRule custom resource (CR) or you've exposed a service but you cannot connect to it, see the troubleshooting guides related to:

- [APIRule CR in version `v2`](./03-00-basic-diagnostics.md)
- [APIRule CR in version `v2alpha1`](./v2alpha1/03-00-basic-diagnostics.md)
- [APIRule CR in version `v1beta1`](./03-00-basic-diagnostics.md)

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