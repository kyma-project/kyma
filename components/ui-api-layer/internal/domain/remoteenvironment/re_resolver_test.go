package remoteenvironment_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/gateway"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRemoteEnvironmentStatusSuccess(t *testing.T) {
	for tn, tc := range map[string]struct {
		givenStatus    gateway.Status
		expectedStatus gqlschema.RemoteEnvironmentStatus
	}{
		"serving": {
			givenStatus:    gateway.StatusServing,
			expectedStatus: gqlschema.RemoteEnvironmentStatusServing,
		},
		"not serving": {
			givenStatus:    gateway.StatusNotServing,
			expectedStatus: gqlschema.RemoteEnvironmentStatusNotServing,
		},
		"not configured": {
			givenStatus:    gateway.StatusNotConfigured,
			expectedStatus: gqlschema.RemoteEnvironmentStatusGatewayNotConfigured,
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			statusGetterStub := automock.NewStatusGetter()
			statusGetterStub.On("GetStatus", "ec-prod").Return(tc.givenStatus, nil)
			resolver := remoteenvironment.NewRemoteEnvironmentResolver(nil, statusGetterStub)

			// when
			status, err := resolver.RemoteEnvironmentStatusField(context.Background(), &gqlschema.RemoteEnvironment{
				Name: "ec-prod",
			})

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, status)
		})
	}
}

func TestRemoteEnvironmentStatusFail(t *testing.T) {
	// given
	statusGetterStub := automock.NewStatusGetter()
	statusGetterStub.On("GetStatus", "ec-prod").Return(gateway.Status("fake"), nil)
	resolver := remoteenvironment.NewRemoteEnvironmentResolver(nil, statusGetterStub)

	// when
	_, err := resolver.RemoteEnvironmentStatusField(context.Background(), &gqlschema.RemoteEnvironment{
		Name: "ec-prod",
	})

	// then
	require.Error(t, err)
	assert.True(t, gqlerror.IsInternal(err))
}

func TestRemoteEnvironmentResolver_CreateRemoteEnvironment(t *testing.T) {
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
			reSvc := automock.NewReSvc()
			defer reSvc.AssertExpectations(t)
			reSvc.On("Create", fixName, tc.expectedDesc, tc.expectedLabels).Return(&v1alpha1.RemoteEnvironment{
				ObjectMeta: v1.ObjectMeta{
					Name: fixName,
				},
				Spec: v1alpha1.RemoteEnvironmentSpec{
					Description: tc.expectedDesc,
					Labels:      tc.expectedLabels,
				},
			}, nil)

			resolver := remoteenvironment.NewRemoteEnvironmentResolver(reSvc, nil)

			// WHEN
			out, err := resolver.CreateRemoteEnvironment(context.Background(), fixName, tc.givenDesc, tc.givenLabels)

			// THEN
			require.NoError(t, err)
			assert.Equal(t, fixName, out.Name)
			assert.Equal(t, tc.expectedDesc, out.Description)
			assert.Equal(t, tc.expectedLabels, out.Labels)
		})
	}
}

func TestRemoteEnvironmentResolver_CreateRemoteEnvironment_Error(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	reSvc := automock.NewReSvc()
	reSvc.On("Create", fixName, "", gqlschema.Labels(nil)).Return(nil, errors.New("fix"))

	// WHEN
	resolver := remoteenvironment.NewRemoteEnvironmentResolver(reSvc, nil)
	_, err := resolver.CreateRemoteEnvironment(context.Background(), fixName, nil, nil)

	// THEN
	assert.EqualError(t, err, fmt.Sprintf("internal error [name: \"%s\"]", fixName))
}

func TestRemoteEnvironmentResolver_DeleteRemoteEnvironment(t *testing.T) {
	// GIVEN
	fixName := "fix"
	reSvc := automock.NewReSvc()
	defer reSvc.AssertExpectations(t)
	reSvc.On("Delete", fixName).Return(nil)

	resolver := remoteenvironment.NewRemoteEnvironmentResolver(reSvc, nil)

	// WHEN
	out, err := resolver.DeleteRemoteEnvironment(context.Background(), fixName)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, fixName, out.Name)
}

func TestRemoteEnvironmentResolver_DeleteRemoteEnvironment_Error(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	reSvc := automock.NewReSvc()
	defer reSvc.AssertExpectations(t)
	reSvc.On("Delete", fixName).Return(errors.New("fix"))

	resolver := remoteenvironment.NewRemoteEnvironmentResolver(reSvc, nil)

	// WHEN
	_, err := resolver.DeleteRemoteEnvironment(context.Background(), fixName)

	// THEN
	assert.EqualError(t, err, fmt.Sprintf("internal error [name: \"%s\"]", fixName))
}

func TestRemoteEnvironmentResolver_UpdateRemoteEnvironment(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	reSvc := automock.NewReSvc()
	defer reSvc.AssertExpectations(t)

	resolver := remoteenvironment.NewRemoteEnvironmentResolver(reSvc, nil)

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
			reSvc.On("Update", fixName, tc.expectedDesc, tc.expectedLabels).Return(&v1alpha1.RemoteEnvironment{
				ObjectMeta: v1.ObjectMeta{
					Name: fixName,
				},
				Spec: v1alpha1.RemoteEnvironmentSpec{
					Description: tc.expectedDesc,
					Labels:      tc.expectedLabels,
				},
			}, nil)

			// WHEN
			out, err := resolver.UpdateRemoteEnvironment(context.Background(), fixName, tc.givenDesc, tc.givenLabels)

			// THEN
			require.NoError(t, err)
			assert.Equal(t, fixName, out.Name)
			assert.Equal(t, tc.expectedDesc, out.Description)
			assert.Equal(t, tc.expectedLabels, out.Labels)
		})
	}
}

func TestRemoteEnvironmentResolver_UpdateRemoteEnvironment_Error(t *testing.T) {
	fixName := "fix-name"
	reSvc := automock.NewReSvc()
	defer reSvc.AssertExpectations(t)
	reSvc.On("Update", fixName, "", gqlschema.Labels(nil)).Return(nil, errors.New("fix"))

	resolver := remoteenvironment.NewRemoteEnvironmentResolver(reSvc, nil)
	_, err := resolver.UpdateRemoteEnvironment(context.Background(), fixName, nil, nil)

	assert.EqualError(t, err, fmt.Sprintf("internal error [name: \"%s\"]", fixName))
}

func TestConnectorServiceQuerySuccess(t *testing.T) {
	// given
	var (
		fixRemoteEnvName = "env-name"
		fixURL           = "http://some-url-with-token"
		fixGQLObj        = gqlschema.ConnectorService{
			URL: "http://some-url-with-token",
		}
	)

	serviceMock := automock.NewReSvc()
	defer serviceMock.AssertExpectations(t)
	serviceMock.On("GetConnectionURL", fixRemoteEnvName).Return(fixURL, nil)

	resolver := remoteenvironment.NewRemoteEnvironmentResolver(serviceMock, nil)

	// when
	gotURLObj, err := resolver.ConnectorServiceQuery(context.Background(), fixRemoteEnvName)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixGQLObj, gotURLObj)
}

func TestConnectorServiceQueryFail(t *testing.T) {
	// given
	var (
		fixRemoteEnvName = ""
		fixErr           = errors.New("something went wrong")
	)

	serviceMock := automock.NewReSvc()
	defer serviceMock.AssertExpectations(t)
	serviceMock.On("GetConnectionURL", fixRemoteEnvName).Return("", fixErr)

	resolver := remoteenvironment.NewRemoteEnvironmentResolver(serviceMock, nil)

	// when
	gotURLObj, err := resolver.ConnectorServiceQuery(context.Background(), fixRemoteEnvName)

	// then
	require.Error(t, err)
	assert.True(t, gqlerror.IsInternal(err))
	assert.Zero(t, gotURLObj)
}
func TestRemoteEnvironmentResolver_RemoteEnvironmentEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		reSvc := automock.NewReSvc()
		reSvc.On("Subscribe", mock.Anything).Once()
		reSvc.On("Unsubscribe", mock.Anything).Once()
		resolver := remoteenvironment.NewRemoteEnvironmentResolver(reSvc, nil)

		_, err := resolver.RemoteEnvironmentEventSubscription(ctx)

		require.NoError(t, err)
		reSvc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		reSvc := automock.NewReSvc()
		reSvc.On("Subscribe", mock.Anything).Once()
		reSvc.On("Unsubscribe", mock.Anything).Once()
		resolver := remoteenvironment.NewRemoteEnvironmentResolver(reSvc, nil)

		channel, err := resolver.RemoteEnvironmentEventSubscription(ctx)
		<-channel

		require.NoError(t, err)
		reSvc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}

func ptrStr(str string) *string {
	return &str
}
