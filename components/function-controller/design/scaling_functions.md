# Serverless Functions as scaled resources

## Summary
Initially, the Serverless API supported the `spec.ScaleConfig` field only. The work flow was as follow:
- Function resources were defaulted to min/max replicas = 1
- If the user sets a different scaling config, the function controller will create an HPA resources with the user defined min/max values. 
- The HPA resource target ref will be the Function runtime deployment.
- The Function controller will not enforce the runtime deployment `spec.Replicas` any more and it will be handled by the HPA resources.


Later, we extended the Serverless API to include a `Scale` subresource, which allows us to directly scale the function resources through the Kubernetes API.

However, there are some implementation conflicts between the two features. This is a design and an implementation plan to unify the UX while using Functions as scaled resources.

### Goals
- Support Function scale subresource and `spec.ScaleConfig` without conflicts.
- Provide frictionless UX for the feature.

## Proposal
Describe and implement two different scaling modes. Both modes already work to some extent. The point here is have an ergonomic UX and flow.

### External scaling mode
This describes the scale subresource use case. It supports:
- Manually scaling the Function up/down through the API.
- Configuring an HPA resource with the Function resources as a target.
- Using an external scaler like KEDA.

This mode requires user intervention to some extent. So, it's not enabled by default. However, it should be simple to enable by just using the scale subresource.

To disable this mode the user needs to edit the Function resource and remove the `spec.Replicas` field.

### Built-in scaling mode
This describes the current implementation of `spec.ScaleConfig`. It's enabled by default and should be supported by default by Busola.

This mode will be disabled automatically once the external mode is enabled by setting the `spec.Replicas` field. Busola UI can be extended to allow users to remove `spec.Replica` and re-enable built-in scaling with minimal effort to the user.

## Implementation details

- The controller should support and accept both `spec.Replicas` and `spec.ScaleConfig`. Current validation rule to block this will be removed.
- `spec.Replcas` will take precedence over `spec.ScaleConfig`.
- `spec.ScaleConfig` will be defaulted based on `spec.Replicas` if it's set (min = max = replicas).
- The HPA resource created by the controller will still target the runtime deployment of the Function. Since the controller will only create the HPA resource if there is a difference between `spec.ScaleConfig` min and max, it will reduce the probability of conflict by having more that one HPA managing the Function and it's resources.
- The current Function status update logic should be fixed to reflect the function current scale.