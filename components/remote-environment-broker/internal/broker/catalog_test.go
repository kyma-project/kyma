package broker_test

import (
	"context"
	"fmt"
	"testing"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/remote-environment-broker/internal"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/broker"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/broker/automock"
)

func TestGetCatalogHappyPath(t *testing.T) {
	// GIVEN
	tc := newCatalogTC()
	defer tc.AssertExpectations(t)
	tc.finderMock.On("FindAll").Return([]*internal.RemoteEnvironment{tc.fixRE()}, nil).Once()
	tc.reEnabledCheckerMock.On("IsRemoteEnvironmentEnabled", "stage", string(tc.fixRE().Name)).Return(true, nil)
	tc.converterMock.On("Convert", tc.fixRE().Name, tc.fixRE().Source, tc.fixRE().Services[0]).Return(tc.fixService(), nil)

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
		givenSource  func() internal.Source
		givenService func() internal.Service

		expectedService func() osb.Service
	}{
		"simpleAPIBasedService": {
			givenSource: fixSource,
			givenService: func() internal.Service {
				svc := fixAPIBasedService()
				svc.DisplayName = "*Service Name\ną-'#$\tÜ"
				return svc
			},
			expectedService: func() osb.Service {
				svc := fixOsbService(fixReName)
				svc.Name = "service-name-c7fe3"
				svc.Metadata["displayName"] = "*Service Name\ną-'#$\tÜ"
				return svc
			},
		},
		"emptyDisplayName": {
			givenSource: fixSource,
			givenService: func() internal.Service {
				svc := fixAPIBasedService()
				svc.DisplayName = ""
				return svc
			},
			expectedService: func() osb.Service {
				svc := fixOsbService(fixReName)
				svc.Name = "c7fe3"
				svc.Metadata["displayName"] = ""
				return svc
			},
		},
		"longDisplayName": {
			givenSource: fixSource,
			givenService: func() internal.Service {
				svc := fixAPIBasedService()
				svc.DisplayName = "12345678901234567890123456789012345678901234567890123456789012345678901234567890"
				return svc
			},
			expectedService: func() osb.Service {
				svc := fixOsbService(fixReName)
				svc.Name = "123456789012345678901234567890123456789012345678901234567-c7fe3"
				svc.Metadata["displayName"] = "12345678901234567890123456789012345678901234567890123456789012345678901234567890"
				return svc
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			converter := broker.NewConverter()

			// when
			result, err := converter.Convert(fixReName, tc.givenSource(), tc.givenService())
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
	_, err := converter.Convert("fix-re-name", fixSource(), internal.Service{
		APIEntry: &internal.APIEntry{},
	})

	// then
	assert.EqualError(t, err, "while creating the metadata object: cannot create binding labels: accessLabel field is required to build bindingLabels")

}

func TestIsBindableFalseForEventsBasedService(t *testing.T) {
	// given
	converter := broker.NewConverter()

	// when
	a, err := converter.Convert("fix-re-name", fixSource(), fixEventsBasedService())

	// then
	assert.NoError(t, err)
	assert.Equal(t, a.Bindable, false)
}

func TestIsBindableTrueForAPIBasedService(t *testing.T) {
	// given
	converter := broker.NewConverter()

	// when
	a, err := converter.Convert("fix-re-name", fixSource(), fixAPIBasedService())

	// then
	assert.NoError(t, err)
	assert.Equal(t, a.Bindable, true)
}

func fixAPIBasedService() internal.Service {
	return internal.Service{
		ID:                  internal.RemoteServiceID("0023-abcd-2098"),
		LongDescription:     "long description",
		DisplayName:         "service name",
		ProviderDisplayName: "HakunaMatata",
		Tags:                []string{"tag1", "tag2"},
		APIEntry: &internal.APIEntry{
			AccessLabel: "access-label-1",
			GatewayURL:  "www.gate.com",
		},
	}
}

func fixEventsBasedService() internal.Service {
	return internal.Service{}
}

func fixSource() internal.Source {
	return internal.Source{
		Environment: "prod",
		Type:        "commerce",
		Namespace:   "com.hakuna.matata",
	}
}

func fixOsbService(reName string) osb.Service {
	return osb.Service{
		ID:          "0023-abcd-2098",
		Description: "long description",
		Bindable:    true,
		Name:        "serviceName",
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
			"providerDisplayName":        "HakunaMatata" + " - " + reName,
			"displayName":                "service-name",
			"longDescription":            "long description",
			"remoteEnvironmentServiceId": "0023-abcd-2098",
			"source": map[string]interface{}{
				"environment": "prod",
				"type":        "commerce",
				"namespace":   "com.hakuna.matata",
			},
			"bindingLabels": map[string]string{
				"access-label-1": "true",
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
