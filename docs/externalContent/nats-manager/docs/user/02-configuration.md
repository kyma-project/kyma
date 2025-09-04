# Configuration

The CustomResourceDefinition (CRD) `nats.operator.kyma-project.io` describes the the NATS custom resource (CR) in detail.
To show the current CRD, run the following command:

   ```shell
   kubectl get crd nats.operator.kyma-project.io -o yaml
   ```

View the complete [NATS CRD](https://github.com/kyma-project/nats-manager/blob/main/config/crd/bases/operator.kyma-project.io_nats.yaml#L1) including detailed descriptions for each field.

The NATS CR configures the settings of NATS JetStream. To edit the settings, run:

   ```shell
   kubectl edit -n kyma-system nats.operator.kyma-project.io <NATS CR Name>
   ```

The CRD is equipped with validation rules and defaulting, so the CR is automatically filled with sensible defaults. You can override the defaults. The validation rules provide guidance when you edit the CR.

## Examples

Use the following sample CRs as guidance. Each can be applied immediately when you [install](../contributor/installation.md) the NATS Manager.

- [Default CR](https://github.com/kyma-project/nats-manager/blob/main/config/samples/default.yaml#L1)
- [Minimal CR](https://github.com/kyma-project/nats-manager/blob/main/config/samples/minimal.yaml#L1)
- [Full spec CR](https://github.com/kyma-project/nats-manager/blob/main/config/samples/nats-full-spec.yaml#L1)

## High availability

For high availability, the NATS servers must be set up across different availability zones for uninterrupted operation and uptime. NATS Manager deploys the NATS servers in the availability zones where your Kubernetes cluster has Nodes. If the Kubernetes cluster has Nodes distributed across at least three availability zones, NATS Manager automatically distributes the NATS servers across these availability zones. If the Kubernetes cluster doesn’t have Nodes distributed across at least three availability zones, high availability is compromised.

## Reference

<!-- The table below was generated automatically -->
<!-- Some special tags (html comments) are at the end of lines due to markdown requirements. -->
<!-- The content between "TABLE-START" and "TABLE-END" will be replaced -->

<!-- TABLE-START -->
### NATS.operator.Kyma-project.io/v1alpha1

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **annotations**  | map\[string\]string | Annotations allows to add annotations to NATS. |
| **cluster**  | object | Cluster defines configurations that are specific to NATS clusters. |
| **cluster.&#x200b;size**  | integer | Size of a NATS cluster, i.e. number of NATS nodes. |
| **jetStream**  | object | JetStream defines configurations that are specific to NATS JetStream. |
| **jetStream.&#x200b;fileStorage**  | object | FileStorage defines configurations to file storage in NATS JetStream. |
| **jetStream.&#x200b;fileStorage.&#x200b;size**  | \{integer or string\} | Size defines the file storage size. |
| **jetStream.&#x200b;fileStorage.&#x200b;storageClassName**  | string | StorageClassName defines the file storage class name. |
| **jetStream.&#x200b;memStorage**  | object | MemStorage defines configurations to memory storage in NATS JetStream. |
| **jetStream.&#x200b;memStorage.&#x200b;enabled**  | boolean | Enabled allows the enablement of memory storage. |
| **jetStream.&#x200b;memStorage.&#x200b;size**  | \{integer or string\} | Size defines the mem. |
| **labels**  | map\[string\]string | Labels allows to add Labels to NATS. |
| **logging**  | object | JetStream defines configurations that are specific to NATS logging in NATS. |
| **logging.&#x200b;debug**  | boolean | Debug allows debug logging. |
| **logging.&#x200b;trace**  | boolean | Trace allows trace logging. |
| **resources**  | object | Resources defines resources for NATS. |
| **resources.&#x200b;claims**  | \[\]object | Claims lists the names of resources, defined in spec.resourceClaims, that are used by this container. This is an alpha field and requires enabling the DynamicResourceAllocation feature gate. This field is immutable. It can only be set for containers.|
| **resources.&#x200b;claims.&#x200b;name** (required) | string | Name must match the name of one entry in Pod.spec.resourceClaims of the Pod where this field is used. It makes that resource available inside a container.|
| **resources.&#x200b;claims.&#x200b;request**  | string | Request is the name chosen for a request in the referenced claim. If empty, everything from the claim is made available, otherwise only the result of this request.|
| **resources.&#x200b;limits**  | map\[string\]\{integer or string\} | Limits describes the maximum amount of compute resources allowed. More info: <https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/>|
| **resources.&#x200b;requests**  | map\[string\]\{integer or string\} | Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. Requests cannot exceed Limits. More info: <https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/> |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **availabilityZonesUsed**  | integer |  |
| **conditions**  | \[\]object | Condition contains details for one aspect of the current state of this API Resource. |
| **conditions.&#x200b;lastTransitionTime** (required) | string | lastTransitionTime is the last time the condition transitioned from one status to another. This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable. |
| **conditions.&#x200b;message** (required) | string | message is a human readable message indicating details about the transition. This may be an empty string.|
| **conditions.&#x200b;observedGeneration**  | integer | observedGeneration represents the .metadata.generation that the condition was set based upon. For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date with respect to the current state of the instance. |
| **conditions.&#x200b;reason** (required) | string | reason contains a programmatic identifier indicating the reason for the condition's last transition. Producers of specific condition types may define expected values and meanings for this field, and whether the values are considered a guaranteed API. The value should be a CamelCase string. This field may not be empty.|
| **conditions.&#x200b;status** (required) | string | status of the condition, one of True, False, Unknown. |
| **conditions.&#x200b;type** (required) | string | type of condition in CamelCase or in foo.example.com/CamelCase. |
| **state** (required) | string |  |
| **url**  | string |  |

<!-- TABLE-END -->
