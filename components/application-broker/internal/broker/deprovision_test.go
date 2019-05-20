package broker

import (
	"context"
	"testing"

	"time"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker/automock"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger/spy"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	// GIVEN
	ts := newDeprovisionServiceTestSuite(t)
	defer ts.AssertExpectations(t)

	ts.OpIDProviderFake = func() (internal.OperationID, error) {
		return fixOperationID(), nil
	}
	ts.mockInstanceStorage.On("Get", fixInstanceID()).Return(fixNewInstance(), nil)
	ts.mockInstanceStorage.On("Remove", fixInstanceID()).Return(nil)
	ts.mockOperationStorage.On("Insert", fixNewRemoveInstanceOperation()).Return(nil)
	ts.mockOperationStorage.On("UpdateStateDesc", fixInstanceID(), fixOperationID(), internal.OperationStateSucceeded, fixDeprovisionSucceeded()).
		Return(nil)
	ts.mockInstanceStateGetter.On("IsDeprovisioned", fixInstanceID()).Return(false, nil).Once()
	ts.mockInstanceStateGetter.On("IsDeprovisioningInProgress", fixInstanceID()).Return(internal.OperationID(""), false, nil).Once()

	logSink := spy.NewLogSink()
	sut := NewDeprovisioner(ts.mockInstanceStorage, ts.mockInstanceStateGetter, ts.mockOperationStorage, ts.mockOperationStorage, ts.OpIDProviderFake, logSink.Logger)

	asyncFinished := make(chan struct{}, 0)
	sut.asyncHook = func() {
		asyncFinished <- struct{}{}
	}

	// WHEN
	actResp, err := sut.Deprovision(context.Background(), osbContext{}, fixDeprovisionRequest())

	// THEN
	assert.NoError(t, err)
	assert.NotNil(t, actResp)
	assert.True(t, actResp.Async)

	select {
	case <-asyncFinished:
	case <-time.After(time.Second):
		assert.Fail(t, "Async processing not finished")
	}
}

func TestErrorInstanceNotFound(t *testing.T) {
	// GIVEN
	ts := newDeprovisionServiceTestSuite(t)
	defer ts.AssertExpectations(t)

	ts.OpIDProviderFake = func() (internal.OperationID, error) {
		return fixOperationID(), nil
	}
	ts.mockInstanceStorage.On("Get", fixInstanceID()).Return(fixNewInstance(), nil)
	ts.mockInstanceStorage.On("Remove", fixInstanceID()).Return(mockNotFoundError{})
	ts.mockOperationStorage.On("Insert", fixNewRemoveInstanceOperation()).Return(nil)
	ts.mockInstanceStateGetter.On("IsDeprovisioned", fixInstanceID()).Return(false, nil).Once()
	ts.mockInstanceStateGetter.On("IsDeprovisioningInProgress", fixInstanceID()).Return(internal.OperationID(""), false, nil).Once()

	sut := NewDeprovisioner(ts.mockInstanceStorage, ts.mockInstanceStateGetter, ts.mockOperationStorage, nil, ts.OpIDProviderFake, spy.NewLogDummy())

	// WHEN
	_, err := sut.Deprovision(context.Background(), osbContext{}, fixDeprovisionRequest())

	// THEN
	assert.Error(t, err)
	assert.True(t, IsNotFoundError(err))

}

func TestErrorOnRemovingInstance(t *testing.T) {
	// GIVEN
	ts := newDeprovisionServiceTestSuite(t)
	defer ts.AssertExpectations(t)

	ts.OpIDProviderFake = func() (internal.OperationID, error) {
		return fixOperationID(), nil
	}
	ts.mockInstanceStorage.On("Get", fixInstanceID()).Return(fixNewInstance(), nil)
	ts.mockInstanceStorage.On("Remove", fixInstanceID()).Return(errors.New("simple error"))
	ts.mockOperationStorage.On("Insert", fixNewRemoveInstanceOperation()).Return(nil)
	ts.mockInstanceStateGetter.On("IsDeprovisioned", fixInstanceID()).Return(false, nil).Once()
	ts.mockInstanceStateGetter.On("IsDeprovisioningInProgress", fixInstanceID()).Return(internal.OperationID(""), false, nil).Once()

	sut := NewDeprovisioner(ts.mockInstanceStorage, ts.mockInstanceStateGetter, ts.mockOperationStorage, nil, ts.OpIDProviderFake, spy.NewLogDummy())

	// WHEN
	_, err := sut.Deprovision(context.Background(), osbContext{}, fixDeprovisionRequest())

	// THEN
	assert.Error(t, err)
	assert.False(t, IsNotFoundError(err))
}

func TestErrorOnIsDeprovisionedInstance(t *testing.T) {
	// GIVEN
	mockStateGetter := &automock.InstanceStateGetter{}
	defer mockStateGetter.AssertExpectations(t)

	mockStateGetter.On("IsDeprovisioned", fixInstanceID()).Return(false, fixError())

	sut := NewDeprovisioner(nil, mockStateGetter, nil, nil, nil, spy.NewLogDummy())
	// WHEN
	_, err := sut.Deprovision(context.Background(), osbContext{}, fixDeprovisionRequest())

	// THEN
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "while checking if instance is already deprovisioned")
}

func TestErrorOnDeprovisioningInProgressInstance(t *testing.T) {
	// GIVEN
	mockStateGetter := &automock.InstanceStateGetter{}
	defer mockStateGetter.AssertExpectations(t)

	mockStateGetter.On("IsDeprovisioned", fixInstanceID()).Return(false, nil)
	mockStateGetter.On("IsDeprovisioningInProgress", fixInstanceID()).Return(internal.OperationID(""), false, fixError())

	sut := NewDeprovisioner(nil, mockStateGetter, nil, nil, nil, spy.NewLogDummy())
	// WHEN
	_, err := sut.Deprovision(context.Background(), osbContext{}, fixDeprovisionRequest())

	// THEN
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "while checking if instance is being deprovisioned")
}

func newDeprovisionServiceTestSuite(t *testing.T) *deprovisionServiceTestSuite {
	return &deprovisionServiceTestSuite{
		t: t,
		mockInstanceStateGetter: &automock.InstanceStateGetter{},
		mockInstanceStorage:     &automock.InstanceStorage{},
		mockOperationStorage:    &automock.OperationStorage{},
	}
}

type deprovisionServiceTestSuite struct {
	t                       *testing.T
	mockInstanceStateGetter *automock.InstanceStateGetter
	mockInstanceStorage     *automock.InstanceStorage
	mockOperationStorage    *automock.OperationStorage
	OpIDProviderFake        func() (internal.OperationID, error)
}

func (ts *deprovisionServiceTestSuite) AssertExpectations(t *testing.T) {
	ts.mockInstanceStateGetter.AssertExpectations(t)
	ts.mockInstanceStorage.AssertExpectations(t)
	ts.mockOperationStorage.AssertExpectations(t)
}

func fixDeprovisionSucceeded() *string {
	s := "deprovision succeeded"
	return &s
}

type mockNotFoundError struct {
}

func (mockNotFoundError) Error() string {
	return "not found error"
}

func (mockNotFoundError) NotFound() bool {
	return true
}
