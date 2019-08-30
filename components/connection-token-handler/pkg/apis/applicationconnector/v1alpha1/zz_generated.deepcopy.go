// +build !ignore_autogenerated

/*
.
*/
// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterContext) DeepCopyInto(out *ClusterContext) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterContext.
func (in *ClusterContext) DeepCopy() *ClusterContext {
	if in == nil {
		return nil
	}
	out := new(ClusterContext)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TokenRequest) DeepCopyInto(out *TokenRequest) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Context = in.Context
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TokenRequest.
func (in *TokenRequest) DeepCopy() *TokenRequest {
	if in == nil {
		return nil
	}
	out := new(TokenRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TokenRequest) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TokenRequestList) DeepCopyInto(out *TokenRequestList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TokenRequest, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TokenRequestList.
func (in *TokenRequestList) DeepCopy() *TokenRequestList {
	if in == nil {
		return nil
	}
	out := new(TokenRequestList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TokenRequestList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TokenRequestStatus) DeepCopyInto(out *TokenRequestStatus) {
	*out = *in
	in.ExpireAfter.DeepCopyInto(&out.ExpireAfter)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TokenRequestStatus.
func (in *TokenRequestStatus) DeepCopy() *TokenRequestStatus {
	if in == nil {
		return nil
	}
	out := new(TokenRequestStatus)
	in.DeepCopyInto(out)
	return out
}
