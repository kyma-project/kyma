---
title: Istio
---

The `istios.operator.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format the Reconciler uses to configure and install Istio. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```shell
kubectl get crd istios.operator.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample Istio custom resource (CR) that the Reconciler uses to configure and install Istio. The following example shows the single supported **numTrustedProxies** configuration setting. There must be only one Istio CR in the cluster.

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: Istio
metadata:
  name: istio
  labels:
    app.kubernetes.io/name: istio
spec:
  config:
    numTrustedProxies: 1
```

The following table lists all the possible parameters of the given resource together with their descriptions:

<!-- TABLE-START -->
### Istio.operator.kyma-project.io/v1alpha1

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **config**  | object | Configures the Istio installation. |
| **config.&#x200b;numTrustedProxies**  | integer | Specifies the number of trusted proxies deployed in front of the Istio gateway proxy. |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **conditions**  | \[\]object | Contains conditions associated with CustomStatus. |
| **conditions.&#x200b;lastTransitionTime** (required) | string | Specifies the last time when the condition transitioned from one status to another. That is, when the underlying condition changed. If not known, using the last time when the API field changed is also acceptable. |
| **conditions.&#x200b;message** (required) | string | Displays a human readable message indicating the details about the transition. It can be an empty string. |
| **conditions.&#x200b;observedGeneration**  | integer | Represents the **.metadata.generation** that the condition was based upon. For example, if **.metadata.generation** is currently 12, but the **.status.conditions[x].observedGeneration** is 9, the condition is out of date with respect to the current state of the instance. |
| **conditions.&#x200b;reason** (required) | string | Contains a programmatic identifier indicating the reason for the condition's last transition. Producers of specific condition types may define expected values and meanings for this field, and whether the values are considered a guaranteed API. The value must be a CamelCase string. This field is required. |
| **conditions.&#x200b;status** (required) | string | Describes the status of the condition. The value is either `True`, `False`, or `Unknown`. |
| **conditions.&#x200b;type** (required) | string | Describes the type of the condition in CamelCase or in `foo.example.com/CamelCase`. Many **.condition.type** values are consistent across all resources, for example `Available`, but because arbitrary conditions can be useful (see **.node.status.conditions**), the ability to deconflict is important. It matches the following regex:/ (dns1123SubdomainFmt/)?(qualifiedNameFmt). |
| **state** (required) | string | Signifies the current state of CustomObject. The value is either `Ready`, `Processing`, `Error`, or `Deleting`. |

<!-- TABLE-END -->
