//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationInput) DeepCopyInto(out *ApplicationInput) {
	*out = *in
	in.Namespaces.DeepCopyInto(&out.Namespaces)
	in.Containers.DeepCopyInto(&out.Containers)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationInput.
func (in *ApplicationInput) DeepCopy() *ApplicationInput {
	if in == nil {
		return nil
	}
	out := new(ApplicationInput)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FileMount) DeepCopyInto(out *FileMount) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FileMount.
func (in *FileMount) DeepCopy() *FileMount {
	if in == nil {
		return nil
	}
	out := new(FileMount)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Filter) DeepCopyInto(out *Filter) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Filter.
func (in *Filter) DeepCopy() *Filter {
	if in == nil {
		return nil
	}
	out := new(Filter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPOutput) DeepCopyInto(out *HTTPOutput) {
	*out = *in
	in.Host.DeepCopyInto(&out.Host)
	in.User.DeepCopyInto(&out.User)
	in.Password.DeepCopyInto(&out.Password)
	out.TLSConfig = in.TLSConfig
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPOutput.
func (in *HTTPOutput) DeepCopy() *HTTPOutput {
	if in == nil {
		return nil
	}
	out := new(HTTPOutput)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Input) DeepCopyInto(out *Input) {
	*out = *in
	in.Application.DeepCopyInto(&out.Application)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Input.
func (in *Input) DeepCopy() *Input {
	if in == nil {
		return nil
	}
	out := new(Input)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InputContainers) DeepCopyInto(out *InputContainers) {
	*out = *in
	if in.Include != nil {
		in, out := &in.Include, &out.Include
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Exclude != nil {
		in, out := &in.Exclude, &out.Exclude
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InputContainers.
func (in *InputContainers) DeepCopy() *InputContainers {
	if in == nil {
		return nil
	}
	out := new(InputContainers)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InputNamespaces) DeepCopyInto(out *InputNamespaces) {
	*out = *in
	if in.Include != nil {
		in, out := &in.Include, &out.Include
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Exclude != nil {
		in, out := &in.Exclude, &out.Exclude
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InputNamespaces.
func (in *InputNamespaces) DeepCopy() *InputNamespaces {
	if in == nil {
		return nil
	}
	out := new(InputNamespaces)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogParser) DeepCopyInto(out *LogParser) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogParser.
func (in *LogParser) DeepCopy() *LogParser {
	if in == nil {
		return nil
	}
	out := new(LogParser)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LogParser) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogParserCondition) DeepCopyInto(out *LogParserCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogParserCondition.
func (in *LogParserCondition) DeepCopy() *LogParserCondition {
	if in == nil {
		return nil
	}
	out := new(LogParserCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogParserList) DeepCopyInto(out *LogParserList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]LogParser, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogParserList.
func (in *LogParserList) DeepCopy() *LogParserList {
	if in == nil {
		return nil
	}
	out := new(LogParserList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LogParserList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogParserSpec) DeepCopyInto(out *LogParserSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogParserSpec.
func (in *LogParserSpec) DeepCopy() *LogParserSpec {
	if in == nil {
		return nil
	}
	out := new(LogParserSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogParserStatus) DeepCopyInto(out *LogParserStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]LogParserCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogParserStatus.
func (in *LogParserStatus) DeepCopy() *LogParserStatus {
	if in == nil {
		return nil
	}
	out := new(LogParserStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogPipeline) DeepCopyInto(out *LogPipeline) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogPipeline.
func (in *LogPipeline) DeepCopy() *LogPipeline {
	if in == nil {
		return nil
	}
	out := new(LogPipeline)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LogPipeline) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogPipelineCondition) DeepCopyInto(out *LogPipelineCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogPipelineCondition.
func (in *LogPipelineCondition) DeepCopy() *LogPipelineCondition {
	if in == nil {
		return nil
	}
	out := new(LogPipelineCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogPipelineList) DeepCopyInto(out *LogPipelineList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]LogPipeline, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogPipelineList.
func (in *LogPipelineList) DeepCopy() *LogPipelineList {
	if in == nil {
		return nil
	}
	out := new(LogPipelineList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LogPipelineList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogPipelineSpec) DeepCopyInto(out *LogPipelineSpec) {
	*out = *in
	in.Input.DeepCopyInto(&out.Input)
	if in.Filters != nil {
		in, out := &in.Filters, &out.Filters
		*out = make([]Filter, len(*in))
		copy(*out, *in)
	}
	in.Output.DeepCopyInto(&out.Output)
	if in.Files != nil {
		in, out := &in.Files, &out.Files
		*out = make([]FileMount, len(*in))
		copy(*out, *in)
	}
	if in.Variables != nil {
		in, out := &in.Variables, &out.Variables
		*out = make([]VariableRef, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogPipelineSpec.
func (in *LogPipelineSpec) DeepCopy() *LogPipelineSpec {
	if in == nil {
		return nil
	}
	out := new(LogPipelineSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogPipelineStatus) DeepCopyInto(out *LogPipelineStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]LogPipelineCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogPipelineStatus.
func (in *LogPipelineStatus) DeepCopy() *LogPipelineStatus {
	if in == nil {
		return nil
	}
	out := new(LogPipelineStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LokiOutput) DeepCopyInto(out *LokiOutput) {
	*out = *in
	in.URL.DeepCopyInto(&out.URL)
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.RemoveKeys != nil {
		in, out := &in.RemoveKeys, &out.RemoveKeys
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LokiOutput.
func (in *LokiOutput) DeepCopy() *LokiOutput {
	if in == nil {
		return nil
	}
	out := new(LokiOutput)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OtlpOutput) DeepCopyInto(out *OtlpOutput) {
	*out = *in
	in.URL.DeepCopyInto(&out.URL)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OtlpOutput.
func (in *OtlpOutput) DeepCopy() *OtlpOutput {
	if in == nil {
		return nil
	}
	out := new(OtlpOutput)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Output) DeepCopyInto(out *Output) {
	*out = *in
	if in.HTTP != nil {
		in, out := &in.HTTP, &out.HTTP
		*out = new(HTTPOutput)
		(*in).DeepCopyInto(*out)
	}
	if in.Loki != nil {
		in, out := &in.Loki, &out.Loki
		*out = new(LokiOutput)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Output.
func (in *Output) DeepCopy() *Output {
	if in == nil {
		return nil
	}
	out := new(Output)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretKeyRef) DeepCopyInto(out *SecretKeyRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretKeyRef.
func (in *SecretKeyRef) DeepCopy() *SecretKeyRef {
	if in == nil {
		return nil
	}
	out := new(SecretKeyRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLSConfig) DeepCopyInto(out *TLSConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSConfig.
func (in *TLSConfig) DeepCopy() *TLSConfig {
	if in == nil {
		return nil
	}
	out := new(TLSConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TracePipeline) DeepCopyInto(out *TracePipeline) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TracePipeline.
func (in *TracePipeline) DeepCopy() *TracePipeline {
	if in == nil {
		return nil
	}
	out := new(TracePipeline)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TracePipeline) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TracePipelineList) DeepCopyInto(out *TracePipelineList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TracePipeline, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TracePipelineList.
func (in *TracePipelineList) DeepCopy() *TracePipelineList {
	if in == nil {
		return nil
	}
	out := new(TracePipelineList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TracePipelineList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TracePipelineOutput) DeepCopyInto(out *TracePipelineOutput) {
	*out = *in
	in.Otlp.DeepCopyInto(&out.Otlp)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TracePipelineOutput.
func (in *TracePipelineOutput) DeepCopy() *TracePipelineOutput {
	if in == nil {
		return nil
	}
	out := new(TracePipelineOutput)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TracePipelineSpec) DeepCopyInto(out *TracePipelineSpec) {
	*out = *in
	in.Output.DeepCopyInto(&out.Output)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TracePipelineSpec.
func (in *TracePipelineSpec) DeepCopy() *TracePipelineSpec {
	if in == nil {
		return nil
	}
	out := new(TracePipelineSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TracePipelineStatus) DeepCopyInto(out *TracePipelineStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TracePipelineStatus.
func (in *TracePipelineStatus) DeepCopy() *TracePipelineStatus {
	if in == nil {
		return nil
	}
	out := new(TracePipelineStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ValueFromSource) DeepCopyInto(out *ValueFromSource) {
	*out = *in
	if in.SecretKeyRef != nil {
		in, out := &in.SecretKeyRef, &out.SecretKeyRef
		*out = new(SecretKeyRef)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ValueFromSource.
func (in *ValueFromSource) DeepCopy() *ValueFromSource {
	if in == nil {
		return nil
	}
	out := new(ValueFromSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ValueType) DeepCopyInto(out *ValueType) {
	*out = *in
	if in.ValueFrom != nil {
		in, out := &in.ValueFrom, &out.ValueFrom
		*out = new(ValueFromSource)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ValueType.
func (in *ValueType) DeepCopy() *ValueType {
	if in == nil {
		return nil
	}
	out := new(ValueType)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VariableRef) DeepCopyInto(out *VariableRef) {
	*out = *in
	in.ValueFrom.DeepCopyInto(&out.ValueFrom)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VariableRef.
func (in *VariableRef) DeepCopy() *VariableRef {
	if in == nil {
		return nil
	}
	out := new(VariableRef)
	in.DeepCopyInto(out)
	return out
}
