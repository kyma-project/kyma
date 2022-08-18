---
title: API Rule
---

The `apirule.gateway.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format the API Gateway Controller listens for. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

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

>**NOTE:** Since Kyma 2.5 the `valpha1` resource has been deprecated. However, you can still create it. It is stored as `v1beta1`. 

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
</div>


This table lists all the possible parameters of a given resource together with their descriptions:

| Field   |      Mandatory      |  Description |
|---|:---:|---|
| **metadata.name** | **YES** | Specifies the name of the exposed API. |
| **spec.gateway** | **YES** | Specifies the Istio Gateway. |
| **spec.host** | **YES** | Specifies the service's communication address for inbound external traffic. If only the leftmost label is provided, the default domain name will be used. |
| **spec.service.name** | **NO** | Specifies the name of the exposed service. |
| **spec.service.port** | **NO** | Specifies the communication port of the exposed service. |
| **spec.rules** | **YES** | Specifies the array of Oathkeeper access rules. |
| **spec.rules.service** | **NO** | Specifies the name of the exposed service. **Note that it has higher precedence than the definition at the spec.service level**.|
| **spec.rules.path** | **YES** | Specifies the path of the exposed service. |
| **spec.rules.methods** | **NO** | Specifies the list of HTTP request methods available for **spec.rules.path**. |
| **spec.rules.mutators** | **NO** | Specifies the array of [Oathkeeper mutators](https://www.ory.sh/docs/next/oathkeeper/pipeline/mutator). |
| **spec.rules.accessStrategies** | **YES** | Specifies the array of [Oathkeeper authenticators](https://www.ory.sh/docs/next/oathkeeper/pipeline/authn). The supported authenticators are `oauth2_introspection`, `jwt`, `noop`, `allow`. |

**CAUTION:** If `service` is not defined at **spec.service** level, all defined rules must have `service` defined at **spec.rules.service** level, otherwise the validation fails. 

## Additional information

When you fetch an existing APIRule CR, the system adds the **status** section which describes the status of the VirtualService and the Oathkeeper Access Rule created for this CR. This table lists the fields of the **status** section.

| Field   |  Description |
|:---|:---|
| **status.apiRuleStatus** | Status code describing the APIRule CR. |
| **status.virtualServiceStatus.code** | Status code describing the VirtualService. |
| **status.virtualService.desc** | Current state of the VirtualService. |
| **status.accessRuleStatus.code** | Status code describing the Oathkeeper Rule. |
| **status.accessRuleStatus.desc** | Current state of the Oathkeeper Rule. |

### Status codes

These are the status codes used to describe the VirtualServices and Oathkeeper Access Rules:

| Code   |  Description |
|---|---|
| **OK** | Resource created. |
| **SKIPPED** | Skipped creating a resource. |
| **ERROR** | Resource not created. |
