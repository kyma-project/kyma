# Function subresources templating for Serverless v1alpha2

## Summary

The current Serverless API allows for limited configuration of the generated Function's subresources (deployment, build job). Additionally, it doesn't allow for separate configuration for build and Function's resources except for memory/CPU resources.

The current v1alpha2 Serverless API supports templating of Function's subresources _without_ the need to roll out a new API version.

## Motivation

Give Serverless users the ability to:
- Separate configuration for Function deployments and build jobs.
- Configure volume mounts for function subresources
- Configure sidecars for function deployments (?)

### Goals

- Add more flexibility to the Serverless API.
- Provide a rollout plan to implement this functionality in small increments and avoid the need to rollout a new API version

### Non-Goals
- Define a new Spec API version

## Proposal

Use a modified version of [PodTemplateSpec](https://github.com/kubernetes/kubernetes/blob/64ed9145452d2d1d324d2437566f1ea1ce76f226/pkg/apis/core/types.go#L3443) as a base for function build/deployment templates. 

A modified version is needed because we need to protect certain parts of the Pod spec (commands and args for the main container in the pod for example).

The proposal doesn't include removing or moving any existing Spec fields and only adds non-breaking changes to the Spec. This ensure that there is no need to rollout a new API version and ensure backward compatibility with released v1alpha2 Spec/API.

## Design details

Currently, we expose the following _relevant_ high level Spec fields:
- `Env`
- `ResourceConfiguration`
- `ScaleConfig`
- `Replicas`
- `Template` (Labels and annotations, currently not applied)

The proposed functionality can be implemented by adding the two following fields to the spec:
- `RuntimeTemplate`
- `BuildTemplate`

*Note:* We should decide if we want to the implement this directly under the function Spec or grouped in a Spec field, e.g. `spec.Templates`.

Both fields will be of the modified version of the PodTemplateSpec.

Both fields will be optional. They should be defaulted based on the values of the high level Spec fields.

### Defaulting

The high level Spec fields should be defaulted by the _defaulting webhook_.

If the template fields are not set, they should be defaulted based on the high level Spec by the _function controller_. 

If the fields are partially set, the controller should fill them with the user defined values and default unset subfields based on the high level fields.

### Precedence

The lower Spec fields should have higher precedence over the high level fields. This it would be possible to only override a limited set of the template subfields without having to fill out the full template.

### Version Conversion

The backward compatibility/conversion for these fields into v1alpha1 is not supported as we are moving away from v1alpha1.

### Upgrades and Backward Compatibility

The proposed changes must not cause any backward compatibility issues with _existing_ v1alpha2 Functions created before releasing these changes. Worst case upgrade scenario should not be worse than triggering a Function rebuild due to specification changes in the Function and its subresources.

### Kyma Cli/Busola Support

These fields should not be supported by Kyma CLI or Busola. The implementation is complex, and this advanced feature is not likely to be used through these interfaces. Additionally, it complicates the gradual rollout of the feature.

## Implementation Breakdown

As high level breakdown of the required effort, it should be possible to implement and merge the following tasks individually into main. Trying to further breakdown the list would be great.

- Define and implement the modified `PodSpecTemplate` type.
- Add required defaulting logic to the controller (no changes in reconciliation).
- Refactor build job reconciliation to be based on the the build template.
- Add Integration tests for build job reconciliation to be based on the build template.
- Refactor runtime deployment reconciliation to be based on runtime template
- Add Integration tests for runtime deployment reconciliation to be based on the runtime deployment template.
