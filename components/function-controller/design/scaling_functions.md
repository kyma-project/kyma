# Serverless Functions Scaling Modes

## Summary
Initially, the Serverless API supported the `spec.ScaleConfig` field only. The workflow was as follows:
- Function resources were defaulted to min/max replicas = 1
- If you set a different scaling config, Function Controller created HPA resources with the user-defined min/max values. 
- The HPA resource target ref was the Function runtime deployment.
- The Function Controller did not enforce the runtime deployment `spec.Replicas` anymore and it was handled by the HPA resources.


Later, the Serverless API was extended by the `Scale` subresource, which allows for direct scaling of the Function resources through the Kubernetes API.

However, there are some implementation conflicts between the two features. This is a design and an implementation plan to unify the UX while using Functions as scaled resources.

### Goals
- Support Function scale subresource and `spec.ScaleConfig` without conflicts.
- Provide frictionless UX for the feature.

## Proposal
Describe and implement two different scaling configuration. Both configurations already work to some extent. The point here is have an ergonomic UX and flow.

### External scaling configuration
This is managed and configured using `spec.Replicas`. It describes the scale subresource use case. It supports:
- Manual scaling of the Function up and down through the API.
- Configuring an HPA resource with the Function resources as a target.
- Using an external scaler like [KEDA](https://keda.sh/).

### Built-in scaling configuration
This is managed and enabled by setting `spec.ScaleConfig`. It is configurable using Busola and it provides the most basic scaling configuration for the Function.

This configuration is disabled by removing `spec.ScaleConfig`. The Busola UI can be extended to allow you to add or remove `spec.ScaleConfig` to manage built-in scaling with minimal effort.

## Implementation details

- The controller should support and accept both `spec.Replicas` and `spec.ScaleConfig`. Current validation rule to block this will be removed.
- `spec.Replicas` is the only source of truth for scaling the Function/runtime deployment.
- `spec.ScaleConfig` is only used to configure the controller internal HPA.
- The internal HPA is removed if `spec.ScaleConfig == nil`
- The HPA resource created by the controller will still target the Function resources.
- The current Function status update logic should be fixed to reflect the function current scale. 