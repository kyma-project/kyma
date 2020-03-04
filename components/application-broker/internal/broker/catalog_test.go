package broker_test

import (
	"context"
	"fmt"
	"testing"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker/automock"
)

func TestGetCatalogHappyPath(t *testing.T) {
	// GIVEN
	tc := newCatalogTC()
	defer tc.AssertExpectations(t)

	app := fixAppWithOneService()
	tc.finderMock.On("FindAll").Return([]*internal.Application{&app}, nil).Once()
	tc.enableServices("stage", string(app.Name), app.Services[0].ID)
	tc.converterMock.On("Convert", tc.checker, app).Return([]osb.Service{fixAPIBasedOsbService()}, nil)

	svc := broker.NewCatalogService(tc.finderMock, tc.serviceCheckerFactory, tc.converterMock)
	osbCtx := broker.NewOSBContext("not", "important", "stage")

	// WHEN
	resp, err := svc.GetCatalog(context.Background(), *osbCtx)

	// THEN
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Services, 1)
	assert.Equal(t, fixAPIBasedOsbService(), resp.Services[0])
}

func TestConvertService(t *testing.T) {
	for tn, tc := range map[string]struct {
		givenApp        internal.Application
		expectedService []osb.Service
		checker         access.ServiceEnabledChecker
	}{
		"should return empty service list as none are enabled": {
			givenApp:        fixAppWithOneService(),
			expectedService: []osb.Service{},
			checker:         newServiceChecker("some-other-id"),
		},
		"should convert service which is enabled": {
			givenApp:        fixAppWithTwoServices(),
			expectedService: []osb.Service{fixAPIBasedOsbService()},
			checker:         newServiceChecker(fixAppWithTwoServices().Services[0].ID),
		},
		"should convert simple api service": {
			givenApp:        fixAppWithOneService(),
			expectedService: []osb.Service{fixAPIBasedOsbService()},
			checker:         newServiceChecker(fixAppWithOneService().Services[0].ID),
		},
		"should support special characters on DisplayName field": {
			givenApp: func() internal.Application {
				app := fixAppWithOneService()
				app.Services[0].DisplayName = "*Service Name\ną-'#$\tÜ"
				return app
			}(),
			expectedService: func() []osb.Service {
				svc := fixAPIBasedOsbService()
				svc.Metadata["displayName"] = "*Service Name\ną-'#$\tÜ"
				return []osb.Service{svc}
			}(),
			checker: newServiceChecker(fixAppWithOneService().Services[0].ID),
		},
		"should override provisionOnlyOnce label to true": {
			givenApp: func() internal.Application {
				app := fixAppWithOneService()
				app.Services[0].Labels["provisionOnlyOnce"] = "false"
				return app
			}(),
			expectedService: func() []osb.Service {
				svc := fixAPIBasedOsbService()
				l := svc.Metadata["labels"].(map[string]string)
				l["provisionOnlyOnce"] = "true"
				return []osb.Service{svc}
			}(),
			checker: newServiceChecker(fixAppWithOneService().Services[0].ID),
		},
		"should always add provisionOnlyOnce label set to true": {
			givenApp: func() internal.Application {
				app := fixAppWithOneService()
				delete(app.Services[0].Labels, "provisionOnlyOnce")
				return app
			}(),
			expectedService: func() []osb.Service {
				svc := fixAPIBasedOsbService()
				l := svc.Metadata["labels"].(map[string]string)
				l["provisionOnlyOnce"] = "true"
				return []osb.Service{svc}
			}(),
			checker: newServiceChecker(fixAppWithOneService().Services[0].ID),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			converter := broker.NewConverter()

			// when
			result, err := converter.Convert(tc.checker, tc.givenApp)
			require.NoError(t, err)

			// then
			require.Equal(t, tc.expectedService, result)
		})
	}
}

func TestFailConvertServiceWhenAccessLabelNotProvided(t *testing.T) {
	// given
	converter := broker.NewConverter()
	svc := fixAPIBasedService()
	svc.Entries[0].AccessLabel = ""

	// when
	_, err := converter.Convert(newServiceChecker(svc.ID), internal.Application{
		Services: []internal.Service{svc},
	})

	// then
	assert.EqualError(t, err, "while converting application to service: while creating the metadata object: cannot create binding labels: accessLabel field is required to build bindingLabels")

}

func TestIsBindableFalseForEventsBasedService(t *testing.T) {
	// given
	converter := broker.NewConverter()
	svc := fixEventsBasedService()

	// when
	services, err := converter.Convert(newServiceChecker(svc.ID), internal.Application{
		Services: []internal.Service{svc},
	})

	// then
	assert.NoError(t, err)
	assert.Len(t, services, 1)
	assert.False(t, services[0].Bindable)
}

func TestIsBindableTrueForAPIBasedService(t *testing.T) {
	// given
	converter := broker.NewConverter()
	svc := fixAPIBasedService()

	// when
	services, err := converter.Convert(newServiceChecker(svc.ID), internal.Application{
		Services: []internal.Service{svc},
	})

	// then
	assert.NoError(t, err)
	assert.Len(t, services, 1)
	assert.True(t, services[0].Bindable)

}

func TestConvertServiceV2(t *testing.T) {
	for tn, tc := range map[string]struct {
		givenApp        internal.Application
		expectedService []osb.Service
		checker         access.ServiceEnabledChecker
	}{
		"should convert only one service which is enabled": {
			givenApp:        fixAppWithTwoServices(),
			expectedService: []osb.Service{fixAPIBasedOsbServiceV2()},
			checker:         newServiceChecker(fixAppWithTwoServices().Services[0].ID),
		},
		"should convert two services which are enabled": {
			givenApp:        fixAppWithTwoServices(),
			expectedService: []osb.Service{fixAPIAndEventBasedOsbServiceV2()},
			checker:         newServiceChecker(fixAppWithTwoServices().Services[0].ID, fixAppWithTwoServices().Services[1].ID),
		},
		"should convert simple api service": {
			givenApp:        fixAppWithOneService(),
			expectedService: []osb.Service{fixAPIBasedOsbServiceV2()},
			checker:         newServiceChecker(fixAppWithOneService().Services[0].ID),
		},
		"should support special characters on DisplayName fields": {
			givenApp: func() internal.Application {
				app := fixAppWithOneService()
				app.DisplayName = "*Service Name\ną-'#$\tÜ"
				app.Services[0].DisplayName = "*Service Name\ną-'#$\tÜ"
				return app
			}(),
			expectedService: func() []osb.Service {
				svc := fixAPIBasedOsbServiceV2()
				svc.Metadata["displayName"] = "*Service Name\ną-'#$\tÜ"
				svc.Plans[0].Metadata["displayName"] = "*Service Name\ną-'#$\tÜ"
				return []osb.Service{svc}
			}(),
			checker: newServiceChecker(fixAppWithOneService().Services[0].ID),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			converter := broker.NewConverterV2()

			// when
			result, err := converter.Convert(tc.checker, tc.givenApp)
			require.NoError(t, err)

			// then
			require.EqualValues(t, tc.expectedService, result)
		})
	}
}

func TestFailConvertServiceV2WhenNoServicesAreMappedToPlans(t *testing.T) {
	// given
	var (
		converter = broker.NewConverterV2()
		givenApp  = fixAppWithOneService()
		checker   = newServiceChecker("some-other-id")
	)

	// when
	_, err := converter.Convert(checker, givenApp)

	// then
	assert.EqualError(t, err, "None plans were mapped from Application Services. Used Checker: Service Checker for testing purpose, Services: [[{ID:api-0023-abcd-2098 Name:api-service-name DisplayName:API Based Service display name Description:API Based Service description Entries:[APIEntry{Name: api-entry-name-for-V2, TargetURL: , GateywaURL:www.gate.com, AccessLabel: access-label-1}] EventProvider:false ServiceInstanceCreateParameterSchema:map[] LongDescription:API Based Service long description ProviderDisplayName:HakunaMatata Tags:[tag1 tag2] Labels:map[connected-app:ec-prod]}]].")

}

func TestConvertServiceV2IsBindableFalseForEventsBasedService(t *testing.T) {
	// given
	converter := broker.NewConverterV2()
	svc := fixEventsBasedService()

	// when
	services, err := converter.Convert(newServiceChecker(svc.ID), internal.Application{
		Services: []internal.Service{svc},
	})

	// then
	assert.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Len(t, services[0].Plans, 1)
	assert.NotNil(t, services[0].Plans[0].Bindable)
	assert.False(t, *services[0].Plans[0].Bindable)
}

func TestConvertServiceV2IsBindableTrueForAPIBasedService(t *testing.T) {
	// given
	converter := broker.NewConverterV2()
	api := fixAPIBasedService()
	event := fixEventsBasedService()

	// when
	services, err := converter.Convert(newServiceChecker(api.ID, event.ID), internal.Application{
		Services: []internal.Service{
			api, event,
		},
	})

	// then
	assert.NoError(t, err)
	assert.Len(t, services, 1)
	assert.True(t, services[0].Bindable)

	require.Len(t, services[0].Plans, 2)

	gotAPIPlan, gotEventPlan := extractAPIAndEventPlan(services[0], api.ID, event.ID)
	assert.True(t, *gotAPIPlan.Bindable)
	assert.False(t, *gotEventPlan.Bindable)
}

func extractAPIAndEventPlan(service osb.Service, apiID, eventID internal.ApplicationServiceID) (osb.Plan, osb.Plan) {
	var gotAPIPlan osb.Plan
	var gotEventPlan osb.Plan
	for _, p := range service.Plans {
		switch internal.ApplicationServiceID(p.ID) {
		case apiID:
			gotAPIPlan = p
		case eventID:
			gotEventPlan = p
		}
	}

	return gotAPIPlan, gotEventPlan
}

func fixAppWithTwoServices() internal.Application {
	app := fixAppWithOneService()
	app.Services = append(app.Services, fixEventsBasedService())
	return app
}

func fixAppWithOneService() internal.Application {
	return internal.Application{
		Name:                "ec-prod",
		Description:         "DescriptionV2",
		DisplayName:         "DisplayNameV2",
		ProviderDisplayName: "ProviderDisplayNameV2",
		LongDescription:     "LongDescriptionV2",
		Tags:                []string{"tag1-V2", "tag2-V2"},
		Labels: map[string]string{
			"LabelKeyV2": "LabelValueV2",
		},
		CompassMetadata: internal.CompassMetadata{
			ApplicationID: "app-id-consumed-by-V2",
		},
		Services: []internal.Service{
			fixAPIBasedService(),
		},
	}
}

func fixAPIBasedService() internal.Service {
	return internal.Service{
		ID:                  "api-0023-abcd-2098",
		Name:                "api-service-name",
		Description:         "API Based Service description",
		DisplayName:         "API Based Service display name",
		LongDescription:     "API Based Service long description",
		ProviderDisplayName: "HakunaMatata",
		Tags:                []string{"tag1", "tag2"},
		Labels: map[string]string{
			"connected-app": "ec-prod",
		},
		Entries: []internal.Entry{{
			Type: "API",
			APIEntry: &internal.APIEntry{
				Name:        "api-entry-name-for-V2",
				GatewayURL:  "www.gate.com",
				AccessLabel: "access-label-1",
			},
		}},
	}
}

func fixEventsBasedService() internal.Service {
	return internal.Service{
		ID:                  "events-0023-abcd-2098",
		DisplayName:         "Events Based Service display name",
		Description:         "Events Based Service short description",
		Name:                "events-service-name",
		EventProvider:       true,
		Entries:             nil,
		LongDescription:     "Events Based Service long description",
		ProviderDisplayName: "HakunaMatata",
		Tags:                []string{"tag1", "tag2"},
		Labels: map[string]string{
			"connected-app": "ec-prod",
		},
	}
}

func fixAPIBasedOsbService() osb.Service {
	return osb.Service{
		ID:          "api-0023-abcd-2098",
		Name:        "api-service-name",
		Description: "API Based Service description",
		Bindable:    true,
		Plans: []osb.Plan{{
			Name:        "default",
			Description: "Default plan",
			ID:          fmt.Sprintf("%s-plan", "api-0023-abcd-2098"),
			Metadata: map[string]interface{}{
				"displayName": "Default",
			},
		}},
		Tags: []string{"tag1", "tag2"},
		Metadata: map[string]interface{}{
			"providerDisplayName":  "HakunaMatata",
			"displayName":          "API Based Service display name",
			"longDescription":      "API Based Service long description",
			"applicationServiceId": "api-0023-abcd-2098",
			"bindingLabels": map[string]string{
				"access-label-1": "true",
			},
			"labels": map[string]string{
				"connected-app":     "ec-prod",
				"provisionOnlyOnce": "true",
			},
		},
	}
}

func fixAPIBasedOsbServiceV2() osb.Service {
	bindableTrue := true
	return osb.Service{
		Name:                "ec-prod",
		Description:         "DescriptionV2",
		ID:                  "app-id-consumed-by-V2",
		Tags:                []string{"tag1-V2", "tag2-V2"},
		Bindable:            bindableTrue,
		BindingsRetrievable: false,
		Plans: []osb.Plan{
			{
				ID:          "api-0023-abcd-2098",
				Name:        "api-service-name",
				Description: "API Based Service description",
				Bindable:    &bindableTrue,
				Metadata:    map[string]interface{}{"displayName": "API Based Service display name"},
			},
		},
		Metadata: map[string]interface{}{
			"displayName": "DisplayNameV2",
			"labels": map[string]string{
				"LabelKeyV2":             "LabelValueV2",
				"documentation-per-plan": "true",
			},
			"longDescription":     "LongDescriptionV2",
			"providerDisplayName": "ProviderDisplayNameV2",
		}}
}

func fixAPIAndEventBasedOsbServiceV2() osb.Service {
	bindableTrue := true
	bindableFalse := false
	return osb.Service{
		Name:                "ec-prod",
		Description:         "DescriptionV2",
		ID:                  "app-id-consumed-by-V2",
		Tags:                []string{"tag1-V2", "tag2-V2"},
		Bindable:            bindableTrue,
		BindingsRetrievable: false,
		Plans: []osb.Plan{
			{
				ID:          "api-0023-abcd-2098",
				Name:        "api-service-name",
				Description: "API Based Service description",
				Bindable:    &bindableTrue,
				Metadata:    map[string]interface{}{"displayName": "API Based Service display name"},
			},
			{
				ID:          "events-0023-abcd-2098",
				Name:        "events-service-name",
				Description: "Events Based Service short description",
				Bindable:    &bindableFalse,
				Metadata:    map[string]interface{}{"displayName": "Events Based Service display name"},
			},
		},
		Metadata: map[string]interface{}{
			"displayName": "DisplayNameV2",
			"labels": map[string]string{
				"LabelKeyV2":             "LabelValueV2",
				"documentation-per-plan": "true",
			},
			"longDescription":     "LongDescriptionV2",
			"providerDisplayName": "ProviderDisplayNameV2",
		}}
}

type catalogTestCase struct {
	finderMock            *automock.AppFinder
	converterMock         *automock.Converter
	serviceCheckerFactory *automock.ServiceCheckerFactory
	checker               access.ServiceEnabledChecker
}

func newCatalogTC() *catalogTestCase {
	return &catalogTestCase{
		finderMock:            &automock.AppFinder{},
		converterMock:         &automock.Converter{},
		serviceCheckerFactory: &automock.ServiceCheckerFactory{},
	}
}

func (tc *catalogTestCase) AssertExpectations(t *testing.T) {
	tc.finderMock.AssertExpectations(t)
	tc.converterMock.AssertExpectations(t)
}

func (tc *catalogTestCase) enableServices(namespace string, name string, serviceIDs ...internal.ApplicationServiceID) {
	tc.checker = newServiceChecker(serviceIDs...)
	tc.serviceCheckerFactory.On("NewServiceChecker", namespace, name).Return(tc.checker, nil)
}

type serviceChecker struct {
	serviceIDs map[internal.ApplicationServiceID]struct{}
}

func (c *serviceChecker) IdentifyYourself() string {
	return "Service Checker for testing purpose"
}

func (c *serviceChecker) IsServiceEnabled(svc internal.Service) bool {
	_, exists := c.serviceIDs[svc.ID]
	return exists
}

func newServiceChecker(ids ...internal.ApplicationServiceID) access.ServiceEnabledChecker {
	serviceIDs := make(map[internal.ApplicationServiceID]struct{})
	for _, id := range ids {
		serviceIDs[id] = struct{}{}
	}
	return &serviceChecker{serviceIDs: serviceIDs}
}
