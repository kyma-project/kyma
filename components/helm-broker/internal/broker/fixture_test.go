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
	Addon       struct {
		ID            internal.AddonID
		Version       semver.Version
		Name          internal.AddonName
		Bindable      bool
		RepositoryURL string
	}
	AddonPlan struct {
		ID           internal.AddonPlanID
		Name         internal.AddonPlanName
		DisplayName  string
		BindTemplate internal.AddonPlanBindTemplate
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

	exp.Addon.ID = internal.AddonID("fix-B-ID")
	exp.Addon.Version = *semver.MustParse("0.1.2")
	exp.Addon.Name = internal.AddonName("fix-B-Name")
	exp.Addon.Bindable = true
	exp.Addon.RepositoryURL = "fix-url"

	exp.AddonPlan.ID = internal.AddonPlanID("fix-P-ID")
	exp.AddonPlan.Name = internal.AddonPlanName("fix-P-Name")
	exp.AddonPlan.DisplayName = "fix-P-DisplayName"
	exp.AddonPlan.BindTemplate = internal.AddonPlanBindTemplate("template")

	exp.Chart.Name = internal.ChartName("fix-C-Name")
	exp.Chart.Version = *semver.MustParse("1.2.3")

	exp.Service.ID = internal.ServiceID(exp.Addon.ID)
	exp.ServicePlan.ID = internal.ServicePlanID(exp.AddonPlan.ID)

	exp.Namespace = internal.Namespace("fix-namespace")
	exp.ReleaseName = internal.ReleaseName(fmt.Sprintf("hb-%s-%s-%s", exp.Addon.Name, exp.AddonPlan.Name, exp.InstanceID))
	exp.ParamsHash = "TODO"
}

func (exp *expAll) NewInstanceCollection() []*internal.Instance {
	return []*internal.Instance{
		&internal.Instance{
			ServiceID: "new-id-not-exist-0",
			Namespace: "fix-namespace",
		},
		&internal.Instance{
			ServiceID: "new-id-not-exist-1",
			Namespace: "fix-namespace",
		},
		&internal.Instance{
			ServiceID: "new-id-not-exist-2",
			Namespace: "fix-namespace",
		},
	}
}

func (exp *expAll) NewChart() *chart.Chart {
	return &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    string(exp.Chart.Name),
			Version: exp.Chart.Version.String(),
		},
	}
}

func (exp *expAll) NewAddon() *internal.Addon {
	return &internal.Addon{
		ID:            exp.Addon.ID,
		Version:       exp.Addon.Version,
		Name:          exp.Addon.Name,
		Bindable:      exp.Addon.Bindable,
		RepositoryURL: exp.Addon.RepositoryURL,
		Plans: map[internal.AddonPlanID]internal.AddonPlan{
			exp.AddonPlan.ID: {
				ID:   exp.AddonPlan.ID,
				Name: exp.AddonPlan.Name,
				ChartRef: internal.ChartRef{
					Name:    exp.Chart.Name,
					Version: exp.Chart.Version,
				},
				Metadata: internal.AddonPlanMetadata{
					DisplayName: exp.AddonPlan.DisplayName,
				},
				BindTemplate: exp.AddonPlan.BindTemplate,
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
