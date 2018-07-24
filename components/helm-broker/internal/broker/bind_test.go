package broker_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker/automock"
)

func TestBindServiceBindSuccess(t *testing.T) {
	// given
	tc := newBindTC()
	defer tc.AssertExpectations(t)
	fixID := tc.FixBindRequest().InstanceID
	expCreds := map[string]string{
		"password": "secret",
	}
	tc.ExpectOnGet(fixID, expCreds)

	svc := broker.NewBindService(tc.InstanceBindDataGetter)
	osbCtx := broker.NewOSBContext("not", "important")

	// when
	resp, err := svc.Bind(context.Background(), *osbCtx, tc.FixBindRequest())

	// then
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"password": "secret",
	}, resp.Credentials)
	assert.Nil(t, resp.RouteServiceURL)
	assert.Nil(t, resp.SyslogDrainURL)
	assert.Nil(t, resp.VolumeMounts)
}

func TestBindServiceBindFailure(t *testing.T) {
	t.Run("On service Get", func(t *testing.T) {
		// given
		tc := newBindTC()
		defer tc.AssertExpectations(t)
		fixID := tc.FixBindRequest().InstanceID
		fixErr := errors.New("Get ERR")
		tc.ExpectOnGetError(fixID, fixErr)

		svc := broker.NewBindService(tc.InstanceBindDataGetter)
		osbCtx := broker.NewOSBContext("not", "important")

		// when
		resp, err := svc.Bind(context.Background(), *osbCtx, tc.FixBindRequest())

		// then
		require.EqualError(t, err, fmt.Sprintf("while getting bind data from storage for instance id: %q: %v", fixID, fixErr.Error()))
		assert.Nil(t, resp)
	})

	t.Run("On unexpected req params", func(t *testing.T) {
		// given
		tc := newBindTC()
		fixReq := tc.FixBindRequest()
		fixReq.Parameters = map[string]interface{}{
			"some-key": "some-value",
		}

		svc := broker.NewBindService(nil)
		osbCtx := broker.NewOSBContext("not", "important")

		// when
		resp, err := svc.Bind(context.Background(), *osbCtx, fixReq)

		// then
		assert.EqualError(t, err, "helm-broker does not support configuration options for the service binding")
		assert.Zero(t, resp)
	})
}

func newBindTC() *bindServiceTestCase {
	return &bindServiceTestCase{
		InstanceBindDataGetter: &automock.InstanceBindDataGetter{},
	}
}

type bindServiceTestCase struct {
	InstanceBindDataGetter *automock.InstanceBindDataGetter
}

func (tc bindServiceTestCase) AssertExpectations(t *testing.T) {
	tc.InstanceBindDataGetter.AssertExpectations(t)
}

func (tc *bindServiceTestCase) ExpectOnGet(iID string, creds map[string]string) {
	tc.InstanceBindDataGetter.On("Get", internal.InstanceID(iID)).
		Return(&internal.InstanceBindData{
			InstanceID:  internal.InstanceID(iID),
			Credentials: internal.InstanceCredentials(creds),
		}, nil).Once()
}

func (tc *bindServiceTestCase) ExpectOnGetError(iID string, err error) {
	tc.InstanceBindDataGetter.On("Get", internal.InstanceID(iID)).
		Return(nil, err).Once()
}

func (tc *bindServiceTestCase) FixBindRequest() *osb.BindRequest {
	return &osb.BindRequest{
		InstanceID: "instance-id",
		ServiceID:  "service-id",
		PlanID:     "plan-id",
	}
}
