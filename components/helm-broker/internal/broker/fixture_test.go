package broker_test

import (
	"fmt"

	"github.com/Masterminds/semver"

	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

type expAll struct {
	InstanceID  internal.InstanceID
	OperationID internal.OperationID
	Bundle      struct {
		ID       internal.BundleID
		Version  semver.Version
		Name     internal.BundleName
		Bindable bool
	}
	BundlePlan struct {
		ID           internal.BundlePlanID
		Name         internal.BundlePlanName
		DisplayName  string
		BindTemplate internal.BundlePlanBindTemplate
	}
	Chart struct {
		Name    internal.ChartName
		Version semver.Version
	}
	Service struct {
		ID internal.ServiceID
	}
	ServicePlan struct {
		ID internal.ServicePlanID
	}
	Namespace   internal.Namespace
	ReleaseName internal.ReleaseName
	ParamsHash  string
}

func (exp *expAll) Populate() {
	exp.InstanceID = internal.InstanceID("fix-I-ID")
	exp.OperationID = internal.OperationID("fix-OP-ID")

	exp.Bundle.ID = internal.BundleID("fix-B-ID")
	exp.Bundle.Version = *semver.MustParse("0.1.2")
	exp.Bundle.Name = internal.BundleName("fix-B-Name")
	exp.Bundle.Bindable = true

	exp.BundlePlan.ID = internal.BundlePlanID("fix-P-ID")
	exp.BundlePlan.Name = internal.BundlePlanName("fix-P-Name")
	exp.BundlePlan.DisplayName = "fix-P-DisplayName"
	exp.BundlePlan.BindTemplate = internal.BundlePlanBindTemplate("template")

	exp.Chart.Name = internal.ChartName("fix-C-Name")
	exp.Chart.Version = *semver.MustParse("1.2.3")

	exp.Service.ID = internal.ServiceID(exp.Bundle.ID)
	exp.ServicePlan.ID = internal.ServicePlanID(exp.BundlePlan.ID)

	exp.Namespace = internal.Namespace("fix-namespace")
	exp.ReleaseName = internal.ReleaseName(fmt.Sprintf("hb-%s-%s-%s", exp.Bundle.Name, exp.BundlePlan.Name, exp.InstanceID))
	exp.ParamsHash = "TODO"
}

func (exp *expAll) NewChart() *chart.Chart {
	return &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    string(exp.Chart.Name),
			Version: exp.Chart.Version.String(),
		},
	}
}

func (exp *expAll) NewBundle() *internal.Bundle {
	return &internal.Bundle{
		ID:       exp.Bundle.ID,
		Version:  exp.Bundle.Version,
		Name:     exp.Bundle.Name,
		Bindable: exp.Bundle.Bindable,
		Plans: map[internal.BundlePlanID]internal.BundlePlan{
			exp.BundlePlan.ID: {
				ID:   exp.BundlePlan.ID,
				Name: exp.BundlePlan.Name,
				ChartRef: internal.ChartRef{
					Name:    exp.Chart.Name,
					Version: exp.Chart.Version,
				},
				Metadata: internal.BundlePlanMetadata{
					DisplayName: exp.BundlePlan.DisplayName,
				},
				BindTemplate: exp.BundlePlan.BindTemplate,
			},
		},
	}
}

func (exp *expAll) NewInstance() *internal.Instance {
	return &internal.Instance{
		ID:            exp.InstanceID,
		ServiceID:     exp.Service.ID,
		ServicePlanID: exp.ServicePlan.ID,
		ReleaseName:   exp.ReleaseName,
		Namespace:     exp.Namespace,
		ParamsHash:    exp.ParamsHash,
	}
}

func (exp *expAll) NewInstanceOperation(tpe internal.OperationType, state internal.OperationState) *internal.InstanceOperation {
	return &internal.InstanceOperation{
		InstanceID:  exp.InstanceID,
		OperationID: exp.OperationID,
		Type:        tpe,
		State:       state,
		ParamsHash:  exp.ParamsHash,
	}
}
