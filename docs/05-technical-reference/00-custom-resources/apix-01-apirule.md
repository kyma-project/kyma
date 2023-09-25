---
title: APIRule
---

The `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format the API Gateway Controller listens for. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```shell
kubectl get crd apirules.gateway.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample custom resource (CR) that the API Gateway Controller listens for to expose a service. This example has the **rules** section specified which makes the API Gateway Controller create an Oathkeeper Access Rule for the service.

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

The following table lists all the possible parameters of a given resource together with their descriptions:

>**CAUTION:** If `service` is not defined at **spec.service** level, all defined rules must have `service` defined at **spec.rules.service** level. Otherwise, the validation fails.

<!-- TABLE-START -->
### APIRule.gateway.kyma-project.io/v1beta1

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **gateway** (required) | string | Specifies the Istio Gateway to be used. |
| **host** (required) | string | Specifies the URL of the exposed service. |
| **rules** (required) | \[\]object | Represents the array of Oathkeeper access rules to be applied. |
| **rules.&#x200b;accessStrategies** (required) | \[\]object | Specifies the list of access strategies. All strategies listed in [Oathkeeper documentation](https://www.ory.sh/docs/oathkeeper/pipeline/authn) are supported. |
| **rules.&#x200b;accessStrategies.&#x200b;config**  | object | Configures the handler. Configuration keys vary per handler. |
| **rules.&#x200b;accessStrategies.&#x200b;config.&#x200b;jwks_urls**  | \[\]string | Specifies the array of URLs from which Ory Oathkeeper can retrieve JSON Web Keys for validating JSON Web Token. The value must begin with either `http://`, `https://`, or `file://`. |
| **rules.&#x200b;accessStrategies.&#x200b;config.&#x200b;trusted_issuers**  | \[\]string | If the **trusted_issuers** field is set, the JWT must contain a value for the claim `iss` that matches exactly (case-sensitive) one of the values of **trusted_issuers**. The value must begin with either `http://`, `https://`, or `file://`. |
| **rules.&#x200b;accessStrategies.&#x200b;handler** (required) | string | Specifies the name of the handler. |
| **rules.&#x200b;methods** (required) | \[\]string | Represents the list of allowed HTTP request methods available for the **spec.rules.path**. |
| **rules.&#x200b;mutators**  | \[\]object | Specifies the list of [Ory Oathkeeper mutators](https://www.ory.sh/docs/oathkeeper/pipeline/mutator). |
| **rules.&#x200b;mutators.&#x200b;config**  | object | Configures the handler. Configuration keys vary per handler. |
| **rules.&#x200b;mutators.&#x200b;handler** (required) | string | Specifies the name of the handler. |
| **rules.&#x200b;path** (required) | string | Specifies the path of the exposed service. |
| **rules.&#x200b;service**  | object | Describes the service to expose. Overwrites the **spec** level service if defined. |
| **rules.&#x200b;service.&#x200b;external**  | boolean | Specifies if the service is internal (in cluster) or external. |
| **rules.&#x200b;service.&#x200b;name** (required) | string | Specifies the name of the exposed service. |
| **rules.&#x200b;service.&#x200b;namespace**  | string | Specifies the Namespace of the exposed service. If not defined, it defaults to the APIRule Namespace. |
| **rules.&#x200b;service.&#x200b;port** (required) | integer | Specifies the communication port of the exposed service. |
| **rules.&#x200b;timeout**  | integer | Specifies the timeout, in seconds, for HTTP requests made to **spec.rules.path**. The maximum timeout is limited to 3900 seconds (65 minutes). Timeout definitions set at this level take precedence over any timeout defined at the **spec.timeout** level. |
| **service**  | object | Describes the service to expose. |
| **service.&#x200b;external**  | boolean | Specifies if the service is internal (in cluster) or external. |
| **service.&#x200b;name** (required) | string | Specifies the name of the exposed service. |
| **service.&#x200b;namespace**  | string | Specifies the Namespace of the exposed service. If not defined, it defaults to the APIRule Namespace. |
| **service.&#x200b;port** (required) | integer | Specifies the communication port of the exposed service. |
| **timeout**  | integer | Specifies the timeout, in seconds, for HTTP requests for all Oathkeeper access rules. However, this value can be overridden for each individual rule. The maximum timeout is limited to 3900 seconds (65 minutes). If no timeout is specified, the default timeout of 180 seconds applies. |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **APIRuleStatus**  | object | Describes the status of APIRule. |
| **APIRuleStatus.&#x200b;code**  | string | Status code describing APIRule. |
| **APIRuleStatus.&#x200b;desc**  | string | Explains the status of APIRule. |
| **accessRuleStatus**  | object | Describes the status of ORY Oathkeeper Rule. |
| **accessRuleStatus.&#x200b;code**  | string | Status code describing ORY Oathkeeper Rule. |
| **accessRuleStatus.&#x200b;desc**  | string | Explains the status of ORY Oathkeeper Rule. |
| **authorizationPolicyStatus**  | object | Describes the status of the Istio Authorization Policy subresource. |
| **authorizationPolicyStatus.&#x200b;code**  | string | Status code describing the Istio Authorization Policy subresource. |
| **authorizationPolicyStatus.&#x200b;desc**  | string | Explains the status of the Istio Authorization Policy subresource. |
| **lastProcessedTime**  | string | Indicates the timestamp when the API Gateway controller last processed APIRule. |
| **observedGeneration**  | integer | Specifies the generation of the resource that was observed by the API Gateway controller. |
| **requestAuthenticationStatus**  | object | Describes the status of the Istio Request Authentication subresource. |
| **requestAuthenticationStatus.&#x200b;code**  | string | Status code describing the state of the Istio Authorization Policy subresource. |
| **requestAuthenticationStatus.&#x200b;desc**  | string | Explains the status of the Istio Request Authentication subresource. |
| **virtualServiceStatus**  | object | Describes the status of Istio VirtualService. |
| **virtualServiceStatus.&#x200b;code**  | string | Status code describing Istio VirtualService. |
| **virtualServiceStatus.&#x200b;desc**  | string | Explains the status of Istio VirtualService. |

<!-- TABLE-END -->

### Status codes

These are the status codes used to describe the VirtualServices and Oathkeeper Access Rules:

| Code   |  Description |
|---|---|
| **OK** | Resource created. |
| **SKIPPED** | Skipped creating a resource. |
| **ERROR** | Resource not created. |