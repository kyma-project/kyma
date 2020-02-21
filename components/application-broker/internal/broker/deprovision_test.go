package broker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker/automock"
	bt "github.com/kyma-project/kyma/components/application-broker/internal/broker/testing"
	"github.com/kyma-project/kyma/components/application-broker/internal/knative"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger/spy"
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
	ts.mockAppFinder.On("FindOneByServiceID", fixAppServiceID()).
		Return(fixApp(), nil).
		Once()

	logSink := spy.NewLogSink()
	sut := NewDeprovisioner(
		ts.mockInstanceStorage,
		ts.mockInstanceStateGetter,
		ts.mockOperationStorage,
		ts.mockOperationStorage,
		ts.OpIDProviderFake,
		ts.mockAppFinder,
		ts.client,
		logSink.Logger,
		false,
	)

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
	ts.mockAppFinder.On("FindOneByServiceID", fixAppServiceID()).
		Return(fixApp(), nil).
		Once()

	sut := NewDeprovisioner(
		ts.mockInstanceStorage,
		ts.mockInstanceStateGetter,
		ts.mockOperationStorage,
		nil,
		ts.OpIDProviderFake,
		ts.mockAppFinder,
		ts.client,
		spy.NewLogDummy(),
		false,
	)

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
	ts.mockAppFinder.On("FindOneByServiceID", fixAppServiceID()).
		Return(fixApp(), nil).
		Once()

	sut := NewDeprovisioner(
		ts.mockInstanceStorage,
		ts.mockInstanceStateGetter,
		ts.mockOperationStorage,
		nil,
		ts.OpIDProviderFake,
		ts.mockAppFinder,
		ts.client,
		spy.NewLogDummy(),
		false,
	)

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

	sut := NewDeprovisioner(
		nil,
		mockStateGetter,
		nil,
		nil,
		nil,
		nil,
		nil,
		spy.NewLogDummy(),
		false,
	)

	// WHEN
	_, err := sut.Deprovision(context.Background(), osbContext{}, fixDeprovisionRequest())

	// THEN
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checking if instance is already deprovisioned")
}

func TestErrorOnDeprovisioningInProgressInstance(t *testing.T) {
	// GIVEN
	mockStateGetter := &automock.InstanceStateGetter{}
	defer mockStateGetter.AssertExpectations(t)

	mockStateGetter.On("IsDeprovisioned", fixInstanceID()).Return(false, nil)
	mockStateGetter.On("IsDeprovisioningInProgress", fixInstanceID()).Return(internal.OperationID(""), false, fixError())

	sut := NewDeprovisioner(
		nil,
		mockStateGetter,
		nil,
		nil,
		nil,
		nil,
		nil,
		spy.NewLogDummy(),
		false,
	)

	// WHEN
	_, err := sut.Deprovision(context.Background(), osbContext{}, fixDeprovisionRequest())

	// THEN
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checking if instance is being deprovisioned")
}

func TestDoDeprovision(t *testing.T) {
	var (
		iID     = fixInstanceID()
		opID    = fixOperationID()
		appNs   = fixNs()
		appName = fixAppName()
	)

	testCases := map[string]struct {
		initialObjs   []runtime.Object
		expectUpdates []runtime.Object
		expectDeletes []string
	}{
		"Nothing to deprovision": {
			initialObjs:   nil,
			expectUpdates: nil,
			expectDeletes: nil,
		},
		// TODO(nachtmaar): Deprovision broker: https://github.com/kyma-project/kyma/issues/6342
		"Everything gets deprovisioned": {
			initialObjs: []runtime.Object{
				bt.NewAppNamespace(string(appNs), true),
				bt.NewDefaultBroker(string(appNs)),
				bt.NewAppSubscription(string(appNs), string(appName), bt.WithNameSuffix(bt.FakeSubscriptionName)),
			},
			expectDeletes: []string{
				fmt.Sprintf("%s/%s-%s", integrationNamespace, knSubscriptionNamePrefix, bt.FakeSubscriptionName), // subscription
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mockOpUpdater := &automock.OperationStorage{}
			mockOpUpdater.On("UpdateStateDesc", iID, opID,
				internal.OperationStateSucceeded,
				fixDeprovisionSucceeded(),
			).Return(nil).Once()

			/* The third return value is an istio client, which is ignored in case of deprovisioning because it's used
			only to create istio policy in case of provisioning in order to enable Prometheus scraping which is required
			even in case of deprovisioning
			*/
			knCli, k8sCli, _ := bt.NewFakeClients(tc.initialObjs...)

			dpr := NewDeprovisioner(
				nil,
				nil,
				nil,
				mockOpUpdater,
				nil,
				nil,
				knative.NewClient(knCli, k8sCli),
				spy.NewLogDummy(),
				false,
			)

			// WHEN
			dpr.do(iID, opID, appName, appNs)

			//THEN
			actionsAsserter := bt.NewActionsAsserter(t, knCli, k8sCli)
			actionsAsserter.AssertUpdates(t, tc.expectUpdates)
			actionsAsserter.AssertDeletes(t, tc.expectDeletes)
			mockOpUpdater.AssertExpectations(t)
		})
	}
}

func newDeprovisionServiceTestSuite(t *testing.T) *deprovisionServiceTestSuite {
	knCli, k8sCli, _ := bt.NewFakeClients()
	return &deprovisionServiceTestSuite{
		t:                       t,
		mockInstanceStateGetter: &automock.InstanceStateGetter{},
		mockInstanceStorage:     &automock.InstanceStorage{},
		mockOperationStorage:    &automock.OperationStorage{},
		mockAppFinder:           &automock.AppFinder{},
		client:                  knative.NewClient(knCli, k8sCli),
	}
}

type deprovisionServiceTestSuite struct {
	t                       *testing.T
	mockInstanceStateGetter *automock.InstanceStateGetter
	mockInstanceStorage     *automock.InstanceStorage
	mockOperationStorage    *automock.OperationStorage
	OpIDProviderFake        func() (internal.OperationID, error)
	mockAppFinder           *automock.AppFinder
	client                  knative.Client
}

func (ts *deprovisionServiceTestSuite) AssertExpectations(t *testing.T) {
	ts.mockInstanceStateGetter.AssertExpectations(t)
	ts.mockInstanceStorage.AssertExpectations(t)
	ts.mockOperationStorage.AssertExpectations(t)
}

type mockNotFoundError struct {
}

func (mockNotFoundError) Error() string {
	return "not found error"
}

func (mockNotFoundError) NotFound() bool {
	return true
}
