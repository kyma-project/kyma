package broker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	osb "github.com/pmorie/go-open-service-broker-client/v2"

	"github.com/kyma-project/kyma/components/helm-broker/platform/ptr"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker/automock"
)

func TestGetCatalog(t *testing.T) {
	// GIVEN
	tc := newCatalogTC()
	defer tc.AssertExpectations(t)
	tc.finderMock.On("FindAll").Return(tc.fixBundles(), nil).Once()
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
	tc.finderMock.On("FindAll").Return(nil, tc.fixError()).Once()
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

	tc.finderMock.On("FindAll").Return(tc.fixBundles(), nil).Once()
	tc.converterMock.On("Convert", tc.fixBundle()).Return(osb.Service{}, tc.fixError())

	svc := broker.NewCatalogService(tc.finderMock, tc.converterMock)
	osbCtx := broker.NewOSBContext("not", "important")
	// WHEN
	_, err := svc.GetCatalog(context.Background(), *osbCtx)
	// THEN
	assert.EqualError(t, err, fmt.Sprintf("while converting bundle to service: %v", tc.fixError()))

}

func TestBundleConversion(t *testing.T) {
	// GIVEN
	tc := newCatalogTC()
	fixBundle := tc.fixBundle()
	conv := broker.NewConverter()
	// WHEN
	s, err := conv.Convert(fixBundle)
	// THEN
	assert.NoError(t, err)
	assert.Equal(t, "bundleID", s.ID)
	assert.Equal(t, "bundleName", s.Name)
	assert.Equal(t, "bundleDescription", s.Description)
	assert.True(t, s.Bindable)
	assert.Equal(t, map[string]interface{}{
		"displayName":         fixBundle.Metadata.DisplayName,
		"providerDisplayName": fixBundle.Metadata.ProviderDisplayName,
		"longDescription":     fixBundle.Metadata.LongDescription,
		"documentationURL":    fixBundle.Metadata.DocumentationURL,
		"supportURL":          fixBundle.Metadata.SupportURL,
		"imageURL":            fixBundle.Metadata.ImageURL,
	}, s.Metadata)

	require.Len(t, s.Plans, 2)

	var p1, p2 osb.Plan
	if s.Plans[0].ID == "planID1" {
		p1 = s.Plans[0]
		p2 = s.Plans[1]
	} else {
		p2 = s.Plans[0]
		p1 = s.Plans[1]
	}

	assert.Equal(t, "planID1", p1.ID)
	assert.Equal(t, "plan1Description", p1.Description)
	assert.Equal(t, "plan1Name", p1.Name)
	require.NotNil(t, p1.Bindable)
	assert.True(t, *p1.Bindable)
	assert.Equal(t, tc.fixProvisionSchema(), p1.ParameterSchemas.ServiceInstances.Create.Parameters)
	assert.Equal(t, tc.fixUpdateSchema(), p1.ParameterSchemas.ServiceInstances.Update.Parameters)
	assert.Equal(t, tc.fixBindSchema(), p1.ParameterSchemas.ServiceBindings.Create.Parameters)
	assert.Equal(t, map[string]interface{}{
		"displayName": fixBundle.Plans["planID1"].Metadata.DisplayName,
	}, p1.Metadata)

	assert.Equal(t, "planID2", p2.ID)
	assert.Equal(t, "plan2Description", p2.Description)
	assert.Equal(t, "plan2Name", p2.Name)
	require.NotNil(t, p2.Bindable)
	assert.True(t, *p2.Bindable)
	assert.Nil(t, p2.ParameterSchemas.ServiceInstances.Create.Parameters)
	assert.Nil(t, p2.ParameterSchemas.ServiceInstances.Update.Parameters)
	assert.Nil(t, p2.ParameterSchemas.ServiceBindings.Create.Parameters)
	assert.Equal(t, map[string]interface{}{
		"displayName": fixBundle.Plans["planID2"].Metadata.DisplayName,
	}, p2.Metadata)
	assert.Equal(t, []string{"awesome-tag"}, s.Tags)
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
			ImageURL:            "ImageURL",
		},
		Tags: []internal.BundleTag{"awesome-tag"},
		Plans: map[internal.BundlePlanID]internal.BundlePlan{
			"planID1": {
				ID:          "planID1",
				Description: "plan1Description",
				Name:        "plan1Name",
				Schemas: map[internal.PlanSchemaType]internal.PlanSchema{
					internal.SchemaTypeProvision: tc.fixProvisionSchema(),
					internal.SchemaTypeUpdate:    tc.fixUpdateSchema(),
					internal.SchemaTypeBind:      tc.fixBindSchema(),
				},
				Metadata: internal.BundlePlanMetadata{
					DisplayName: "displayName-1",
				},
				Bindable: ptr.Bool(true),
			},
			"planID2": {
				ID:          "planID2",
				Description: "plan2Description",
				Name:        "plan2Name",
				Bindable:    ptr.Bool(true),
			},
		},
	}
}

func (tc catalogTestCase) fixProvisionSchema() internal.PlanSchema {
	return internal.PlanSchema{}
}

func (tc catalogTestCase) fixUpdateSchema() internal.PlanSchema {
	return internal.PlanSchema{}
}

func (tc catalogTestCase) fixBindSchema() internal.PlanSchema {
	return internal.PlanSchema{}
}

func (tc catalogTestCase) fixService() osb.Service {
	return osb.Service{ID: "bundleID"}
}

func (tc catalogTestCase) fixError() error {
	return errors.New("some error")
}
