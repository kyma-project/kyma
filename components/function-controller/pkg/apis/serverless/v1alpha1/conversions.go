package v1alpha1

import (
	v1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	conversion "k8s.io/apimachinery/pkg/conversion"
)

func Convert_v1alpha2_FunctionSpec_To_v1alpha1_FunctionSpec(in *v1alpha2.FunctionSpec, out *FunctionSpec, s conversion.Scope) error {
	return nil
}
func Convert_v1alpha1_FunctionSpec_To_v1alpha2_FunctionSpec(in *FunctionSpec, out *v1alpha2.FunctionSpec, s conversion.Scope) error {
	return nil
}
func Convert_v1alpha1_FunctionStatus_To_v1alpha2_FunctionStatus(in *FunctionStatus, out *v1alpha2.FunctionStatus, s conversion.Scope) error {
	return nil
}
