package ybundle

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

func TestFormToModelSuccess(t *testing.T) {
	// given
	fixForm := fixValidForm()
	fixForm.Plans = nil // Already tested by TestFormPlanToModelSuccess, triggering such flow here, will be difficult (postponed)

	fixChart := fixValidChart()

	ver, err := semver.NewVersion(fixForm.Meta.Version)
	require.NoError(t, err)

	expBundle := internal.Bundle{
		ID:          internal.BundleID(fixForm.Meta.ID),
		Name:        internal.BundleName(fixForm.Meta.Name),
		Description: fixForm.Meta.Description,
		Metadata: internal.BundleMetadata{
			DisplayName:         fixForm.Meta.DisplayName,
			DocumentationURL:    fixForm.Meta.DocumentationURL,
			ImageURL:            fixForm.Meta.ImageURL,
			LongDescription:     fixForm.Meta.LongDescription,
			ProviderDisplayName: fixForm.Meta.ProviderDisplayName,
			SupportURL:          fixForm.Meta.SupportURL,
		},
		Tags:     []internal.BundleTag{"go", "golang"},
		Bindable: fixForm.Meta.Bindable,
		Version:  *ver,
		Plans:    make(map[internal.BundlePlanID]internal.BundlePlan),
	}
	// when
	gotBundle, err := fixForm.ToModel(&fixChart)

	// then
	require.NoError(t, err)
	assert.Equal(t, expBundle, gotBundle)
}

func TestFormToModelFailure(t *testing.T) {
	// given
	fixChart := fixValidChart()

	fixForm := fixValidForm()
	fixForm.Meta.Version = "abc"

	// when
	gotBundle, err := fixForm.ToModel(&fixChart)

	// then
	require.EqualError(t, err, "while converting form string version to semver type: Invalid Semantic Version")
	assert.Zero(t, gotBundle)
}

func TestFormValidateSuccess(t *testing.T) {
	for tn, tc := range map[string]struct {
		fixForm form
	}{
		"all fields provided": {
			fixForm: fixValidForm(),
		},
		"not required fields are empty": {
			fixForm: func() form {
				f := fixValidForm()
				// change to zero values
				f.Meta.Tags = ""
				f.Meta.ProviderDisplayName = ""
				f.Meta.LongDescription = ""
				f.Meta.DocumentationURL = ""
				f.Meta.SupportURL = ""
				f.Meta.ImageURL = ""
				f.Meta.Bindable = false
				return f
			}(),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// when
			err := tc.fixForm.Validate()

			// then
			assert.NoError(t, err)
		})
	}
}

func TestFormValidateFailure(t *testing.T) {
	for tn, tc := range map[string]struct {
		fixForm form
		errMsgs []string
	}{
		"missing form meta field": {
			fixForm: func() form {
				fix := fixValidForm()
				fix.Meta = nil
				return fix
			}(),
			errMsgs: []string{"missing metadata information about bundle. Please check if bundle contains \"meta.yaml\" file"},
		},
		"missing form ID field": {
			fixForm: func() form {
				fix := fixValidForm()
				fix.Meta.ID = ""
				return fix
			}(),
			errMsgs: []string{"while validating bundle meta: missing ID field"},
		},
		"missing form Name field": {
			fixForm: func() form {
				fix := fixValidForm()
				fix.Meta.Name = ""
				return fix
			}(),
			errMsgs: []string{"while validating bundle meta: missing Name field"},
		},
		"missing form Version field": {
			fixForm: func() form {
				fix := fixValidForm()
				fix.Meta.Version = ""
				return fix
			}(),
			errMsgs: []string{"while validating bundle meta: missing Version field"},
		},
		"missing form Description field": {
			fixForm: func() form {
				fix := fixValidForm()
				fix.Meta.Description = ""
				return fix
			}(),
			errMsgs: []string{"while validating bundle meta: missing Description field"},
		},
		"missing form displayName field": {
			fixForm: func() form {
				fix := fixValidForm()
				fix.Meta.DisplayName = ""
				return fix
			}(),
			errMsgs: []string{"while validating bundle meta: missing displayName field"},
		},
		"invalid form plan entry": {
			fixForm: func() form {
				fix := fixValidForm()
				for k := range fix.Plans { // remove meta from plans
					fix.Plans[k].Meta = nil
				}
				return fix
			}(),
			errMsgs: []string{"while validating \"micro-plan-id-123\" plan: missing metadata information about plan. Please check if plan contains \"meta.yaml\" file",
				"while validating \"myk-plan-id-123\" plan: missing metadata information about plan. Please check if plan contains \"meta.yaml\" file"},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// when
			err := tc.fixForm.Validate()

			// then
			for _, msg := range tc.errMsgs {
				assert.Contains(t, err.Error(), msg)
			}
		})
	}
}

func TestFormPlanToModelSuccess(t *testing.T) {
	// given
	fixPlan := fixValidFormPlan("test-to-model-success")
	fixChart := fixValidChart()

	charVer, err := semver.NewVersion(fixValidChart().Metadata.Version)
	require.NoError(t, err)

	expBundlePlan := internal.BundlePlan{
		ID:          internal.BundlePlanID(fixPlan.Meta.ID),
		Name:        internal.BundlePlanName(fixPlan.Meta.Name),
		Description: fixPlan.Meta.Description,
		Metadata: internal.BundlePlanMetadata{
			DisplayName: fixPlan.Meta.DisplayName,
		},
		Schemas: map[internal.PlanSchemaType]internal.PlanSchema{
			internal.SchemaTypeProvision: *fixPlan.SchemasCreate,
		},
		ChartRef: internal.ChartRef{
			Name:    internal.ChartName(fixChart.Metadata.Name),
			Version: *charVer,
		},
		ChartValues:  fixPlan.Values,
		Bindable:     fixPlan.Meta.Bindable,
		BindTemplate: fixPlan.BindTemplate,
	}

	// when
	gotBundlePlan, err := fixPlan.ToModel(&fixChart)

	// then
	require.NoError(t, err)
	assert.Equal(t, expBundlePlan, gotBundlePlan)
}

func TestFormPlanToModelFailure(t *testing.T) {
	t.Run("on input chart param", func(t *testing.T) {
		for tn, tc := range map[string]struct {
			errMsg   string
			fixChart *chart.Chart
		}{
			"nil input chart": {
				fixChart: nil,
				errMsg:   "missing input param chart",
			},
			"mising chart metadata": {
				fixChart: func() *chart.Chart {
					fixChart := fixValidChart()
					fixChart.Metadata = nil
					return &fixChart
				}(),
				errMsg: "missing Metadata field in input param chart",
			},
			"invalid chart version": {
				fixChart: func() *chart.Chart {
					fixChart := fixValidChart()
					fixChart.Metadata.Version = "abc"
					return &fixChart
				}(),
				errMsg: "while converting chart string version to semver type: Invalid Semantic Version",
			},
		} {
			t.Run(tn, func(t *testing.T) {
				// given
				fixPlan := fixValidFormPlan("test-to-model-success")

				// when
				gotBundlePlan, err := fixPlan.ToModel(tc.fixChart)

				// then
				require.EqualError(t, err, tc.errMsg)
				assert.Zero(t, gotBundlePlan)
			})
		}
	})
}

func TestFormPlanValidateSuccess(t *testing.T) {
	// given
	fixFormPlan := fixValidFormPlan("valid-success-test")

	// when
	err := fixFormPlan.Validate()

	// then
	assert.NoError(t, err)
}

func TestFormPlanValidateFailure(t *testing.T) {
	for tn, tc := range map[string]struct {
		fixFormPlan formPlan
		errMsg      string
	}{
		"missing meta field": {
			fixFormPlan: func() formPlan {
				fix := fixValidFormPlan("missing-fields")
				fix.Meta = nil
				return fix
			}(),
			errMsg: "missing metadata information about plan. Please check if plan contains \"meta.yaml\" file",
		},
		"missing ID field": {
			fixFormPlan: func() formPlan {
				fix := fixValidFormPlan("missing-fields")
				fix.Meta.ID = ""
				return fix
			}(),
			errMsg: "while validating plan meta: missing ID field",
		},
		"missing Name field": {
			fixFormPlan: func() formPlan {
				fix := fixValidFormPlan("missing-fields")
				fix.Meta.Name = ""
				return fix
			}(),
			errMsg: "while validating plan meta: missing Name field",
		},
		"missing Description field": {
			fixFormPlan: func() formPlan {
				fix := fixValidFormPlan("missing-fields")
				fix.Meta.Description = ""
				return fix
			}(),
			errMsg: "while validating plan meta: missing Description field",
		},
		"missing displayName field": {
			fixFormPlan: func() formPlan {
				fix := fixValidFormPlan("missing-fields")
				fix.Meta.DisplayName = ""
				return fix
			}(),
			errMsg: "while validating plan meta: missing displayName field",
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// when
			err := tc.fixFormPlan.Validate()

			// then
			assert.EqualError(t, err, tc.errMsg)
		})
	}
}

func fixValidForm() form {
	micro := fixValidFormPlan("micro-plan-id-123")
	myk := fixValidFormPlan("myk-plan-id-123")

	return form{
		Meta: &formMeta{
			ID:                  "123-id-123",
			Description:         "form-desc",
			Name:                "form-name",
			Version:             "1.0.1",
			SupportURL:          "http://support.url",
			ProviderDisplayName: "Gopherek Inc.",
			LongDescription:     "Gopher Gopherowi Idefixem",
			ImageURL:            "http://image.url",
			DocumentationURL:    "http://documentation.url",
			DisplayName:         "Gopher Form",
			Tags:                "go, golang",
			Bindable:            true,
		},
		Plans: map[string]*formPlan{
			micro.Meta.ID: &micro,
			myk.Meta.ID:   &myk,
		},
	}
}

func fixValidFormPlan(id string) formPlan {
	return formPlan{
		Meta: &formPlanMeta{
			ID:          id,
			Name:        "name",
			Description: "desc",
			DisplayName: "Plan Display Name",
		},
		SchemasCreate: &internal.PlanSchema{},
		Values: map[string]interface{}{
			"par1": "val1",
		},
		BindTemplate: []byte(`bindTemplate`),
	}
}

func fixValidChart() chart.Chart {
	return chart.Chart{
		Metadata: &chart.Metadata{
			Version: "9.9.9",
			Name:    "test-chart",
		},
	}
}
