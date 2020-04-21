---
title: Migration from API resources to APIRules
type: Details
---

The migration is done automatically by a job that runs during the Kyma upgrade. During this process, the [API Gateway Migrator tool](https://github.com/kyma-project/kyma/blob/master/components/api-gateway-migrator/README.md#api-gateway-migrator) translates existing API resources to APIRule objects and deletes the original resource.


>**CAUTION:** Some API configurations are too complex or do not meet all requirements to pass the automatic migration. To ensure all your services are exposed using APIRules, [verify the migration outcome](#verify-automatic-migration). If there still are API resources in use, follow [this](#manual-migration) guide to migrate them.


Skipping some resources during migration will not break the way existing services are exposed. However, bear in mind that if you want to introduce changes or remove the API resource, it won't have any effect on how the service is exposed because the service will still use the original configuration. 


## Verify automatic migration

Follow these steps to verify if all the API resources were migrated to APIRules:

1. Once the Kyma upgrade finishes, list the existing API objects:

    ```shell script
      kubectl get apis --all-namespaces
    ```

2. As API objects are removed after a successful migration, the list shows all resources that the migration job skipped. Fetch the logs from the job's Pod to learn the reasons for this behavior: 

    ```shell script
    kubectl logs api-gateway-api-migrator-job -n kyma-system
    ```

>**NOTE:** If the migration process skipped the API resources due to the invalid status or a blacklisted label, you either should not or will not be able to migrate them.

## Manual migration

This guide shows how you can manually migrate API resources to API Rules.

>**NOTE:** Before you start the manual migration process, see the [API CR](/components/api-gateway/#custom-resource-api-sample-custom-resource) and [APIRule CR](/components/api-gateway-v2#custom-resource-api-rule) documents for the custom resource detailed definition.

Follow these steps:

1. Fetch the API resource you want to migrate:

    ```shell script
    kubectl get api {NAME} -n {NAMESPACE} -o yaml
    ```

2. Create an APIRule resource based on the API object's specification:

>**NOTE:** This step focuses on the`spec` part of the CRD. You can `metadata` field in a preferred way and do not copy the [`status`](components/api-gateway-v2#custom-resource-api-rule-additional-information) field from the original API resource.

a. Replace the default value (`kyma-gateway.kyma-system.svc.cluster.local`) for the **gateway** parameter with Istio Gateway used to expose services. 

b. Copy the values for **service.name** and **service.port** parameters from the corresponding fields of the API object. 

c. Set the value for **service.host** parameter to any temporary value which includes the domain. For example, if the value for the **hostname** parameter of the API object was set to `sample-service.kyma.local`, change it to `temp-sample-service.kyma.local` so that differs from the original value. Make sure other services on your cluster do not use this hostname. 

d. Configure the **rules** parameter of APIRule.

>**NOTE:** For details on authentication configuration and possible values you can use, see [this](#authentication-of-api-resources-and-apirules) section.

When the configuration is ready, create the APIRule object and test if the service is working as expected on the new host. It should work the same on both hosts.


3. Remove the dependent resources of the API object:

   ```shell script
     kubectl delete virtualservice {API_NAME} -n {NAMESPACE}
    ```
Make sure that the Virtual Service resource is deleted before proceeding.

4. If the API resource was secured with an authentication mechanism, delete the Policy resource:

    ```shell script
     kubectl delete policy {API_NAME} -n {NAMESPACE}
    ```

5. To make the service returns to its original host, set the **spec.service.host** parameter of APIRule to the value used by the API resource.

    ```shell script
    kubectl edit apirule {APIRULE_NAME} -n {NAMESPACE}
    ```

After saving that configuration, the service should be available on the original hostname.

6. Remove API resource

    ```shell script
    kubectl delete api {API_NAME} -n {NAMESPACE}
    ```

## Authentication of API resources and APIRules

To properly configure the authentication mechanism for the new APIRule, learn more about the differences in authentication configuration for API resources and APIRules:

| API resource | APIRule| 
|--------------| -------|
|Configured with a list of JWT issuers set for all paths of a service.| Configured with a list of path definitions along with their respective authentication methods.|
| You can define a list of excluded paths for a given issuer. Requests for these paths won't require authentication by this issuer but will require authentication from other configured issuers.| You can specify what authentication to use per path definition and use a regular expression to match every path, or a subset of possible paths, or a single one. |


>**NOTE:** Path definitions in the APIRule must not overlap, otherwise the requests to such paths are rejected. That's why migrating an API resource that has multiple issuers and different excluded paths often requires providing complex regex definitions for a corresponding APIRule.

### Authentication configuration 

<div tabs>
  <details>
  <summary>
  Authentication configuration for API objects
 </summary>

 This example shows the authentication configuration for an API object. 

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
In this configuration:

 * To access the `/exact/path/to/resource.jpg` path, you need a token issued by `https://auth.kyma.local`.
 * To access any path starting with `/pref/` you need a token issued by `https://dex.kyma.local`. 
 * To access the `/no/auth/needed/resource.html` you don't need any token because the path is excluded for both configurations. 
 * To access all other paths you need a token from one of the issuers.

  </details>
  <details>
  <summary>
  Authentication configuration for APIRules
  </summary>

This APIRule configuration enforces the same authentication policies as for the API object:

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
The value for the last **rules.path** parameter is a regular expression with a negative lookahead. Its purpose is to exclude the paths already handled by other configuration settings from the `/.*` path to avoid two different configurations for a single path. 
However, if the same `excludedPaths` element is present throughout the authentication settings, a particular path doesn't require any authentication. In that case, create a rule with `handler: allow`, so that you don't have to exclude the path using a negative lookahead.

  </details>
</div>


The table shows how the APIRule's path value corresponds to the **excludedPaths** values for the API resource, and what negative lookahead value you should add to the `/.*` path:

| Expression type | Description | Sample value | Path value | Negative lookahead value |
|---|---|---|---|---|
|**exact**| no changes to be made | `/exact/path/to/resource.jpg` | `/exact/path/to/resource.jpg` | `(?!/exact/path/to/resource.jpg)` |
|**prefix**| add `.*` at the end | `/pref/` | `/pref/.*` | `(?!/pref/.*)` |
|**suffix**| add `.*` at the beginning; in negative lookahead add `\s` at the end | `/suffix.ico` | `.*/suffix.ico | (?!.*/suffix.ico\s)` |
|**regex**| no changes to be made | `/anything.*` | `/anything.*` | `(?!/anything.*)` |
