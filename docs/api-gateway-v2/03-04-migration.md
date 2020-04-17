---
title: Migration from the previous Api resources
type: Details
---

The migration is done automatically by a job, which runs during the Kyma upgrade. During the migration, an old Api object is being translated to an APIRule object, which may result in a temporary downtime of the exposed service. The original resource is deleted as a part of the migration process. The tool used by the job is described [in this document](https://github.com/kyma-project/kyma/blob/master/components/api-gateway-migrator/README.md#api-gateway-migrator).

>**CAUTION:** Some Api configurations are too complex or do not meet all requirements to be automatically migrated. To make sure all your services are exposed via APIRules, [verify the outcome](#verify-automatic-migration) of the migration job and follow the [manual migration process](#manual-migration) guide if necessary. Skipping the migration won't break the existing service exposure, but any further changes or removal of the Api resource won't affect how the service is exposed - the original configuration will be used.

## Verify automatic migration

List the existing Api objects after the Kyma upgrade is finished. Run this command:

```shell script
kubectl get apis --all-namespaces
```

As Api objects are removed after a successful migration, all the listed resources are considered not migrated. Fetch logs from the job's pod to learn the reason for skipping the migration. 

```shell script
kubectl logs api-gateway-api-migrator-job -n kyma-system
```


If the Api was not migrated due to the invalid status or a blacklisted label, you either shouldn't or will not be able to do the migration.

## Manual migration

The manual migration process consists of the following steps:

1. Fetch the Api resource you want to migrate:

```shell script
kubectl get api <NAME> -n <NAMESPACE> -o yaml
```

2. Create an APIRule resource based on the original Api object

> **TIP:** Here is the documentation for both custom resources, which may be useful during the migration: 
> * [Api CR](https://kyma-project.io/docs/1.11/components/api-gateway/#custom-resource-api-sample-custom-resource), 
> * [APIRule CR](https://kyma-project.io/docs/1.11/components/api-gateway-v2#custom-resource-api-rule).

Only the `spec` part of the CRD will be described. `metadata` field can be adjusted in any way and `status` field should not be copied.

Starting from the `service` field of APIRule:
 * Copy `name` and `port` fields from the `service` field of Api object. 
 * Set `host` value to any temporary value including the domain. For example, if the `hostname` value of Api was set to `sample-service.kyma.local` you can change it to `temp-sample-service.kyma.local`. Just make sure that the hostname is not used by other services on your cluster. That value will be changed in a later step, but for now it needs to be different than the original value. 
 * Set `gateway` field of APIRule to the Istio Gateway used to expose services. By default it will be `kyma-gateway.kyma-system.svc.cluster.local`.

The configuration of the `rules` field of APIRule is more complex and depends on the `authentication` configuration of Api. As all the basic scenarios are covered by automatic migration, this explanation will only concern the configurations that are not handled automatically.

The most important difference between Apis and APIRules authentication configuration is that Api is configured with a list of JWT issuers set for all paths of a service and APIRule is configured with a list of per path definitions.
In the Api you can define a list of excluded paths for an issuer. Requests for these paths will not require authentication by this issuer but will require authentication from other configured issuers.
The APIRule has an approach where you specify what authentication should be used per the path definition.
Because paths are defined using regular expressions, you can define an expression that matches every path, or a subset of possible paths, or a single one.
Remember that path definitions in the APIRule must not overlap - if they do, requests to such paths are rejected. That's why translating an Api with multiple issuers having different excluded paths requires sophisticated regex definitions in a corresponding APIRule.

In the example below:
 * to access `/exact/path/to/resource.jpg` path the token issued from `https://auth.kyma.local` is required, 
 * to access any path starting with `/pref/` the token issued from `https://dex.kyma.local` is required, 
 * to access the `/no/auth/needed/resource.html` no token is required because it is excluded for both settings, 
 * to access all other paths the token from one of the issuers is required.

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

Note that the last path configuration contains a regular expression with a negative lookahead. It is used to exclude paths handled by other path settings from the `/.*` path, as there can't be two different configurations for a single path. There is one exception to that, which applies only if the same `excludedPaths` element is present on all `authentication` settings, so the specific path doesn't require any authentication at all. In that case, a rule with `handler: allow` should be created, and the path doesn't have to be excluded using a negative lookahead.

Below is the list showing how ApiRule path value corresponds to the excludedPaths values from Api resource, and what negative lookahead value should be added to the `/.*` path:

| excludedPaths expression type | description | excludedPaths sample value | path value | negative lookahead value |
|---|---|---|---|---|
|**exact**| no changes to be made | /exact/path/to/resource.jpg | /exact/path/to/resource.jpg | (?!/exact/path/to/resource.jpg) |
|**prefix**| add `.*` to the end | /pref/ | /pref/.* | (?!/pref/.*) |
|**suffix**| add `.*` to the beginning; in negative lookahead add `\s` to the end | /suffix.ico | .*/suffix.ico | (?!.*/suffix.ico\s) |
|**regex**| no changes to be made | /anything.* | /anything.* | (?!/anything.*) |

When the configuration is ready, create the APIRule object and test if the service is working as expected on the new host. It should work the same on both hosts.

3. Remove dependent resources of an Api

```shell script
kubectl delete virtualservice <API_NAME> -n <NAMESPACE>
```

Make sure that the Virtual Service resource is deleted before proceeding to the next instruction.

If the api was secured with an authentication mechanism delete the Policy resource:

```shell script
kubectl delete policy <API_NAME> -n <NAMESPACE>
```

4. Edit host of APIRule

To make the service return to its original host, edit the `spec.service.host` field of APIRule to an original value.

```shell script
kubectl edit apirule <APIRULE_NAME> -n <NAMESPACE>
```

After saving that configuration, the service should be available on the original hostname.

5. Remove Api resource

```shell script
kubectl delete api <API_NAME> -n <NAMESPACE>
```