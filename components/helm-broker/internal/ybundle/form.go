package ybundle

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

type form struct {
	Meta  *formMeta
	Plans map[string]*formPlan
}

type formMeta struct {
	ID                  string `yaml:"id"`
	Name                string `yaml:"name"`
	Version             string `yaml:"version"`
	Description         string `yaml:"description"`
	DisplayName         string `yaml:"displayName"`
	Tags                string `yaml:"tags"`
	ProviderDisplayName string `yaml:"providerDisplayName"`
	LongDescription     string `yaml:"longDescription"`
	DocumentationURL    string `yaml:"documentationURL"`
	SupportURL          string `yaml:"supportURL"`
	ImageURL            string `yaml:"imageURL"`
	Bindable            bool   `yaml:"bindable"`
}

func (m *formMeta) Validate() error {
	messages := []string{}

	if m.ID == "" {
		messages = append(messages, "missing ID field")
	}
	if m.Name == "" {
		messages = append(messages, "missing Name field")
	}
	if m.Version == "" {
		messages = append(messages, "missing Version field")
	}
	if m.Description == "" {
		messages = append(messages, "missing Description field")
	}
	if m.DisplayName == "" {
		messages = append(messages, "missing displayName field")
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ", "))
	}

	return nil
}
func (f *form) Validate() error {
	messages := []string{}

	if f.Meta == nil {
		messages = append(messages, fmt.Sprintf("missing metadata information about bundle. Please check if bundle contains %q file", bundleMetaName))
	}
	if len(f.Plans) == 0 {
		messages = append(messages, "bundle does not contains any plans")
	}
	for name, plan := range f.Plans {
		if err := plan.Validate(); err != nil {
			messages = append(messages, fmt.Sprintf("while validating %q plan: %s", name, err.Error()))
		}
	}

	if f.Meta != nil {
		if err := f.Meta.Validate(); err != nil {
			messages = append(messages, fmt.Sprintf("while validating bundle meta: %s", err.Error()))
		}
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ", "))
	}

	return nil
}

func (f *form) ToModel(c *chart.Chart) (internal.Bundle, error) {
	ybVer, err := semver.NewVersion(f.Meta.Version)
	if err != nil {
		return internal.Bundle{}, errors.Wrap(err, "while converting form string version to semver type")
	}

	mappedPlans := make(map[internal.BundlePlanID]internal.BundlePlan)
	for name, plan := range f.Plans {
		dm, err := plan.ToModel(c)
		if err != nil {
			return internal.Bundle{}, errors.Wrapf(err, "while mapping to model %q plan", name)
		}
		mappedPlans[internal.BundlePlanID(plan.Meta.ID)] = dm
	}

	return internal.Bundle{
		ID:          internal.BundleID(f.Meta.ID),
		Name:        internal.BundleName(f.Meta.Name),
		Description: f.Meta.Description,
		Bindable:    f.Meta.Bindable,
		Metadata: internal.BundleMetadata{
			DisplayName:         f.Meta.DisplayName,
			DocumentationURL:    f.Meta.DocumentationURL,
			ImageURL:            f.Meta.ImageURL,
			LongDescription:     f.Meta.LongDescription,
			ProviderDisplayName: f.Meta.ProviderDisplayName,
			SupportURL:          f.Meta.SupportURL,
		},
		Tags:    f.mapTagToModel(),
		Version: *ybVer,
		Plans:   mappedPlans,
	}, nil
}

func (f *form) mapTagToModel() []internal.BundleTag {
	splittedTags := strings.Split(f.Meta.Tags, ",")
	mapped := make([]internal.BundleTag, 0, len(splittedTags))
	for i := range splittedTags {
		mapped = append(mapped, internal.BundleTag(strings.TrimSpace(splittedTags[i])))
	}
	return mapped
}

type formPlan struct {
	Meta          *formPlanMeta
	SchemasUpdate *internal.PlanSchema
	SchemasCreate *internal.PlanSchema
	Values        map[string]interface{}
	BindTemplate  []byte
}

func (p *formPlan) Validate() error {
	if p.Meta == nil {
		return fmt.Errorf("missing metadata information about plan. Please check if plan contains %q file", bundlePlanMetaName)
	}

	if p.Meta.Bindable != nil && *p.Meta.Bindable == true && p.BindTemplate == nil {
		return fmt.Errorf("plans is marked as bindable but %s file was not found in plan %s", bundlePlanBindTemplateFileName, p.Meta.Name)
	}

	if err := p.Meta.Validate(); err != nil {
		return errors.Wrap(err, "while validating plan meta")
	}

	return nil
}

func (p *formPlan) ToModel(c *chart.Chart) (internal.BundlePlan, error) {
	if c == nil {
		return internal.BundlePlan{}, errors.New("missing input param chart")
	}
	if c.Metadata == nil {
		return internal.BundlePlan{}, errors.New("missing Metadata field in input param chart")
	}

	cVer, err := semver.NewVersion(c.Metadata.Version)
	if err != nil {
		return internal.BundlePlan{}, errors.Wrap(err, "while converting chart string version to semver type")
	}

	cRef := internal.ChartRef{
		Name:    internal.ChartName(c.Metadata.Name),
		Version: *cVer,
	}

	mappedSchemas := make(map[internal.PlanSchemaType]internal.PlanSchema)

	if p.SchemasUpdate != nil {
		mappedSchemas[internal.SchemaTypeUpdate] = *p.SchemasUpdate
	}
	if p.SchemasCreate != nil {
		mappedSchemas[internal.SchemaTypeProvision] = *p.SchemasCreate
	}

	return internal.BundlePlan{
		ID:          internal.BundlePlanID(p.Meta.ID),
		Name:        internal.BundlePlanName(p.Meta.Name),
		Description: p.Meta.Description,
		Metadata: internal.BundlePlanMetadata{
			DisplayName: p.Meta.DisplayName,
		},
		ChartValues:  internal.ChartValues(p.Values),
		Schemas:      mappedSchemas,
		ChartRef:     cRef,
		Bindable:     p.Meta.Bindable,
		BindTemplate: p.BindTemplate,
	}, nil
}

type formPlanMeta struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	DisplayName string `yaml:"displayName"`
	Bindable    *bool  `yaml:"bindable"`
}

func (f *formPlanMeta) Validate() error {
	messages := []string{}
	if f.ID == "" {
		messages = append(messages, "missing ID field")
	}
	if f.Name == "" {
		messages = append(messages, "missing Name field")
	}
	if f.Description == "" {
		messages = append(messages, "missing Description field")
	}
	if f.DisplayName == "" {
		messages = append(messages, "missing displayName field")
	}
	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ", "))
	}
	return nil
}
