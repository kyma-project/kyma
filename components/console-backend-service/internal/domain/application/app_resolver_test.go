package application_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/gateway"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApplicationStatusSuccess(t *testing.T) {
	for tn, tc := range map[string]struct {
		givenStatus    gateway.Status
		expectedStatus gqlschema.ApplicationStatus
	}{
		"serving": {
			givenStatus:    gateway.StatusServing,
			expectedStatus: gqlschema.ApplicationStatusServing,
		},
		"not serving": {
			givenStatus:    gateway.StatusNotServing,
			expectedStatus: gqlschema.ApplicationStatusNotServing,
		},
		"not configured": {
			givenStatus:    gateway.StatusNotConfigured,
			expectedStatus: gqlschema.ApplicationStatusGatewayNotConfigured,
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			statusGetterStub := automock.NewStatusGetter()
			statusGetterStub.On("GetStatus", "ec-prod").Return(tc.givenStatus, nil)
			resolver := application.NewApplicationResolver(nil, statusGetterStub)

			// when
			status, err := resolver.ApplicationStatusField(context.Background(), &gqlschema.Application{
				Name: "ec-prod",
			})

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, status)
		})
	}
}

func TestApplicationStatusFail(t *testing.T) {
	// given
	statusGetterStub := automock.NewStatusGetter()
	statusGetterStub.On("GetStatus", "ec-prod").Return(gateway.Status("fake"), nil)
	resolver := application.NewApplicationResolver(nil, statusGetterStub)

	// when
	_, err := resolver.ApplicationStatusField(context.Background(), &gqlschema.Application{
		Name: "ec-prod",
	})

	// then
	require.Error(t, err)
	assert.True(t, gqlerror.IsInternal(err))
}

func TestApplicationResolver_CreateApplication(t *testing.T) {
	// GIVEN
	for tn, tc := range map[string]struct {
		givenDesc      *string
		givenLabels    *gqlschema.Labels
		expectedDesc   string
		expectedLabels gqlschema.Labels
	}{
		"nothing provided": {},
		"fully parametrized": {
			givenDesc: ptrStr("desc"),
			givenLabels: &gqlschema.Labels{
				"lol": "test",
			},
			expectedDesc: "desc",
			expectedLabels: gqlschema.Labels{
				"lol": "test",
			},
		},
		"only desc provided": {
			givenDesc:    ptrStr("desc"),
			expectedDesc: "desc",
		},
		"only labels provided": {
			givenLabels: &gqlschema.Labels{
				"lol": "test",
			},
			expectedLabels: gqlschema.Labels{
				"lol": "test",
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			fixName := "fix-name"
			appSvc := automock.NewApplicationSvc()
			defer appSvc.AssertExpectations(t)
			appSvc.On("Create", fixName, tc.expectedDesc, tc.expectedLabels).Return(&v1alpha1.Application{
				ObjectMeta: v1.ObjectMeta{
					Name: fixName,
				},
				Spec: v1alpha1.ApplicationSpec{
					Description: tc.expectedDesc,
					Labels:      tc.expectedLabels,
				},
			}, nil)

			resolver := application.NewApplicationResolver(appSvc, nil)

			// WHEN
			out, err := resolver.CreateApplication(context.Background(), fixName, tc.givenDesc, tc.givenLabels)

			// THEN
			require.NoError(t, err)
			assert.Equal(t, fixName, out.Name)
			assert.Equal(t, tc.expectedDesc, out.Description)
			assert.Equal(t, tc.expectedLabels, out.Labels)
		})
	}
}

func TestApplicationResolver_CreateApplication_Error(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	appSvc := automock.NewApplicationSvc()
	appSvc.On("Create", fixName, "", gqlschema.Labels(nil)).Return(nil, errors.New("fix"))

	// WHEN
	resolver := application.NewApplicationResolver(appSvc, nil)
	_, err := resolver.CreateApplication(context.Background(), fixName, nil, nil)

	// THEN
	assert.EqualError(t, err, fmt.Sprintf("internal error [name: \"%s\"]", fixName))
}

func TestApplicationResolver_DeleteApplication(t *testing.T) {
	// GIVEN
	fixName := "fix"
	appSvc := automock.NewApplicationSvc()
	defer appSvc.AssertExpectations(t)
	appSvc.On("Delete", fixName).Return(nil)

	resolver := application.NewApplicationResolver(appSvc, nil)

	// WHEN
	out, err := resolver.DeleteApplication(context.Background(), fixName)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, fixName, out.Name)
}

func TestApplicationResolver_DeleteApplication_Error(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	appSvc := automock.NewApplicationSvc()
	defer appSvc.AssertExpectations(t)
	appSvc.On("Delete", fixName).Return(errors.New("fix"))

	resolver := application.NewApplicationResolver(appSvc, nil)

	// WHEN
	_, err := resolver.DeleteApplication(context.Background(), fixName)

	// THEN
	assert.EqualError(t, err, fmt.Sprintf("internal error [name: \"%s\"]", fixName))
}

func TestApplicationResolver_UpdateApplication(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	appSvc := automock.NewApplicationSvc()
	defer appSvc.AssertExpectations(t)

	resolver := application.NewApplicationResolver(appSvc, nil)

	for tn, tc := range map[string]struct {
		givenDesc      *string
		givenLabels    *gqlschema.Labels
		expectedDesc   string
		expectedLabels gqlschema.Labels
	}{
		"nothing provided": {},
		"fully parametrized": {
			givenDesc: ptrStr("desc"),
			givenLabels: &gqlschema.Labels{
				"lol": "test",
			},
			expectedDesc: "desc",
			expectedLabels: gqlschema.Labels{
				"lol": "test",
			},
		},
		"only desc provided": {
			givenDesc:    ptrStr("desc"),
			expectedDesc: "desc",
		},
		"only labels provided": {
			givenLabels: &gqlschema.Labels{
				"lol": "test",
			},
			expectedLabels: gqlschema.Labels{
				"lol": "test",
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			appSvc.On("Update", fixName, tc.expectedDesc, tc.expectedLabels).Return(&v1alpha1.Application{
				ObjectMeta: v1.ObjectMeta{
					Name: fixName,
				},
				Spec: v1alpha1.ApplicationSpec{
					Description: tc.expectedDesc,
					Labels:      tc.expectedLabels,
				},
			}, nil)

			// WHEN
			out, err := resolver.UpdateApplication(context.Background(), fixName, tc.givenDesc, tc.givenLabels)

			// THEN
			require.NoError(t, err)
			assert.Equal(t, fixName, out.Name)
			assert.Equal(t, tc.expectedDesc, out.Description)
			assert.Equal(t, tc.expectedLabels, out.Labels)
		})
	}
}

func TestApplicationResolver_UpdateApplication_Error(t *testing.T) {
	fixName := "fix-name"
	appSvc := automock.NewApplicationSvc()
	defer appSvc.AssertExpectations(t)
	appSvc.On("Update", fixName, "", gqlschema.Labels(nil)).Return(nil, errors.New("fix"))

	resolver := application.NewApplicationResolver(appSvc, nil)
	_, err := resolver.UpdateApplication(context.Background(), fixName, nil, nil)

	assert.EqualError(t, err, fmt.Sprintf("internal error [name: \"%s\"]", fixName))
}

func TestConnectorServiceQuerySuccess(t *testing.T) {
	// given
	var (
		fixAppName = "app-name"
		fixURL     = "http://some-url-with-token"
		fixGQLObj  = gqlschema.ConnectorService{
			URL: "http://some-url-with-token",
		}
	)

	serviceMock := automock.NewApplicationSvc()
	defer serviceMock.AssertExpectations(t)
	serviceMock.On("GetConnectionURL", fixAppName).Return(fixURL, nil)

	resolver := application.NewApplicationResolver(serviceMock, nil)

	// when
	gotURLObj, err := resolver.ConnectorServiceQuery(context.Background(), fixAppName)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLObj, gotURLObj)
}

func TestConnectorServiceQueryFail(t *testing.T) {
	// given
	var (
		fixAppName = ""
		fixErr     = errors.New("something went wrong")
	)

	serviceMock := automock.NewApplicationSvc()
	defer serviceMock.AssertExpectations(t)
	serviceMock.On("GetConnectionURL", fixAppName).Return("", fixErr)

	resolver := application.NewApplicationResolver(serviceMock, nil)

	// when
	gotURLObj, err := resolver.ConnectorServiceQuery(context.Background(), fixAppName)

	// then
	require.Error(t, err)
	assert.True(t, gqlerror.IsInternal(err))
	assert.Zero(t, gotURLObj)
}

func TestApplicationResolver_ApplicationEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		appSvc := automock.NewApplicationSvc()
		appSvc.On("Subscribe", mock.Anything).Once()
		appSvc.On("Unsubscribe", mock.Anything).Once()
		resolver := application.NewApplicationResolver(appSvc, nil)

		_, err := resolver.ApplicationEventSubscription(ctx)

		require.NoError(t, err)
		appSvc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		appSvc := automock.NewApplicationSvc()
		appSvc.On("Subscribe", mock.Anything).Once()
		appSvc.On("Unsubscribe", mock.Anything).Once()
		resolver := application.NewApplicationResolver(appSvc, nil)

		channel, err := resolver.ApplicationEventSubscription(ctx)
		<-channel

		require.NoError(t, err)
		appSvc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}

func TestApplicationResolver_EnableApplicationMutation(t *testing.T) {
	fixNamespace := "fix-namespace"
	fixName := "fix-name"
	trueVal := true
	falseVal := false

	for name, tc := range map[string]struct {
		allServices *bool
		services    []*gqlschema.ApplicationMappingService
		amExpected  []mappingTypes.ApplicationMappingService
		expected    []*gqlschema.ApplicationMappingService
	}{
		"nil Allservices, nil service list": {
			allServices: nil,
			services:    nil,
			amExpected:  nil,
			expected:    nil,
		},
		"true Allservices, nil service list": {
			allServices: &trueVal,
			services:    nil,
			amExpected:  nil,
			expected:    nil,
		},
		"false Allservices, nil service list": {
			allServices: &falseVal,
			services:    nil,
			amExpected:  []mappingTypes.ApplicationMappingService{},
			expected:    []*gqlschema.ApplicationMappingService{},
		},
		"nil Allservices, empty service list": {
			allServices: nil,
			services:    []*gqlschema.ApplicationMappingService{},
			amExpected:  nil,
			expected:    nil,
		},
		"true Allservices, empty service list": {
			allServices: &trueVal,
			services:    []*gqlschema.ApplicationMappingService{},
			amExpected:  nil,
			expected:    nil,
		},
		"false Allservices, empty service list": {
			allServices: &falseVal,
			services:    []*gqlschema.ApplicationMappingService{},
			amExpected:  []mappingTypes.ApplicationMappingService{},
			expected:    []*gqlschema.ApplicationMappingService{},
		},
		"nil Allservices, not empty service list": {
			allServices: nil,
			services: []*gqlschema.ApplicationMappingService{
				{
					ID: "30a09ece-ea06-42cd-ba6b-79f0c88675a0",
				},
			},
			amExpected: nil,
			expected:   nil,
		},
		"true Allservices, not empty service list": {
			allServices: &trueVal,
			services: []*gqlschema.ApplicationMappingService{
				{
					ID: "30a09ece-ea06-42cd-ba6b-79f0c88675a0",
				},
			},
			amExpected: nil,
			expected:   nil,
		},
		"false Allservices, not empty service list": {
			allServices: &falseVal,
			services: []*gqlschema.ApplicationMappingService{
				{
					ID: "30a09ece-ea06-42cd-ba6b-79f0c88675a0",
				},
			},
			amExpected: []mappingTypes.ApplicationMappingService{
				{
					ID: "30a09ece-ea06-42cd-ba6b-79f0c88675a0",
				},
			},
			expected: []*gqlschema.ApplicationMappingService{
				{
					ID: "30a09ece-ea06-42cd-ba6b-79f0c88675a0",
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			appSvc := automock.NewApplicationSvc()
			defer appSvc.AssertExpectations(t)
			appSvc.On("Enable", fixNamespace, fixName, tc.expected).Return(&mappingTypes.ApplicationMapping{
				ObjectMeta: v1.ObjectMeta{
					Name:      fixName,
					Namespace: fixNamespace,
				},
				Spec: mappingTypes.ApplicationMappingSpec{
					Services: tc.amExpected,
				},
			}, nil)
			resolver := application.NewApplicationResolver(appSvc, nil)

			// WHEN
			out, err := resolver.EnableApplicationMutation(context.Background(), fixName, fixNamespace, tc.allServices, tc.services)
			require.NoError(t, err)

			// THEN
			if tc.allServices == nil || *tc.allServices {
				assert.Nil(t, out.Services)
			} else {
				assert.Len(t, out.Services, len(tc.expected))
			}
		})
	}
}

func ptrStr(str string) *string {
	return &str
}
