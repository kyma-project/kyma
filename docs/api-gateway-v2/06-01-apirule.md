---
title: APIRule
type: Custom Resource
---

The `apirule.gateway.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format the API Gateway Controller listens for. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```shell
kubectl get crd apirule.gateway.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample custom resource (CR) that the API Gateway Controller listens for to expose a service. This example has the **rules** section specified which makes the API Gateway Controller create an Oathkeeper Access Rule for the service.

```yaml
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: service-secured
spec:
  gateway: kyma-gateway.kyma-system.svc.cluster.local
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

This table lists all the possible parameters of a given resource together with their descriptions:

| Field   |      Mandatory      |  Description |
|---|:---:|---|
| **metadata.name** | **YES** | Specifies the name of the exposed API. |
| **spec.gateway** | **YES** | Specifies the Istio Gateway. |
| **spec.service.name** | **YES** | Specifies the name of the exposed service. |
| **spec.service.port** | **YES** | Specifies the communication port of the exposed service. |
| **spec.service.host** | **YES** | Specifies the service's communication address for inbound external traffic. |
| **spec.rules** | **YES** | Specifies the array of Oathkeeper access rules. |
| **spec.rules.path** | **YES** | Specifies the path of the exposed service. |
| **spec.rules.methods** | **NO** | Specifies the list of HTTP request methods available for **spec.rules.path**. |
| **spec.rules.mutators** | **NO** | Specifies the array of [Oathkeeper mutators](https://www.ory.sh/docs/oathkeeper/pipeline/mutator). |
| **spec.rules.accessStrategies** | **YES** | Specifies the array of [Oathkeeper authenticators](https://www.ory.sh/docs/oathkeeper/pipeline/authn). The supported authenticators are `oauth2_introspection`, `jwt`, `noop`, `allow`. |

## Additional information

When you fetch an existing APIRule CR, the system adds the **status** section which describes the status of the Virtual Service and the Oathkeeper Access Rule created for this CR. This table lists the fields of the **status** section.

| Field   |  Description |
|:---|:---|
| **status.apiRuleStatus** | Status code describing the APIRule CR. |
| **status.virtualServiceStatus.code** | Status code describing the Virtual Service. |
| **status.virtualService.desc** | Current state of the Virtual Service. |
| **status.accessRuleStatus.code** | Status code describing the Oathkeeper Rule. |
| **status.accessRuleStatus.desc** | Current state of the Oathkeeper Rule. |

### Status codes

These are the status codes used to describe the Virtual Services and Oathkeeper Access Rules:

| Code   |  Description |
|---|---|
| **OK** | Resource created. |
| **SKIPPED** | Skipped creating a resource. |
| **ERROR** | Resource not created. |
