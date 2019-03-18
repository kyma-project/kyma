// +build !ignore_autogenerated

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterMicroFrontend) DeepCopyInto(out *ClusterMicroFrontend) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterMicroFrontend.
func (in *ClusterMicroFrontend) DeepCopy() *ClusterMicroFrontend {
	if in == nil {
		return nil
	}
	out := new(ClusterMicroFrontend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterMicroFrontend) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterMicroFrontendList) DeepCopyInto(out *ClusterMicroFrontendList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ClusterMicroFrontend, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterMicroFrontendList.
func (in *ClusterMicroFrontendList) DeepCopy() *ClusterMicroFrontendList {
	if in == nil {
		return nil
	}
	out := new(ClusterMicroFrontendList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterMicroFrontendList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterMicroFrontendSpec) DeepCopyInto(out *ClusterMicroFrontendSpec) {
	*out = *in
	in.CommonMicroFrontendSpec.DeepCopyInto(&out.CommonMicroFrontendSpec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterMicroFrontendSpec.
func (in *ClusterMicroFrontendSpec) DeepCopy() *ClusterMicroFrontendSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterMicroFrontendSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CommonMicroFrontendSpec) DeepCopyInto(out *CommonMicroFrontendSpec) {
	*out = *in
	if in.NavigationNodes != nil {
		in, out := &in.NavigationNodes, &out.NavigationNodes
		*out = make([]NavigationNode, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CommonMicroFrontendSpec.
func (in *CommonMicroFrontendSpec) DeepCopy() *CommonMicroFrontendSpec {
	if in == nil {
		return nil
	}
	out := new(CommonMicroFrontendSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MicroFrontend) DeepCopyInto(out *MicroFrontend) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MicroFrontend.
func (in *MicroFrontend) DeepCopy() *MicroFrontend {
	if in == nil {
		return nil
	}
	out := new(MicroFrontend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MicroFrontend) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MicroFrontendList) DeepCopyInto(out *MicroFrontendList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MicroFrontend, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MicroFrontendList.
func (in *MicroFrontendList) DeepCopy() *MicroFrontendList {
	if in == nil {
		return nil
	}
	out := new(MicroFrontendList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MicroFrontendList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MicroFrontendSpec) DeepCopyInto(out *MicroFrontendSpec) {
	*out = *in
	in.CommonMicroFrontendSpec.DeepCopyInto(&out.CommonMicroFrontendSpec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MicroFrontendSpec.
func (in *MicroFrontendSpec) DeepCopy() *MicroFrontendSpec {
	if in == nil {
		return nil
	}
	out := new(MicroFrontendSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NavigationNode) DeepCopyInto(out *NavigationNode) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NavigationNode.
func (in *NavigationNode) DeepCopy() *NavigationNode {
	if in == nil {
		return nil
	}
	out := new(NavigationNode)
	in.DeepCopyInto(out)
	return out
}
