package rafter

import (
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

const (
	ModuleName                                    = "rafter"
	ViewContextLabel                              = "rafter.kyma-project.io/view-context"
	GroupNameLabel                                = "rafter.kyma-project.io/group-name"
	OrderLabel                                    = "rafter.kyma-project.io/order"
	SourceType       v1beta1.AssetGroupSourceType = "markdown"
	SourceName       v1beta1.AssetGroupSourceName = "markdown"
	MockiceSvcName   string                       = "cbs-rafter-test-svc"
	MockiceNamespace string                       = "kyma-system"
)
