package broker_test

import (
	"context"
	"fmt"
	"testing"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker/automock"
)

func TestGetCatalogHappyPath(t *testing.T) {
	// GIVEN
	tc := newCatalogTC()
	defer tc.AssertExpectations(t)
	tc.finderMock.On("FindAll").Return([]*internal.RemoteEnvironment{tc.fixRE()}, nil).Once()
	tc.reEnabledCheckerMock.On("IsRemoteEnvironmentEnabled", "stage", string(tc.fixRE().Name)).Return(true, nil)
	tc.converterMock.On("Convert", tc.fixRE().Name, tc.fixRE().Services[0]).Return(tc.fixService(), nil)

	svc := broker.NewCatalogService(tc.finderMock, tc.reEnabledCheckerMock, tc.converterMock)
	osbCtx := broker.NewOSBContext("not", "important", "stage")

	// WHEN
	resp, err := svc.GetCatalog(context.Background(), *osbCtx)

	// THEN
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Services, 1)
	assert.Equal(t, tc.fixService(), resp.Services[0])
}

func TestGetCatalogNotEnabled(t *testing.T) {
	// GIVEN
	tc := newCatalogTC()
	defer tc.AssertExpectations(t)
	tc.finderMock.On("FindAll").Return([]*internal.RemoteEnvironment{tc.fixRE()}, nil).Once()
	tc.reEnabledCheckerMock.On("IsRemoteEnvironmentEnabled", "stage", string(tc.fixRE().Name)).Return(false, nil)

	svc := broker.NewCatalogService(tc.finderMock, tc.reEnabledCheckerMock, tc.converterMock)
	osbCtx := broker.NewOSBContext("not", "important", "stage")

	// WHEN
	resp, err := svc.GetCatalog(context.Background(), *osbCtx)

	// THEN
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Services, 0)
}

func TestConvertService(t *testing.T) {
	const fixReName = "fix-re-name"

	for tn, tc := range map[string]struct {
		givenService func() internal.Service

		expectedService func() osb.Service
	}{
		"simpleAPIBasedService": {
			givenService: func() internal.Service {
				svc := fixAPIBasedService()
				svc.DisplayName = "*Service Name\ną-'#$\tÜ"
				return svc
			},
			expectedService: func() osb.Service {
				svc := fixOsbService()
				svc.Metadata["displayName"] = "*Service Name\ną-'#$\tÜ"
				return svc
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			converter := broker.NewConverter()

			// when
			result, err := converter.Convert(fixReName, tc.givenService())
			require.NoError(t, err)

			// then
			assert.Equal(t, tc.expectedService(), result)
			assert.True(t, len(tc.expectedService().Name) < 64)
		})
	}

}

func TestFailConvertServiceWhenAccessLabelNotProvided(t *testing.T) {
	// given
	converter := broker.NewConverter()

	// when
	_, err := converter.Convert("fix-re-name", internal.Service{
		APIEntry: &internal.APIEntry{},
	})

	// then
	assert.EqualError(t, err, "while creating the metadata object: cannot create binding labels: accessLabel field is required to build bindingLabels")

}

func TestIsBindableFalseForEventsBasedService(t *testing.T) {
	// given
	converter := broker.NewConverter()

	// when
	a, err := converter.Convert("fix-re-name", fixEventsBasedService())

	// then
	assert.NoError(t, err)
	assert.Equal(t, a.Bindable, false)
}

func TestIsBindableTrueForAPIBasedService(t *testing.T) {
	// given
	converter := broker.NewConverter()

	// when
	a, err := converter.Convert("fix-re-name", fixAPIBasedService())

	// then
	assert.NoError(t, err)
	assert.Equal(t, a.Bindable, true)
}

func fixAPIBasedService() internal.Service {
	return internal.Service{
		ID:                  internal.RemoteServiceID("0023-abcd-2098"),
		LongDescription:     "long description",
		Name:                "servicename",
		Description:         "short description",
		DisplayName:         "Service Name",
		ProviderDisplayName: "HakunaMatata",
		Tags:                []string{"tag1", "tag2"},
		Labels: map[string]string{
			"connected-app": "ec-prod",
		},
		APIEntry: &internal.APIEntry{
			AccessLabel: "access-label-1",
			GatewayURL:  "www.gate.com",
		},
	}
}

func fixEventsBasedService() internal.Service {
	return internal.Service{}
}

func fixOsbService() osb.Service {
	return osb.Service{
		ID:          "0023-abcd-2098",
		Name:        "servicename",
		Description: "short description",
		Bindable:    true,
		Plans: []osb.Plan{{
			Name:        "default",
			Description: "Default plan",
			ID:          fmt.Sprintf("%s-plan", "0023-abcd-2098"),
			Metadata: map[string]interface{}{
				"displayName": "Default",
			},
		}},
		Tags: []string{"tag1", "tag2"},
		Metadata: map[string]interface{}{
			"providerDisplayName":        "HakunaMatata",
			"displayName":                "service-name",
			"longDescription":            "long description",
			"remoteEnvironmentServiceId": "0023-abcd-2098",
			"bindingLabels": map[string]string{
				"access-label-1": "true",
			},
			"labels": map[string]string{
				"connected-app": "ec-prod",
			},
		},
	}
}

type catalogTestCase struct {
	finderMock           *automock.ReFinder
	converterMock        *automock.Converter
	reEnabledCheckerMock *automock.ReEnabledChecker
}

func newCatalogTC() *catalogTestCase {
	return &catalogTestCase{
		finderMock:           &automock.ReFinder{},
		converterMock:        &automock.Converter{},
		reEnabledCheckerMock: &automock.ReEnabledChecker{},
	}
}

func (tc *catalogTestCase) AssertExpectations(t *testing.T) {
	tc.finderMock.AssertExpectations(t)
	tc.converterMock.AssertExpectations(t)
}

func (tc *catalogTestCase) fixRE() *internal.RemoteEnvironment {
	return &internal.RemoteEnvironment{
		Name: "ec-prod",
		Services: []internal.Service{
			{
				ID: "00-1",
				APIEntry: &internal.APIEntry{
					GatewayURL:  "www.gate1.com",
					AccessLabel: "free",
				},
			},
		},
	}
}

func (tc *catalogTestCase) fixService() osb.Service {
	return osb.Service{ID: "bundleID"}
}
