---
title: API Rule
---

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format the API Gateway Controller listens for. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```shell
kubectl get crd apirules.gateway.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample custom resource (CR) that the API Gateway Controller listens for to expose a service. This example has the **rules** section specified which makes the API Gateway Controller create an Oathkeeper Access Rule for the service.


<div tabs name="api-rule" group="sample-cr">
  <details>
  <summary label="v1beta1">
  v1beta1
  </summary>

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
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategies:
        - handler: oauth2_introspection
          config:
            required_scope: ["read"]
```

  </details>
  <details>
  <summary label="v1alpha1">
  v1alpha1
  </summary>

>**NOTE:** Since Kyma 2.5 the `v1alpha1` resource has been deprecated. However, you can still create it. It is stored as `v1beta1`.

```yaml
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: service-secured
spec:
  gateway: kyma-system/kyma-gateway
  service:
    name: foo-service
    port: 8080
    host: foo.bar
  rules:
    - path: /.*
      methods: ["GET"]
      mutators: []
      accessStrategies:
        - handler: oauth2_introspection
          config:
            required_scope: ["read"]
```

  </details>
</div>


This following tables list all the possible parameters of a given resource together with their descriptions:

>**CAUTION:** If `service` is not defined at **spec.service** level, all defined rules must have `service` defined at **spec.rules.service** level, otherwise the validation fails.

<!-- TABLE-START -->
### APIRule.gateway.kyma-project.io/v1beta1

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **gateway** (required) | string | Specifies the Istio Gateway to be used. |
| **host** (required) | string | Specifies the URL of the exposed service. |
| **rules** (required) | \[\]object | Represents the the array of Oathkeeper access rules to be applied. |
| **rules.&#x200b;accessStrategies** (required) | \[\]object | Specifies the list of access strategies. The supported access strategies are [Oathkeeper](https://www.ory.sh/docs/oathkeeper/pipeline/authn) `oauth2_introspection`, `jwt`, `noop`, and `allow`. |
| **rules.&#x200b;accessStrategies.&#x200b;config**  | object | Configures the handler. Configuration keys vary per handler. |
| **rules.&#x200b;accessStrategies.&#x200b;config.&#x200b;jwks_urls**  | \[\]string | . |
| **rules.&#x200b;accessStrategies.&#x200b;config.&#x200b;trusted_issuers**  | \[\]string | . |
| **rules.&#x200b;accessStrategies.&#x200b;handler** (required) | string | Specifies the name of the handler. |
| **rules.&#x200b;methods** (required) | \[\]string | Represents the list of allowed HTTP request methods available for the **spec.rules.path**. |
| **rules.&#x200b;mutators**  | \[\]object | Specifies the list of [Ory Oathkeeper](https://www.ory.sh/docs/oathkeeper/pipeline/mutator) mutators. |
| **rules.&#x200b;mutators.&#x200b;config**  | object | Configures the handler. Configuration keys vary per handler. |
| **rules.&#x200b;mutators.&#x200b;handler** (required) | string | Specifies the name of the handler. |
| **rules.&#x200b;path** (required) | string | Specifies the path of the exposed service. |
| **rules.&#x200b;service**  | object | Describes the service to expose. Overwrites the **spec** level service, if defined. |
| **rules.&#x200b;service.&#x200b;external**  | boolean | Specifies if the service is internal (in cluster) or external. |
| **rules.&#x200b;service.&#x200b;name** (required) | string | Specifies the name of the exposed service. |
| **rules.&#x200b;service.&#x200b;namespace**  | string | Specifies the Namespace of the exposed service. If not defined, it defaults to the APIRule Namespace. |
| **rules.&#x200b;service.&#x200b;port** (required) | integer | Specifies the communication port of the exposed service. |
| **service**  | object | Describes the service to expose. |
| **service.&#x200b;external**  | boolean | Specifies if the service is internal (in cluster) or external. |
| **service.&#x200b;name** (required) | string | Specifies the name of the exposed service. |
| **service.&#x200b;namespace**  | string | Specifies the Namespace of the exposed service. If not defined, it defaults to the APIRule Namespace. |
| **service.&#x200b;port** (required) | integer | Specifies the port of the exposed service. |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **APIRuleStatus**  | object | Describes the status of APIRule. |
| **APIRuleStatus.&#x200b;code**  | string | Status code describing APIRule. |
| **APIRuleStatus.&#x200b;desc**  | string | . |
| **accessRuleStatus**  | object | Describes the status of ORY Oathkeeper Rule. |
| **accessRuleStatus.&#x200b;code**  | string | Status code describing ORY Oathkeeper Rule. |
| **accessRuleStatus.&#x200b;desc**  | string | . |
| **authorizationPolicyStatus**  | object | APIRuleResourceStatus . |
| **authorizationPolicyStatus.&#x200b;code**  | string | StatusCode . |
| **authorizationPolicyStatus.&#x200b;desc**  | string | . |
| **lastProcessedTime**  | string | . |
| **observedGeneration**  | integer | . |
| **requestAuthenticationStatus**  | object | APIRuleResourceStatus . |
| **requestAuthenticationStatus.&#x200b;code**  | string | StatusCode . |
| **requestAuthenticationStatus.&#x200b;desc**  | string | . |
| **virtualServiceStatus**  | object | Describes the status of Istio VirtualService. |
| **virtualServiceStatus.&#x200b;code**  | string | Status code describing Istio VirtualService. |
| **virtualServiceStatus.&#x200b;desc**  | string | . |

### APIRule.gateway.kyma-project.io/v1alpha1

>**CAUTION**: Since Kyma 2.5.X, APIRule in version v1alpha1 has been deprecated. Consider using v1beta1.

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **gateway** (required) | string | Specifies the Istio Gateway to be used. |
| **rules** (required) | \[\]object | Represents the array of Oathkeeper access rules to be applied. |
| **rules.&#x200b;accessStrategies** (required) | \[\]object | Specifies the list of access strategies. The supported access strategies are [Oathkeeper](https://www.ory.sh/docs/oathkeeper/pipeline/authn) `oauth2_introspection`, `jwt`, `noop`, and `allow`. |
| **rules.&#x200b;accessStrategies.&#x200b;config**  | object | Configures the handler. Configuration keys vary per handler. |
| **rules.&#x200b;accessStrategies.&#x200b;config.&#x200b;jwks_urls**  | \[\]string | . |
| **rules.&#x200b;accessStrategies.&#x200b;config.&#x200b;trusted_issuers**  | \[\]string | . |
| **rules.&#x200b;accessStrategies.&#x200b;handler** (required) | string | Specifies the name of the handler. |
| **rules.&#x200b;methods** (required) | \[\]string | Represents the list of allowed HTTP request methods available for the **spec.rules.path**. |
| **rules.&#x200b;mutators**  | \[\]object | Specifies the list of [Oathkeeper mutators](https://www.ory.sh/docs/oathkeeper/pipeline/mutator). |
| **rules.&#x200b;mutators.&#x200b;config**  | object | Configures the handler. Configuration keys vary per handler. |
| **rules.&#x200b;mutators.&#x200b;handler** (required) | string | Specifies the name of the handler. |
| **rules.&#x200b;path** (required) | string | Specifies the path of the exposed service. |
| **service** (required) | object | Describes the service to expose. |
| **service.&#x200b;external**  | boolean | Defines if the service is internal (in cluster) or external. |
| **service.&#x200b;host** (required) | string | Specifies the URL of the exposed service. |
| **service.&#x200b;name** (required) | string | Specifies the name of the exposed service. |
| **service.&#x200b;port** (required) | integer | Specifies the communication port of the exposed service. |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **APIRuleStatus**  | object | Describes the status of APIRule. |
| **APIRuleStatus.&#x200b;code**  | string | Status code describing APIRule. |
| **APIRuleStatus.&#x200b;desc**  | string | . |
| **accessRuleStatus**  | object | Describes the status of ORY Oathkeeper Rule. |
| **accessRuleStatus.&#x200b;code**  | string | Status code describing ORY Oathkeeper Rule. |
| **accessRuleStatus.&#x200b;desc**  | string | . |
| **lastProcessedTime**  | string | . |
| **observedGeneration**  | integer | . |
| **virtualServiceStatus**  | object | Describes the status of Istio VirtualService. |
| **virtualServiceStatus.&#x200b;code**  | string | Status code describing Istio VirtualService. |
| **virtualServiceStatus.&#x200b;desc**  | string | . |

<!-- TABLE-END -->

### Status codes

These are the status codes used to describe the VirtualServices and Oathkeeper Access Rules:

| Code   |  Description |
|---|---|
| **OK** | Resource created. |
| **SKIPPED** | Skipped creating a resource. |
| **ERROR** | Resource not created. |
