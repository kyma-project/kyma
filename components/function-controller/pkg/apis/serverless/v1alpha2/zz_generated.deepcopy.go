//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha2

import (
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BuildJobDefaulting) DeepCopyInto(out *BuildJobDefaulting) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BuildJobDefaulting.
func (in *BuildJobDefaulting) DeepCopy() *BuildJobDefaulting {
	if in == nil {
		return nil
	}
	out := new(BuildJobDefaulting)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BuildJobResourcesDefaulting) DeepCopyInto(out *BuildJobResourcesDefaulting) {
	*out = *in
	if in.Presets != nil {
		in, out := &in.Presets, &out.Presets
		*out = make(map[string]ResourcesPreset, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BuildJobResourcesDefaulting.
func (in *BuildJobResourcesDefaulting) DeepCopy() *BuildJobResourcesDefaulting {
	if in == nil {
		return nil
	}
	out := new(BuildJobResourcesDefaulting)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Condition) DeepCopyInto(out *Condition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Condition.
func (in *Condition) DeepCopy() *Condition {
	if in == nil {
		return nil
	}
	out := new(Condition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigMapRef) DeepCopyInto(out *ConfigMapRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigMapRef.
func (in *ConfigMapRef) DeepCopy() *ConfigMapRef {
	if in == nil {
		return nil
	}
	out := new(ConfigMapRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DefaultingConfig) DeepCopyInto(out *DefaultingConfig) {
	*out = *in
	in.Function.DeepCopyInto(&out.Function)
	in.BuildJob.DeepCopyInto(&out.BuildJob)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DefaultingConfig.
func (in *DefaultingConfig) DeepCopy() *DefaultingConfig {
	if in == nil {
		return nil
	}
	out := new(DefaultingConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Function) DeepCopyInto(out *Function) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Function.
func (in *Function) DeepCopy() *Function {
	if in == nil {
		return nil
	}
	out := new(Function)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Function) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FunctionDefaulting) DeepCopyInto(out *FunctionDefaulting) {
	*out = *in
	in.Replicas.DeepCopyInto(&out.Replicas)
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FunctionDefaulting.
func (in *FunctionDefaulting) DeepCopy() *FunctionDefaulting {
	if in == nil {
		return nil
	}
	out := new(FunctionDefaulting)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FunctionList) DeepCopyInto(out *FunctionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Function, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FunctionList.
func (in *FunctionList) DeepCopy() *FunctionList {
	if in == nil {
		return nil
	}
	out := new(FunctionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FunctionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FunctionReplicasDefaulting) DeepCopyInto(out *FunctionReplicasDefaulting) {
	*out = *in
	if in.Presets != nil {
		in, out := &in.Presets, &out.Presets
		*out = make(map[string]ReplicasPreset, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FunctionReplicasDefaulting.
func (in *FunctionReplicasDefaulting) DeepCopy() *FunctionReplicasDefaulting {
	if in == nil {
		return nil
	}
	out := new(FunctionReplicasDefaulting)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FunctionResourcesDefaulting) DeepCopyInto(out *FunctionResourcesDefaulting) {
	*out = *in
	if in.Presets != nil {
		in, out := &in.Presets, &out.Presets
		*out = make(map[string]ResourcesPreset, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.RuntimePresets != nil {
		in, out := &in.RuntimePresets, &out.RuntimePresets
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FunctionResourcesDefaulting.
func (in *FunctionResourcesDefaulting) DeepCopy() *FunctionResourcesDefaulting {
	if in == nil {
		return nil
	}
	out := new(FunctionResourcesDefaulting)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FunctionSpec) DeepCopyInto(out *FunctionSpec) {
	*out = *in
	if in.CustomRuntimeConfiguration != nil {
		in, out := &in.CustomRuntimeConfiguration, &out.CustomRuntimeConfiguration
		*out = new(ConfigMapRef)
		**out = **in
	}
	in.Source.DeepCopyInto(&out.Source)
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.ResourceConfiguration.DeepCopyInto(&out.ResourceConfiguration)
	if in.ScaleConfig != nil {
		in, out := &in.ScaleConfig, &out.ScaleConfig
		*out = new(ScaleConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FunctionSpec.
func (in *FunctionSpec) DeepCopy() *FunctionSpec {
	if in == nil {
		return nil
	}
	out := new(FunctionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FunctionStatus) DeepCopyInto(out *FunctionStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.Repository = in.Repository
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FunctionStatus.
func (in *FunctionStatus) DeepCopy() *FunctionStatus {
	if in == nil {
		return nil
	}
	out := new(FunctionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitRepositorySource) DeepCopyInto(out *GitRepositorySource) {
	*out = *in
	if in.Auth != nil {
		in, out := &in.Auth, &out.Auth
		*out = new(RepositoryAuth)
		**out = **in
	}
	out.Repository = in.Repository
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitRepositorySource.
func (in *GitRepositorySource) DeepCopy() *GitRepositorySource {
	if in == nil {
		return nil
	}
	out := new(GitRepositorySource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InlineSource) DeepCopyInto(out *InlineSource) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InlineSource.
func (in *InlineSource) DeepCopy() *InlineSource {
	if in == nil {
		return nil
	}
	out := new(InlineSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MinBuildJobResourcesValues) DeepCopyInto(out *MinBuildJobResourcesValues) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MinBuildJobResourcesValues.
func (in *MinBuildJobResourcesValues) DeepCopy() *MinBuildJobResourcesValues {
	if in == nil {
		return nil
	}
	out := new(MinBuildJobResourcesValues)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MinBuildJobValues) DeepCopyInto(out *MinBuildJobValues) {
	*out = *in
	out.Resources = in.Resources
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MinBuildJobValues.
func (in *MinBuildJobValues) DeepCopy() *MinBuildJobValues {
	if in == nil {
		return nil
	}
	out := new(MinBuildJobValues)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MinFunctionReplicasValues) DeepCopyInto(out *MinFunctionReplicasValues) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MinFunctionReplicasValues.
func (in *MinFunctionReplicasValues) DeepCopy() *MinFunctionReplicasValues {
	if in == nil {
		return nil
	}
	out := new(MinFunctionReplicasValues)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MinFunctionResourcesValues) DeepCopyInto(out *MinFunctionResourcesValues) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MinFunctionResourcesValues.
func (in *MinFunctionResourcesValues) DeepCopy() *MinFunctionResourcesValues {
	if in == nil {
		return nil
	}
	out := new(MinFunctionResourcesValues)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MinFunctionValues) DeepCopyInto(out *MinFunctionValues) {
	*out = *in
	out.Replicas = in.Replicas
	out.Resources = in.Resources
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MinFunctionValues.
func (in *MinFunctionValues) DeepCopy() *MinFunctionValues {
	if in == nil {
		return nil
	}
	out := new(MinFunctionValues)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReplicasPreset) DeepCopyInto(out *ReplicasPreset) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReplicasPreset.
func (in *ReplicasPreset) DeepCopy() *ReplicasPreset {
	if in == nil {
		return nil
	}
	out := new(ReplicasPreset)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Repository) DeepCopyInto(out *Repository) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Repository.
func (in *Repository) DeepCopy() *Repository {
	if in == nil {
		return nil
	}
	out := new(Repository)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RepositoryAuth) DeepCopyInto(out *RepositoryAuth) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RepositoryAuth.
func (in *RepositoryAuth) DeepCopy() *RepositoryAuth {
	if in == nil {
		return nil
	}
	out := new(RepositoryAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceConfiguration) DeepCopyInto(out *ResourceConfiguration) {
	*out = *in
	in.Build.DeepCopyInto(&out.Build)
	in.Function.DeepCopyInto(&out.Function)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceConfiguration.
func (in *ResourceConfiguration) DeepCopy() *ResourceConfiguration {
	if in == nil {
		return nil
	}
	out := new(ResourceConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceRequirements) DeepCopyInto(out *ResourceRequirements) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceRequirements.
func (in *ResourceRequirements) DeepCopy() *ResourceRequirements {
	if in == nil {
		return nil
	}
	out := new(ResourceRequirements)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourcesPreset) DeepCopyInto(out *ResourcesPreset) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourcesPreset.
func (in *ResourcesPreset) DeepCopy() *ResourcesPreset {
	if in == nil {
		return nil
	}
	out := new(ResourcesPreset)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScaleConfig) DeepCopyInto(out *ScaleConfig) {
	*out = *in
	if in.MinReplicas != nil {
		in, out := &in.MinReplicas, &out.MinReplicas
		*out = new(int32)
		**out = **in
	}
	if in.MaxReplicas != nil {
		in, out := &in.MaxReplicas, &out.MaxReplicas
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScaleConfig.
func (in *ScaleConfig) DeepCopy() *ScaleConfig {
	if in == nil {
		return nil
	}
	out := new(ScaleConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Source) DeepCopyInto(out *Source) {
	*out = *in
	if in.GitRepository != nil {
		in, out := &in.GitRepository, &out.GitRepository
		*out = new(GitRepositorySource)
		(*in).DeepCopyInto(*out)
	}
	if in.Inline != nil {
		in, out := &in.Inline, &out.Inline
		*out = new(InlineSource)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Source.
func (in *Source) DeepCopy() *Source {
	if in == nil {
		return nil
	}
	out := new(Source)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ValidationConfig) DeepCopyInto(out *ValidationConfig) {
	*out = *in
	if in.ReservedEnvs != nil {
		in, out := &in.ReservedEnvs, &out.ReservedEnvs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	out.Function = in.Function
	out.BuildJob = in.BuildJob
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ValidationConfig.
func (in *ValidationConfig) DeepCopy() *ValidationConfig {
	if in == nil {
		return nil
	}
	out := new(ValidationConfig)
	in.DeepCopyInto(out)
	return out
}
