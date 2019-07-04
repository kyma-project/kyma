package broker_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/alecthomas/jsonschema"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker/automock"
	"github.com/kyma-project/kyma/components/helm-broker/platform/ptr"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden (.out) files")

func TestGetCatalog(t *testing.T) {
	// GIVEN
	tc := newCatalogTC()
	defer tc.AssertExpectations(t)
	tc.finderMock.On("FindAll", internal.ClusterWide).Return(tc.fixBundles(), nil).Once()
	tc.converterMock.On("Convert", tc.fixBundle()).Return(tc.fixService(), nil)

	svc := broker.NewCatalogService(tc.finderMock, tc.converterMock)
	osbCtx := broker.NewOSBContext("not", "important")
	// WHEN
	resp, err := svc.GetCatalog(context.Background(), *osbCtx)
	// THEN
	assert.Nil(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Services, 1)
	assert.Equal(t, tc.fixService(), resp.Services[0])

}

func TestGetCatalogOnFindError(t *testing.T) {
	// GIVEN
	tc := newCatalogTC()
	defer tc.AssertExpectations(t)
	tc.finderMock.On("FindAll", internal.ClusterWide).Return(nil, tc.fixError()).Once()
	svc := broker.NewCatalogService(tc.finderMock, nil)
	osbCtx := broker.NewOSBContext("not", "important")
	// WHEN
	_, err := svc.GetCatalog(context.Background(), *osbCtx)
	// THEN
	assert.EqualError(t, err, fmt.Sprintf("while finding all bundles: %v", tc.fixError()))

}

func TestGetCatalogOnConversionError(t *testing.T) {
	// GIVEN
	tc := newCatalogTC()
	defer tc.AssertExpectations(t)

	tc.finderMock.On("FindAll", internal.ClusterWide).Return(tc.fixBundles(), nil).Once()
	tc.converterMock.On("Convert", tc.fixBundle()).Return(osb.Service{}, tc.fixError())

	svc := broker.NewCatalogService(tc.finderMock, tc.converterMock)
	osbCtx := broker.NewOSBContext("not", "important")
	// WHEN
	_, err := svc.GetCatalog(context.Background(), *osbCtx)
	// THEN
	assert.EqualError(t, err, fmt.Sprintf("while converting bundle to service: %v", tc.fixError()))

}

func TestBundleConversion(t *testing.T) {
	tests := map[string]struct {
		fixSchemas map[internal.PlanSchemaType]internal.PlanSchema

		expGoldenName string
	}{
		"empty schemas": {
			fixSchemas: nil,

			expGoldenName: "TestBundleConversion-without-schemas.golden.json",
		},
		"schemas provided": {
			fixSchemas: map[internal.PlanSchemaType]internal.PlanSchema{
				internal.SchemaTypeProvision: fixProvisionSchema(),
				internal.SchemaTypeUpdate:    fixUpdateSchema(),
				internal.SchemaTypeBind:      fixBindSchema(),
			},

			expGoldenName: "TestBundleConversion-with-schemas.golden.json",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			tc := newCatalogTC()
			goldenPath := filepath.Join("testdata", test.expGoldenName)

			conv := broker.NewConverter()

			// WHEN
			convertedSvc, err := conv.Convert(tc.fixBundleWithSchemas(test.fixSchemas))

			// THEN
			require.NoError(t, err)

			normalizedGotSvc := tc.marshal(t, convertedSvc)

			updateGoldenFileIfRequested(t, goldenPath, normalizedGotSvc)

			exp := tc.fixtureMarshaledOsbService(t, goldenPath)
			assert.JSONEq(t, exp, string(normalizedGotSvc))
		})
	}
}

func TestBundleConversionOverridesLocalLabel(t *testing.T) {
	// GIVEN
	tc := newCatalogTC()
	conv := broker.NewConverter()
	fixBundle := tc.fixBundle()
	fixBundle.Metadata.Labels["local"] = "false"

	// WHEN
	convertedSvc, err := conv.Convert(fixBundle)

	// THEN
	require.NoError(t, err)

	gotLabels, ok := convertedSvc.Metadata["labels"].(internal.Labels)
	require.True(t, ok, "cannot cast metadata labels to internal.Labels")
	assert.Equal(t, "true", gotLabels["local"])
}

type catalogTestCase struct {
	finderMock    *automock.BundleStorage
	converterMock *automock.Converter
}

func newCatalogTC() *catalogTestCase {
	return &catalogTestCase{
		finderMock:    &automock.BundleStorage{},
		converterMock: &automock.Converter{},
	}
}

func (tc catalogTestCase) AssertExpectations(t *testing.T) {
	tc.finderMock.AssertExpectations(t)
	tc.converterMock.AssertExpectations(t)
}

func (tc catalogTestCase) fixBundles() []*internal.Bundle {
	return []*internal.Bundle{tc.fixBundle()}
}

func (tc catalogTestCase) fixBundle() *internal.Bundle {
	return tc.fixBundleWithSchemas(tc.fixPlanSchemas())
}

func (tc catalogTestCase) fixBundleWithSchemas(schemas map[internal.PlanSchemaType]internal.PlanSchema) *internal.Bundle {
	return &internal.Bundle{
		Name:        "bundleName",
		ID:          "bundleID",
		Description: "bundleDescription",
		Bindable:    true,
		Version:     *semver.MustParse("1.2.3"),
		Metadata: internal.BundleMetadata{
			DisplayName:         "DisplayName",
			ProviderDisplayName: "ProviderDisplayName",
			LongDescription:     "LongDescription",
			DocumentationURL:    "DocumentationURL",
			SupportURL:          "SupportURL",
			ProvisionOnlyOnce:   true,
			ImageURL:            "ImageURL",
			Labels: internal.Labels{
				"testing":           "true",
				"provisionOnlyOnce": "true",
			},
		},
		Tags: []internal.BundleTag{"awesome-tag"},
		Plans: map[internal.BundlePlanID]internal.BundlePlan{
			"planID": {
				ID:          "planID",
				Description: "planDescription",
				Name:        "planName",
				Metadata: internal.BundlePlanMetadata{
					DisplayName: "displayName-1",
				},
				Bindable: ptr.Bool(true),
				Schemas:  schemas,
			},
		},
	}
}

func (tc catalogTestCase) fixPlanSchemas() map[internal.PlanSchemaType]internal.PlanSchema {
	return map[internal.PlanSchemaType]internal.PlanSchema{
		internal.SchemaTypeProvision: fixProvisionSchema(),
		internal.SchemaTypeUpdate:    fixUpdateSchema(),
		internal.SchemaTypeBind:      fixBindSchema(),
	}
}

func fixProvisionSchema() internal.PlanSchema {
	return internal.PlanSchema{
		Type: &jsonschema.Type{
			Version: "http://json-schema.org/draft-04/schema#",
			Type:    "string",
			Title:   "ProvisionSchema",
		},
	}
}

func fixUpdateSchema() internal.PlanSchema {
	return internal.PlanSchema{
		Type: &jsonschema.Type{
			Version: "http://json-schema.org/draft-04/schema#",
			Type:    "string",
			Title:   "UpdateSchema",
		},
	}
}

func fixBindSchema() internal.PlanSchema {
	return internal.PlanSchema{
		Type: &jsonschema.Type{
			Version: "http://json-schema.org/draft-04/schema#",
			Type:    "string",
			Title:   "BindSchema",
		},
	}
}

func (tc catalogTestCase) fixService() osb.Service {
	return osb.Service{ID: "bundleID"}
}

func (tc catalogTestCase) fixError() error {
	return errors.New("some error")
}

func (tc catalogTestCase) marshal(t *testing.T, in interface{}) []byte {
	t.Helper()
	out, err := json.Marshal(in)
	require.NoError(t, err)
	return out
}

func (tc catalogTestCase) fixtureMarshaledOsbService(t *testing.T, testdataBasePath string) string {
	t.Helper()
	data, err := ioutil.ReadFile(testdataBasePath)
	require.NoError(t, err, "failed reading .golden")

	return string(data)
}

func updateGoldenFileIfRequested(t *testing.T, goldenPath string, obj []byte) {
	t.Helper()
	if *update {
		t.Log("update golden file")
		err := ioutil.WriteFile(goldenPath, obj, 0644)
		require.NoError(t, err, "failed to update golden file")
	}
}
