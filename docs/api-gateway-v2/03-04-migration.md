---
title: Migration from the previous Api resources
type: Details
---

The migration is done automatically by a job. During the migration the old `Api` object is being migrated to an `ApiRule` object, what may cause temporary unavailability of the exposed service. After the successful migration the old resource is removed. The migration might be skipped for some Api configuration, in which case the manual migration process may be done. It won't break the existing service exposure, but any changes to the Api object won't be effective.

List remaining apis to check whether any Api migration has been skipped

```shell script
kubectl get apis --all-namespaces
```

As api resources are removed after a successful migration, all the resources that are left should be considered as not migrated. To get more details fetch logs from the migrator job

```shell script
kubectl logs api-gateway-api-migrator-job -n kyma-system
```

You will find logs saying which api was not migrated and a reason for that.

If the Api was not migrated due to the invalid status or a blacklisted label, you either shouldn't or will not be able to do the migration.

## The manual migration process

The manual migration process consists of the following steps:

1. Get Api resource you want to migrate. It would be helpful to have it open during the whole migration process.

```shell script
kubectl get api <NAME> -n <NAMESPACE>
```

2. Create ApiRule resource with a different host

Here is the documentation for both custom resources, which may be useful during the migration:

https://kyma-project.io/docs/1.11/components/api-gateway/#custom-resource-api-sample-custom-resource
https://kyma-project.io/docs/1.11/components/api-gateway-v2#custom-resource-api-rule

Only the `spec` part of the CRD will be described. `metadata` field can be adjusted in any way and `status` field should not be copied.

Starting from the `service` field of ApiRule, copy `name` and `port` fields from the `service` field of `Api`. Please set `host` value to any temporary value including the domain. For example if the `hostname` value of Api was set to `sample-service.kyma.local` you can change it to `temp-sample-service.kyma.local`. Just make sure that the hostname is not used by other services on your cluster. That value will be changed in later step, but for now it needs to be different than the original value. Set `gateway` field of ApiRule to the Istio Gateway used to expose services. By default it will be `kyma-gateway.kyma-system.svc.cluster.local`.

The configuration of `rules` field of ApiRule is more complex and depends on the `authentication` configuration of ApiRule. As all the basic scenarios are covered by automatic migration, this explanation will only concern the configurations that are not handled automatically.

The basic difference between Apis and ApiRules authentication configuration is that while Api allows to enable the authentication and disable it on specific paths, the ApiRule has an approach where you specifies what authentication should be used per specific path (including the possibility to set it for all paths) and the paths might not cross each other. Another important difference is that Api supports a list of issuers and jwks URIs, but excluded paths are set independently on both. It means that in the example below to access `/exact/path/to/resource.jpg` path the token issued from `https://auth.kyma.local` is required, to access any path starting with `/pref/` the token issued from `https://dex.kyma.local` is required, to access the `/no/auth/needed/resource.html` no token is required because it is excluded for both settings and to access all other paths the token from one of the issuers is required.

```yaml
authentication:
- type: JWT
  jwt:
    issuer: https://dex.kyma.local
    jwksUri: http://dex-service.kyma-system.svc.cluster.local:5556/keys
    triggerRule:
      excludedPaths:
      - exact: /exact/path/to/resource.jpg
      - exact: /no/auth/needed/resource.html
- type: JWT
  jwt:
    issuer: https://auth.kyma.local
    jwksUri: http://auth-service.kyma-system.svc.cluster.local:5556/keys
    triggerRule:
      excludedPaths:
      - prefix: /pref/
      - exact: /no/auth/needed/resource.html
```

Here is the ApiRule configuration which enforces the same authentication policies:

```yaml
rules:
  - path: /no/auth/needed/resource.html
    methods: ["GET", "POST", "PUT", "DELETE"]
    accessStrategies:
    - handler: allow
  - path: /pref/.*
    methods: ["GET", "POST", "PUT", "DELETE"]
    accessStrategies:
    - handler: jwt
      config:
        trusted_issuers:
        - "https://dex.kyma.local"
        jwks_urls:
        - "http://dex-service.kyma-system.svc.cluster.local:5556/keys"
  - path: /exact/path/to/resource.jpg
    methods: ["GET", "POST", "PUT", "DELETE"]
    accessStrategies:
    - handler: jwt
      config:
        trusted_issuers:
        - "https://auth.kyma.local"
        jwks_urls:
        - "http://auth-service.kyma-system.svc.cluster.local:5556/keys"
  - path: (?!/pref/.*)(?!/exact/path/to/resource.jpg)/.*
    methods: ["GET", "POST", "PUT", "DELETE"]
    accessStrategies:
    - handler: jwt
      config:
        trusted_issuers:
        - "https://dex.kyma.local"
        - "https://auth.kyma.local"
        jwks_urls:
        - "http://dex-service.kyma-system.svc.cluster.local:5556/keys"
        - "http://auth-service.kyma-system.svc.cluster.local:5556/keys"
```

The important part to notice is the last path configuration containing regex with a negative lookahead. It is used to exclude paths handled by other path settings from the `/.*` path, as there can't be two different configuration for a single path. There is one exception to that, which applies only if the same `excludedPaths` element is present on all `authentication` settings, so the specific path doesn't require any authentication at all. In that case a rule with `handler: allow` should be created, and the path doesn't have to be excluded using a negative lookahead.

Below is the list showing how ApiRule path value corresponds to the excludedPaths values from Api resource, and what negative lookahead value should be added to the `/.*` path:

| excludedPaths expression type | description | excludedPaths sample value | path value | negative lookahead value |
|---|---|---|---|---|
|**exact**| no changes to be made | /exact/path/to/resource.jpg | /exact/path/to/resource.jpg | (?!/exact/path/to/resource.jpg) |
|**prefix**| add `.*` to the end | /pref/ | /pref/.* | (?!/pref/.*) |
|**suffix**| add `.*` to the beginning; in negative lookahead add `\s` to the end | /suffix.ico | .*/suffix.ico | (?!.*/suffix.ico\s) |
|**regex**| no changes to be made | /anything.* | /anything.* | (?!/anything.*) |

When the configuration is ready please create the ApiRule object and test if the service is working as expected on the new host. It should work exactly the same on both hosts.

3. Remove Api resource and dependant resources

```shell script
kubectl delete api <API_NAME> -n <NAMESPACE>
kubectl delete virtualservice <API_NAME> -n <NAMESPACE>
```

Please make sure that the virtualservice resource is deleted before proceeding to the next instruction.

If the api was secured with authentication mechanism delete the Policy resource:

```shell script
kubectl delete policy <API_NAME> -n <NAMESPACE>
```

4. Edit host of ApiRule

To make the service return to its original host, edit the `spec.service.host` field of ApiRule to an original value.

```shell script
kubectl edit apirule <APIRULE_NAME> -n <NAMESPACE>
```

After saving that configuration, the service should be available on the original hostname.