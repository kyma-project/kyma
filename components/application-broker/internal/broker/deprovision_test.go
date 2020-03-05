package broker

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker/automock"
	bt "github.com/kyma-project/kyma/components/application-broker/internal/broker/testing"
	"github.com/kyma-project/kyma/components/application-broker/internal/knative"
	eaFake "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger/spy"
)

func TestDeprovisionSkippingDeletingResourcesSuccess(t *testing.T) {
	// GIVEN
	ts := newDeprovisionServiceTestSuite(t)
	defer ts.AssertExpectations(t)

	// simulate that we should skip deleting resources
	// because we are not the last instance fo the given plan and service ID
	instanceWithSamePlanAndServiceID := fixNewInstance()
	instanceWithSamePlanAndServiceID.ID = "123-123"
	findAllResponse := []*internal.Instance{instanceWithSamePlanAndServiceID}

	ts.mockInstanceStateGetter.On("IsDeprovisioned", fixInstanceID()).
		Return(false, nil).Once()
	ts.mockInstanceStateGetter.On("IsDeprovisioningInProgress", fixInstanceID()).
		Return(internal.OperationID(""), false, nil).Once()
	ts.mockInstanceStorage.On("Get", fixInstanceID()).
		Return(fixNewInstance(), nil).Once()
	ts.mockInstanceStorage.On("FindAll", mock.Anything).
		Run(assertMatcherReturnTrueOnInstance(t, instanceWithSamePlanAndServiceID)).
		Return(findAllResponse, nil).Once()
	ts.mockInstanceStorage.On("Remove", fixInstanceID()).
		Return(nil).Once()

	sut := NewDeprovisioner(
		ts.mockInstanceStorage,
		ts.mockInstanceStateGetter,
		ts.mockOperationStorage,
		ts.mockOperationStorage,
		nil,
		nil,
		nil,
		nil,
		spy.NewLogDummy(),
		&IDSelector{false},
	)

	// WHEN
	actResp, err := sut.Deprovision(context.Background(), osbContext{}, fixDeprovisionRequest())

	// THEN
	assert.NoError(t, err)
	assert.NotNil(t, actResp)
	assert.False(t, actResp.Async)
}

// TestDeprovisionDeletingResourcesSuccess test that the resource cleanup is triggered
// but in simple scenario where they are not found on cluster. More complex scenarios
// are tested directly on 'do' method in TestDoDeprovision test.
func TestDeprovisionDeletingResourcesSuccess(t *testing.T) {
	// GIVEN
	ts := newDeprovisionServiceTestSuite(t)
	defer ts.AssertExpectations(t)

	ts.OpIDProviderFake = func() (internal.OperationID, error) {
		return fixOperationID(), nil
	}
	ts.mockInstanceStateGetter.On("IsDeprovisioned", fixInstanceID()).
		Return(false, nil).Once()
	ts.mockInstanceStateGetter.On("IsDeprovisioningInProgress", fixInstanceID()).
		Return(internal.OperationID(""), false, nil).Once()
	ts.mockInstanceStorage.On("Get", fixInstanceID()).
		Return(fixNewInstance(), nil).Once()
	ts.mockInstanceStorage.On("FindAll", mock.Anything).
		Return([]*internal.Instance{}, nil).Once()
	ts.mockInstanceStorage.On("Remove", fixInstanceID()).
		Return(nil).Once()
	ts.mockOperationStorage.On("Insert", fixNewRemoveInstanceOperation()).
		Return(nil).Once()
	ts.mockOperationStorage.On("UpdateStateDesc", fixInstanceID(), fixOperationID(), internal.OperationStateSucceeded, fixDeprovisionSucceeded()).
		Return(nil).Once()
	ts.mockAppFinder.On("FindOneByServiceID", fixAppServiceID()).
		Return(fixApp(), nil).Once()

	sut := NewDeprovisioner(
		ts.mockInstanceStorage,
		ts.mockInstanceStateGetter,
		ts.mockOperationStorage,
		ts.mockOperationStorage,
		ts.OpIDProviderFake,
		ts.mockAppFinder,
		ts.knClient,
		ts.eaClient.ApplicationconnectorV1alpha1(),
		spy.NewLogDummy(),
		&IDSelector{false},
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

func assertMatcherReturnTrueOnInstance(t *testing.T, instance *internal.Instance) func(args mock.Arguments) {
	return func(args mock.Arguments) {
		matcher := args.Get(0).(func(*internal.Instance) bool)
		assert.True(t, matcher(instance))
	}
}

func TestErrorInstanceNotFound(t *testing.T) {
	t.Run("when getting the instance by ID", func(t *testing.T) {
		// GIVEN
		ts := newDeprovisionServiceTestSuite(t)
		defer ts.AssertExpectations(t)

		ts.mockInstanceStateGetter.On("IsDeprovisioned", fixInstanceID()).Return(false, nil).Once()
		ts.mockInstanceStateGetter.On("IsDeprovisioningInProgress", fixInstanceID()).Return(internal.OperationID(""), false, nil).Once()
		ts.mockInstanceStorage.On("Get", fixInstanceID()).Return(nil, mockNotFoundError{})

		sut := NewDeprovisioner(
			ts.mockInstanceStorage,
			ts.mockInstanceStateGetter,
			ts.mockOperationStorage,
			nil,
			nil,
			nil,
			nil,
			nil,
			spy.NewLogDummy(),
			&IDSelector{false},
		)

		// WHEN
		_, err := sut.Deprovision(context.Background(), osbContext{}, fixDeprovisionRequest())

		// THEN
		assert.Error(t, err)
		assert.True(t, IsNotFoundError(err))
	})

	t.Run("when removing the instance by ID", func(t *testing.T) {
		// GIVEN
		ts := newDeprovisionServiceTestSuite(t)
		defer ts.AssertExpectations(t)

		ts.mockInstanceStateGetter.On("IsDeprovisioned", fixInstanceID()).Return(false, nil).Once()
		ts.mockInstanceStateGetter.On("IsDeprovisioningInProgress", fixInstanceID()).Return(internal.OperationID(""), false, nil).Once()
		ts.mockInstanceStorage.On("Get", fixInstanceID()).Return(fixNewInstance(), nil)
		ts.mockInstanceStorage.On("FindAll", mock.Anything).Return([]*internal.Instance{fixNewInstance()}, nil).Once()
		ts.mockInstanceStorage.On("Remove", fixInstanceID()).Return(mockNotFoundError{})

		sut := NewDeprovisioner(
			ts.mockInstanceStorage,
			ts.mockInstanceStateGetter,
			ts.mockOperationStorage,
			nil,
			nil,
			nil,
			nil,
			nil,
			spy.NewLogDummy(),
			&IDSelector{false},
		)

		// WHEN
		_, err := sut.Deprovision(context.Background(), osbContext{}, fixDeprovisionRequest())

		// THEN
		assert.Error(t, err)
		assert.True(t, IsNotFoundError(err))
	})

}

func TestGenericErrorOnRemovingInstance(t *testing.T) {
	// GIVEN
	ts := newDeprovisionServiceTestSuite(t)
	defer ts.AssertExpectations(t)

	ts.OpIDProviderFake = func() (internal.OperationID, error) {
		return fixOperationID(), nil
	}
	ts.mockInstanceStorage.On("Get", fixInstanceID()).Return(fixNewInstance(), nil)
	ts.mockInstanceStorage.On("FindAll", mock.Anything).Return([]*internal.Instance{fixNewInstance()}, nil).Once()
	ts.mockInstanceStorage.On("Remove", fixInstanceID()).Return(errors.New("simple error"))
	ts.mockInstanceStateGetter.On("IsDeprovisioned", fixInstanceID()).Return(false, nil).Once()
	ts.mockInstanceStateGetter.On("IsDeprovisioningInProgress", fixInstanceID()).Return(internal.OperationID(""), false, nil).Once()

	sut := NewDeprovisioner(
		ts.mockInstanceStorage,
		ts.mockInstanceStateGetter,
		ts.mockOperationStorage,
		nil,
		nil,
		nil,
		nil,
		nil,
		spy.NewLogDummy(),
		&IDSelector{false},
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
		nil,
		spy.NewLogDummy(),
		nil,
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
		nil,
		spy.NewLogDummy(),
		nil,
	)

	// WHEN
	_, err := sut.Deprovision(context.Background(), osbContext{}, fixDeprovisionRequest())

	// THEN
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checking if instance is being deprovisioned")
}

func TestDoDeprovision(t *testing.T) {
	var (
		iID      = fixInstanceID()
		opID     = fixOperationID()
		appNs    = fixNs()
		appName  = fixAppName()
		appSvcID = internal.ApplicationServiceID("123")
	)

	testCases := map[string]struct {
		initialObjs   []runtime.Object
		expectUpdates []runtime.Object
		expectDeletes []string
	}{
		"Nothing to deprovision": {
			initialObjs:   nil,
			expectUpdates: nil,
			expectDeletes: []string{
				fmt.Sprintf("%s/%s", appNs, appSvcID), // deleting event activation is always triggered but not found error does not fail deprovisioning
			},
		},
		// TODO(nachtmaar): Deprovision broker: https://github.com/kyma-project/kyma/issues/6342
		"Everything gets deprovisioned": {
			initialObjs: []runtime.Object{
				bt.NewAppNamespace(string(appNs), true),
				bt.NewDefaultBroker(string(appNs)),
				bt.NewAppSubscription(string(appNs), string(appName), bt.WithNameSuffix(bt.FakeSubscriptionName)),
				bt.NewEventActivation(string(appNs), string(iID)),
			},
			expectDeletes: []string{
				fmt.Sprintf("%s/%s-%s", integrationNamespace, knSubscriptionNamePrefix, bt.FakeSubscriptionName), // subscription
				fmt.Sprintf("%s/%s", appNs, appSvcID), // event activation
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var (
				instance = fixNewInstance()
			)

			mockOpUpdater := &automock.OperationStorage{}
			mockOpUpdater.On("UpdateStateDesc", iID, opID,
				internal.OperationStateSucceeded,
				fixDeprovisionSucceeded(),
			).Return(nil).Once()
			mockInstanceStorage := &automock.InstanceStorage{}
			mockInstanceStorage.On("Remove", iID).Return(nil).Once()

			/* The third return value is an istio client, which is ignored in case of deprovisioning because it's used
			only to create istio policy in case of provisioning in order to enable Prometheus scraping which is required
			even in case of deprovisioning
			*/
			knCli, k8sCli, _, eaClient := bt.NewFakeClients(tc.initialObjs...)

			dpr := NewDeprovisioner(
				mockInstanceStorage,
				nil,
				nil,
				mockOpUpdater,
				nil,
				nil,
				knative.NewClient(knCli, k8sCli),
				eaClient.ApplicationconnectorV1alpha1(),
				spy.NewLogDummy(),
				nil,
			)

			// WHEN
			dpr.do(instance, opID, appSvcID, appName)

			//THEN
			actionsAsserter := bt.NewActionsAsserter(t, knCli, k8sCli, eaClient)
			actionsAsserter.AssertUpdates(t, tc.expectUpdates)
			actionsAsserter.AssertDeletes(t, tc.expectDeletes)
			mockOpUpdater.AssertExpectations(t)
			mockInstanceStorage.AssertExpectations(t)
		})
	}
}

func newDeprovisionServiceTestSuite(t *testing.T) *deprovisionServiceTestSuite {
	knCli, k8sCli, _, eaCli := bt.NewFakeClients()

	return &deprovisionServiceTestSuite{
		t:                       t,
		mockInstanceStateGetter: &automock.InstanceStateGetter{},
		mockInstanceStorage:     &automock.InstanceStorage{},
		mockOperationStorage:    &automock.OperationStorage{},
		mockAppFinder:           &automock.AppFinder{},
		knClient:                knative.NewClient(knCli, k8sCli),
		eaClient:                eaCli,
	}
}

type deprovisionServiceTestSuite struct {
	t                       *testing.T
	mockInstanceStateGetter *automock.InstanceStateGetter
	mockInstanceStorage     *automock.InstanceStorage
	mockOperationStorage    *automock.OperationStorage
	OpIDProviderFake        func() (internal.OperationID, error)
	mockAppFinder           *automock.AppFinder
	knClient                knative.Client
	eaClient                *eaFake.Clientset
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
