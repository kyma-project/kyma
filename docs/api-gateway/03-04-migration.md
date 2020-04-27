---
title: Migration from Api to APIRule custom resources
type: Details
---


Migration from Api to APIRule custom resources (CRs) is performed automatically by a job that runs during the Kyma upgrade. During this process, the [API Gateway Migrator tool](https://github.com/kyma-project/kyma/blob/master/components/api-gateway-migrator/README.md#api-gateway-migrator) translates the existing Api CRs to APIRule CRs and deletes the original resources.

>**CAUTION:** Migrating resources may result in a temporary downtime of the exposed service. 

During migration, it may turn out that some resource specifications are too complex or fail to meet all the migration requirements. In such a case the process skips them but doesn't break the way existing services are exposed. However, if you want to introduce further changes or remove the Api CR, your actions won't have any effect on how the service is exposed because it will still use the original configuration.  

>**NOTE:** If the migration process skipped the Api resources due to the invalid status or a blacklisted label, migration is not possible.

## Prerequisites

Before the migration process starts, ensure that all Api resources have a status. To do so, fetch all Apis without the status:

```shell script
kubectl get apis --all-namespaces -o json | jq '.items | .[] | select(.status == null)'
```

Receiving no results means that you can perform the upgrade. If, however, you can see any Api resources in the output, recreate each Api using this script:

```shell script
# set variables
export API_NAME={INSERT_API_NAME_HERE}
export API_NAMESPACE={INSERT_API_NAMESPACE_HERE}
# create temporary file
export TMP_FILE=$(mktemp)
# save Api to temporary file
kubectl get api -n ${API_NAMESPACE} ${API_NAME} > ${TMP_FILE}
# remove Api
kubectl delete api -n ${API_NAMESPACE} ${API_NAME}
# remove dependent resources
kubectl delete virtualservice -n ${API_NAMESPACE} ${API_NAME}
kubectl delete policy -n ${API_NAMESPACE} ${API_NAME} --ignore-not-found
# recreate Api from saved file
kubectl apply -f ${TMP_FILE}
```

Once the Apis are recreated check again if all of them have a status. The output should not include any Api resources. You can now proceed with the migration. Here are the steps you need to follow to ensure your services are properly migrated:

1. [Verify the migration outcome](#details-migration-from-api-to-api-rule-custom-resources-verify-the-automatic-migration). 
2. If you can still see any Api CRs in use, use the [manual migration](#details-migration-from-api-to-api-rule-custom-resources-manual-migration) guide to migrate them.

## Verify the automatic migration

Follow these steps to verify if all Api CRs were migrated to APIRule CRs.

1. Once the Kyma upgrade finishes, list the existing Api CRs:

    ```shell script
      kubectl get apis --all-namespaces
    ```

2. As Api CRs are removed after a successful migration, the list shows all resources skipped by the migration job. Fetch the logs from the job's Pod to find out the reasons: 

    ```shell script
    kubectl logs api-gateway-api-migrator-job -n kyma-system
    ```

## Manual migration

This guide shows how you can manually migrate Api CRs to APIRule CRs.

>**NOTE:** Before you start the manual migration process, see the [Api CR](https://kyma-project.io/docs/1.11/components/api-gateway#custom-resource-api-sample-custom-resource) and [APIRule CR](/components/api-gateway#custom-resource-api-rule) documents for the custom resource detailed definition.

Follow these steps:

1. Fetch the Api CRs you want to migrate:

    ```shell script
    kubectl get api {NAME} -n {NAMESPACE} -o yaml
    ```

2. Create an APIRule CR based on the Api CR's specification.

>**NOTE:** Do not copy the [**status**](/components/api-gateway#custom-resource-api-rule-additional-information) parameter from the original Api CR.

| Parameter | Action  |
|-----------|---------|
| **metadata.name**| Specify the name of the API Rule.| 
| **spec.gateway**| Use `kyma-gateway.kyma-system.svc.cluster.local` which is the default Istio Gateway used to expose services.|
| **spec.service.name**| Copy the value of the corresponding parameter of the Api CR.|
| **spec.service.port**| Copy the value of the corresponding parameter of the Api CR.|
| **spec.service.host**| Set the value for **service.host** parameter to any temporary value which includes the domain. For example, if the value for the **hostname** parameter of the Api CR was set to `sample-service.kyma.local`, change it to `temp-sample-service.kyma.local` so that differs from the original value. Make sure other services on your cluster do not use this hostname. |
| **spec.rules**| The **spec.rules** parameter allows you to provide authentication by defining a list of paths along with the authentication methods. The way of handling authentication differs for Api CR and APIRule CR. Consider the section below as a point of reference when defining this part of the resource.|  

See the examples of an Api and APIRule CRs to understand how the configuration settings map from one to another.

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
In this configuration, to access:

* The `/exact/path/to/resource.jpg` path, you need a token issued by `https://auth.kyma.local`.
* Any path starting with `/pref/` you need a token issued by `https://dex.kyma.local`.
* The `/no/auth/needed/resource.html` you don't need any token because the path is excluded for both configurations.
* All other paths you need a token from one of the issuers.

The APIRule CR corresponding to this example would look as follows:

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
>**NOTE:** Path definitions in the APIRule CR cannot overlap, otherwise the requests to such paths are rejected. 

The **spec.rules** parameter corresponds to **spec.authentication** parameter set in the Api CR where you can specify a [list multiple JWT token issuers](https://kyma-project.io/docs/1.11/components/api-gateway/#details-security-specify-multiple-jwt-token-issuers) to allow secured access and use the **triggerRule.excludedPaths** parameter to exclude paths that don't require any authentication. 

If you want to exclude paths in the APIRule CR definition, for example to avoid two different configurations for a single path, you can use a regular expression with a negative lookahead. In the example, such an expression is used in the last path to exclude the paths already handled by other configuration settings from the `/.*` path. 

Use this table to learn how the APIRule CR's path values correspond to the [**triggerRule.excludedPaths**](https://kyma-project.io/docs/1.11/components/api-gateway/#details-security-specify-service-resource-paths-not-secured-with-jwt-authentication) values for the Api CR, and what negative lookahead value you should add to the `/.*` path when defining your APIRule CR.

| Expression type | Description | Sample value | Path value | Negative lookahead value |
|---|---|---|---|---|
|**exact**| no changes | `/exact/path/to/resource.jpg` | `/exact/path/to/resource.jpg` | `(?!/exact/path/to/resource.jpg)` |
|**prefix**| add `.*` at the end | `/pref/` | `/pref/.*` | `(?!/pref/.*)` |
|**suffix**| add `.*` at the beginning; in negative lookahead add `\s` at the end | `/suffix.ico` | `.*/suffix.ico | (?!.*/suffix.ico\s)` |
|**regex**| no changes | `/anything.*` | `/anything.*` | `(?!/anything.*)` |

If the same **excludedPaths** element is present throughout the authentication settings, a particular path doesn't require any authentication. In that case, create a rule with `handler: allow`, so that you don't have to exclude the path using a negative lookahead.
 
3. When the configuration is ready, create the APIRule object and test if the service is working as expected on the new host. It should work the same on both hosts.

4. Remove the VirtualService resource and make sure it is deleted before proceeding:

   ```shell script
     kubectl delete virtualservice {API_NAME} -n {NAMESPACE}
    ```

5. If the Api CR was secured with an authentication mechanism, remove the Policy resource:

    ```shell script
     kubectl delete policy {API_NAME} -n {NAMESPACE}
    ```

6. To make the service return to its original host, set the **spec.service.host** parameter of APIRule CR to the value used by the Api CR.

    ```shell script
    kubectl edit apirule {APIRULE_NAME} -n {NAMESPACE}
    ```

  After saving that configuration, the service should be available on the original hostname.

7. Remove the Api CR:

    ```shell script
    kubectl delete api {API_NAME} -n {NAMESPACE}
    ```
