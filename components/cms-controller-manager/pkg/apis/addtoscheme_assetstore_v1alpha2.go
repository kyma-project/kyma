package apis

import (
	api "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
)

func init() {
	// https://github.com/kubernetes-sigs/kubebuilder/blob/master/docs/using_an_external_type.md
	AddToSchemes = append(AddToSchemes, api.AddToScheme)
}
