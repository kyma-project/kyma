package repository

import (
	"fmt"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

type AddonController struct {
	ID     string
	Addon  v1alpha1.Addon
	bundle *internal.Bundle
	chart  []*chart.Chart
}

func NewAddon(n, v string) *AddonController {
	return &AddonController{
		Addon: v1alpha1.Addon{Name: n, Version: v},
	}
}

func (a *AddonController) SetID(ID string) {
	a.ID = ID
}

func (a *AddonController) AddBundle(b *internal.Bundle) {
	a.bundle = b
}

func (a *AddonController) AddCharts(ch []*chart.Chart) {
	a.chart = ch
}
func (a *AddonController) IsReady() bool {
	return a.Addon.Status == v1alpha1.AddonStatusReady
}

func (a *AddonController) Failed() {
	a.Addon.Status = v1alpha1.AddonStatusFailed
}

func (a *AddonController) Ready() {
	a.Addon.Status = v1alpha1.AddonStatusReady
}

func (a *AddonController) SetAddonFailedInfo(reason v1alpha1.AddonStatusReason, message string) {
	a.Addon.Reason = reason
	a.Addon.Message = fmt.Sprintf(reason.Message(), message)
}
