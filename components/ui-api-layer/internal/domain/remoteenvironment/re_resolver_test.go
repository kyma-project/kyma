package remoteenvironment_test

import (
	"context"
	"testing"

	"time"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/gateway"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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
		givenLabels    *gqlschema.JSON
		expectedDesc   string
		expectedLabels gqlschema.JSON
	}{
		"nothing provided": {},
		"fully parametrized": {
			givenDesc: ptrStr("desc"),
			givenLabels: &gqlschema.JSON{
				"lol": "test",
			},
			expectedDesc: "desc",
			expectedLabels: gqlschema.JSON{
				"lol": "test",
			},
		},
		"only desc provided": {
			givenDesc:    ptrStr("desc"),
			expectedDesc: "desc",
		},
		"only labels provided": {
			givenLabels: &gqlschema.JSON{
				"lol": "test",
			},
			expectedLabels: gqlschema.JSON{
				"lol": "test",
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			fixName := "fix-name"
			client := fake.NewSimpleClientset()
			svc, err := remoteenvironment.NewRemoteEnvironmentService(client.ApplicationconnectorV1alpha1(), remoteenvironment.Config{}, newDummyInformer(), nil, nil)
			require.NoError(t, err)
			resolver := remoteenvironment.NewRemoteEnvironmentResolver(svc, nil)

			// WHEN
			out, err := resolver.CreateRemoteEnvironment(context.Background(), fixName, tc.givenDesc, tc.givenLabels)
			require.NoError(t, err)

			// THEN
			assert.Equal(t, fixName, out.Name)
			assert.Equal(t, tc.expectedDesc, out.Description)
			assert.Equal(t, tc.expectedLabels, out.Labels)
		})
	}
}

func TestRemoteEnvironmentResolver_DeleteRemoteEnvironment(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	client := fake.NewSimpleClientset(fixRemoteEnvironmentCR(fixName))
	svc, err := remoteenvironment.NewRemoteEnvironmentService(client.ApplicationconnectorV1alpha1(), remoteenvironment.Config{}, newDummyInformer(), nil, nil)
	require.NoError(t, err)

	resolver := remoteenvironment.NewRemoteEnvironmentResolver(svc, nil)

	// WHEN
	name, err := resolver.DeleteRemoteEnvironment(context.Background(), fixName)
	require.NoError(t, err)

	// THEN
	assert.Equal(t, gqlschema.DeleteRemoteEnvironmentOutput{Name: fixName}, name)

	_, err = client.ApplicationconnectorV1alpha1().RemoteEnvironments().Get(fixName, v1.GetOptions{})
	assert.Error(t, err)
}

func TestRemoteEnvironmentResolver_UpdateRemoteEnvironment(t *testing.T) {
	// GIVEN
	fixName := "fix-name"
	client := fake.NewSimpleClientset(fixRemoteEnvironmentCR(fixName))
	informerFactory := externalversions.NewSharedInformerFactory(client, time.Second)
	informer := informerFactory.Applicationconnector().V1alpha1().RemoteEnvironments().Informer()

	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	svc, err := remoteenvironment.NewRemoteEnvironmentService(client.ApplicationconnectorV1alpha1(), remoteenvironment.Config{}, newDummyInformer(), nil, informer)
	require.NoError(t, err)

	resolver := remoteenvironment.NewRemoteEnvironmentResolver(svc, nil)

	for tn, tc := range map[string]struct {
		givenDesc      *string
		givenLabels    *gqlschema.JSON
		expectedDesc   string
		expectedLabels gqlschema.JSON
	}{
		"nothing provided": {},
		"fully parametrized": {
			givenDesc: ptrStr("desc"),
			givenLabels: &gqlschema.JSON{
				"lol": "test",
			},
			expectedDesc: "desc",
			expectedLabels: gqlschema.JSON{
				"lol": "test",
			},
		},
		"only desc provided": {
			givenDesc:    ptrStr("desc"),
			expectedDesc: "desc",
		},
		"only labels provided": {
			givenLabels: &gqlschema.JSON{
				"lol": "test",
			},
			expectedLabels: gqlschema.JSON{
				"lol": "test",
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// WHEN
			out, err := resolver.UpdateRemoteEnvironment(context.Background(), fixName, tc.givenDesc, tc.givenLabels)
			require.NoError(t, err)

			// THEN
			assert.Equal(t, fixName, out.Name)
			assert.Equal(t, tc.expectedDesc, out.Description)
			assert.Equal(t, tc.expectedLabels, out.Labels)
		})
	}
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

func ptrStr(str string) *string {
	return &str
}
