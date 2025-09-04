# Migrate APIRule from Version `v1beta1` to Version `v2`
Learn how to obtain the full **spec** of an APIRule in version `v1beta1` and migrate it to version `v2`. 

To identify which APIRules must be migrated, run the following command:
```bash
kubectl get apirules.gateway.kyma-project.io -A -o json | jq '.items[] | select(.metadata.annotations["gateway.kyma-project.io/original-version"] == "v1beta1") | {namespace: .metadata.namespace, name: .metadata.name}'
```


To obtain the complete **spec** with the **rules** field of an APIRule in version `v1beta1`, see [Retrieve the Complete **spec** of an APIRule in Version `v1beta1`](./01-81-retrieve-v1beta1-spec.md).


To migrate an APIRule from version `v1beta1` to version `v2`, see:
- [Migrate APIRule `v1beta1` of Type **noop**, **allow** or **no_auth** to Version `v2`](./01-82-migrate-allow-noop-no_auth-v1beta1-to-v2.md)
- [Migrate APIRule `v1beta1` of Type **jwt** to Version `v2`](./01-83-migrate-jwt-v1beta1-to-v2.md)
- [Migrate APIRule `v1beta1` of Type **oauth2_introspection** to Version `v2`](./01-84-migrate-oauth2-v1beta1-to-v2.md)

For more information about APIRule `v2`, see also:
- [APIRule `v2` Custom Resource](../custom-resources/apirule/04-10-apirule-custom-resource.md)
- [Changes Introduced in APIRule `v2`](../custom-resources/apirule/04-70-changes-in-apirule-v2.md)