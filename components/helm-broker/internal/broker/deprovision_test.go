package broker_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	osb "github.com/pmorie/go-open-service-broker-client/v2"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker"
	"github.com/kyma-project/kyma/components/helm-broker/internal/broker/automock"
)

func TestDeprovisionServiceDeprovisionSuccess(t *testing.T) {
	// GIVEN
	ts := newDeprovisionServiceTestSuite(t)
	ts.SetUp()

	defer ts.AssertExpectations(t)

	ts.InstStateGetterMock.ExpectOnIsDeprovisioned(ts.Exp.InstanceID, false).Once()
	ts.InstStateGetterMock.ExpectOnIsDeprovisioningInProgress(ts.Exp.InstanceID, internal.OperationID(""), false).Once()

	ts.InstStorageMock.ExpectOnGet(ts.Exp.InstanceID, ts.FixInstance()).Once()

	ts.OpStorageMock.ExpectOnInsert(ts.FixInstanceOperation()).Once()
	ts.OpStorageMock.ExpectOnUpdateStateDesc(ts.Exp.InstanceID, ts.Exp.OperationID, internal.OperationStateSucceeded, "deprovisioning succeeded").
		Run(func(args mock.Arguments) {
			close(ts.UpdateStateDescMethodCalled)
		}).Once()

	ts.HelmClientMock.ExpectOnDelete(ts.Exp.ReleaseName).Once()

	ts.InstBindDataMock.ExpectOnRemove(ts.Exp.InstanceID).Once()

	ts.OpIDProviderFake = func() (internal.OperationID, error) {
		return ts.Exp.OperationID, nil
	}

	svc := broker.NewDeprovisionService(ts.GetAllMocks())

	osbCtx := *broker.NewOSBContext("", "v1")
	req := ts.FixDeprovisionRequest()

	// WHEN
	resp, err := svc.Deprovision(context.Background(), osbCtx, &req)

	// THEN
	assert.NoError(t, err)
	assert.True(t, resp.Async)
	assert.EqualValues(t, ts.Exp.OperationID, *resp.OperationKey)

	select {
	case <-ts.UpdateStateDescMethodCalled:
	case <-time.After(time.Millisecond * 100):
		t.Fatal("timeout on operation succeeded")
	}
}

func TestDeprovisionServiceDeprovisionFailureAsync(t *testing.T) {
	fixErr := errors.New("fake Err")

	for tn, testCaseSetExpectionOnMocks := range map[string]func(ts *deprovisionServiceTestSuite){
		"on Helm Delete": func(ts *deprovisionServiceTestSuite) {
			ts.InstStateGetterMock.ExpectOnIsDeprovisioned(ts.Exp.InstanceID, false).Once()
			ts.InstStateGetterMock.ExpectOnIsDeprovisioningInProgress(ts.Exp.InstanceID, internal.OperationID(""), false).Once()

			ts.InstStorageMock.ExpectOnGet(ts.Exp.InstanceID, ts.FixInstance()).Once()

			ts.OpStorageMock.ExpectOnInsert(ts.FixInstanceOperation()).Once()
			expDesc := fmt.Sprintf("deprovisioning failed on error: while deleting helm release: %s", fixErr)
			ts.OpStorageMock.ExpectOnUpdateStateDesc(ts.Exp.InstanceID, ts.Exp.OperationID, internal.OperationStateFailed, expDesc).
				Run(func(args mock.Arguments) {
					close(ts.UpdateStateDescMethodCalled)
				}).Once()

			ts.HelmClientMock.ExpectErrorOnDelete(ts.Exp.ReleaseName, fixErr).Once()
		},
		"on bind data Remove": func(ts *deprovisionServiceTestSuite) {
			ts.InstStateGetterMock.ExpectOnIsDeprovisioned(ts.Exp.InstanceID, false).Once()
			ts.InstStateGetterMock.ExpectOnIsDeprovisioningInProgress(ts.Exp.InstanceID, internal.OperationID(""), false).Once()

			ts.InstStorageMock.ExpectOnGet(ts.Exp.InstanceID, ts.FixInstance()).Once()

			ts.OpStorageMock.ExpectOnInsert(ts.FixInstanceOperation()).Once()
			expDesc := fmt.Sprintf("deprovisioning failed on error: cannot remove instance bind data from storage: %s", fixErr)
			ts.OpStorageMock.ExpectOnUpdateStateDesc(ts.Exp.InstanceID, ts.Exp.OperationID, internal.OperationStateFailed, expDesc).
				Run(func(args mock.Arguments) {
					close(ts.UpdateStateDescMethodCalled)
				}).Once()

			ts.HelmClientMock.ExpectOnDelete(ts.Exp.ReleaseName).Once()

			ts.InstBindDataMock.ExpectErrorRemove(ts.Exp.InstanceID, fixErr).Once()
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// GIVEN
			ts := newDeprovisionServiceTestSuite(t)
			ts.SetUp()

			defer ts.AssertExpectations(t)

			testCaseSetExpectionOnMocks(ts)

			ts.OpIDProviderFake = func() (internal.OperationID, error) {
				return ts.Exp.OperationID, nil
			}

			svc := broker.NewDeprovisionService(ts.GetAllMocks())

			osbCtx := *broker.NewOSBContext("", "v1")
			req := ts.FixDeprovisionRequest()

			// WHEN
			resp, err := svc.Deprovision(context.Background(), osbCtx, &req)

			// THEN
			assert.NoError(t, err)
			assert.True(t, resp.Async)
			assert.EqualValues(t, ts.Exp.OperationID, *resp.OperationKey)

			select {
			case <-ts.UpdateStateDescMethodCalled:
			case <-time.After(time.Millisecond * 100):
				t.Fatal("timeout on operation failed")
			}
		})
	}

}

func TestDeprovisionServiceDeprovisionSuccessOnAlreadyDeprovisionedInstance(t *testing.T) {
	// GIVEN
	ts := newDeprovisionServiceTestSuite(t)
	ts.SetUp()

	defer ts.AssertExpectations(t)

	ts.InstStateGetterMock.ExpectOnIsDeprovisioned(ts.Exp.InstanceID, true).Once()

	ts.OpIDProviderFake = func() (internal.OperationID, error) {
		t.Error("operation ID provider called when it should not be")
		return ts.Exp.OperationID, nil
	}

	testHookCalled := make(chan struct{})

	svc := broker.NewDeprovisionService(ts.GetAllMocks()).
		WithTestHookOnAsyncCalled(func(internal.OperationID) { close(testHookCalled) })

	osbCtx := *broker.NewOSBContext("", "v1")
	req := ts.FixDeprovisionRequest()

	// WHEN
	resp, err := svc.Deprovision(context.Background(), osbCtx, &req)

	// THEN
	assert.NoError(t, err)
	assert.False(t, resp.Async)
	assert.Nil(t, resp.OperationKey)

	select {
	case <-testHookCalled:
		t.Fatal("async test hook called")
	default:
	}
}

func TestDeprovisionServiceDeprovisionSuccessOnDeprovisioningInProgressInstance(t *testing.T) {
	// GIVEN
	ts := newDeprovisionServiceTestSuite(t)
	ts.SetUp()

	defer ts.AssertExpectations(t)

	ts.InstStateGetterMock.ExpectOnIsDeprovisioned(ts.Exp.InstanceID, false).Once()
	ts.InstStateGetterMock.ExpectOnIsDeprovisioningInProgress(ts.Exp.InstanceID, ts.Exp.OperationID, true).Once()

	ts.OpIDProviderFake = func() (internal.OperationID, error) {
		t.Error("operation ID provider called when it should not be")
		return ts.Exp.OperationID, nil
	}

	testHookCalled := make(chan struct{})

	svc := broker.NewDeprovisionService(ts.GetAllMocks()).
		WithTestHookOnAsyncCalled(func(internal.OperationID) { close(testHookCalled) })

	osbCtx := *broker.NewOSBContext("", "v1")
	req := ts.FixDeprovisionRequest()

	// WHEN
	resp, err := svc.Deprovision(context.Background(), osbCtx, &req)

	// THEN
	assert.NoError(t, err)
	assert.True(t, resp.Async)
	assert.EqualValues(t, ts.Exp.OperationID, *resp.OperationKey)

	select {
	case <-testHookCalled:
		t.Fatal("async test hook called")
	default:
	}
}

func TestDeprovisionServiceDeprovisionFailureNotFoundOnIsDeprovisionedCheck(t *testing.T) {
	// GIVEN
	ts := newDeprovisionServiceTestSuite(t)
	ts.SetUp()

	defer ts.AssertExpectations(t)

	ts.InstStateGetterMock.ExpectErrorIsDeprovisioned(ts.Exp.InstanceID, notFoundError{}).Once()

	ts.OpIDProviderFake = func() (internal.OperationID, error) {
		t.Error("operation ID provider called when it should not be")
		return ts.Exp.OperationID, nil
	}

	testHookCalled := make(chan struct{})

	svc := broker.NewDeprovisionService(ts.GetAllMocks()).
		WithTestHookOnAsyncCalled(func(internal.OperationID) { close(testHookCalled) })

	osbCtx := *broker.NewOSBContext("", "v1")
	req := ts.FixDeprovisionRequest()

	// WHEN
	_, err := svc.Deprovision(context.Background(), osbCtx, &req)

	// THEN
	assert.True(t, broker.IsNotFoundError(err))

	select {
	case <-testHookCalled:
		t.Fatal("async test hook called")
	default:
	}
}

func TestDeprovisionServiceDeprovisionFailureNotFoundOnIsDeprovisioningInProgressCheck(t *testing.T) {
	// GIVEN
	ts := newDeprovisionServiceTestSuite(t)
	ts.SetUp()

	defer ts.AssertExpectations(t)

	ts.InstStateGetterMock.ExpectOnIsDeprovisioned(ts.Exp.InstanceID, false).Once()
	ts.InstStateGetterMock.ExpectErrorOnIsDeprovisioningInProgress(ts.Exp.InstanceID, notFoundError{}).Once()

	ts.OpIDProviderFake = func() (internal.OperationID, error) {
		t.Error("operation ID provider called when it should not be")
		return ts.Exp.OperationID, nil
	}

	testHookCalled := make(chan struct{})

	svc := broker.NewDeprovisionService(ts.GetAllMocks()).
		WithTestHookOnAsyncCalled(func(internal.OperationID) { close(testHookCalled) })

	osbCtx := *broker.NewOSBContext("", "v1")
	req := ts.FixDeprovisionRequest()

	// WHEN
	_, err := svc.Deprovision(context.Background(), osbCtx, &req)

	// THEN
	assert.True(t, broker.IsNotFoundError(err))

	select {
	case <-testHookCalled:
		t.Fatal("async test hook called")
	default:
	}
}

func TestDeprovisionServiceDeprovisionFailureNotFoundOnGettingInstance(t *testing.T) {
	// GIVEN
	ts := newDeprovisionServiceTestSuite(t)
	ts.SetUp()

	defer ts.AssertExpectations(t)

	ts.InstStateGetterMock.ExpectOnIsDeprovisioned(ts.Exp.InstanceID, false).Once()
	ts.InstStateGetterMock.ExpectOnIsDeprovisioningInProgress(ts.Exp.InstanceID, internal.OperationID(""), false).Once()

	ts.InstStorageMock.ExpectErrorOnGet(ts.Exp.InstanceID, notFoundError{})

	ts.OpIDProviderFake = func() (internal.OperationID, error) {
		return ts.Exp.OperationID, nil
	}

	testHookCalled := make(chan struct{})

	svc := broker.NewDeprovisionService(ts.GetAllMocks()).
		WithTestHookOnAsyncCalled(func(internal.OperationID) { close(testHookCalled) })

	osbCtx := *broker.NewOSBContext("", "v1")
	req := ts.FixDeprovisionRequest()

	// WHEN
	_, err := svc.Deprovision(context.Background(), osbCtx, &req)

	// THEN
	assert.True(t, broker.IsNotFoundError(err))

	select {
	case <-testHookCalled:
		t.Fatal("async test hook called")
	default:
	}
}

func newDeprovisionServiceTestSuite(t *testing.T) *deprovisionServiceTestSuite {
	return &deprovisionServiceTestSuite{
		t:                           t,
		InstStateGetterMock:         &automock.InstanceStateGetter{},
		InstStorageMock:             &automock.InstanceStorage{},
		InstBindDataMock:            &automock.InstanceBindDataRemover{},
		OpStorageMock:               &automock.OperationStorage{},
		HelmClientMock:              &automock.HelmClient{},
		UpdateStateDescMethodCalled: make(chan struct{}),
	}
}

type deprovisionServiceTestSuite struct {
	t *testing.T

	Exp expAll

	InstStateGetterMock         *automock.InstanceStateGetter
	InstStorageMock             *automock.InstanceStorage
	OpStorageMock               *automock.OperationStorage
	HelmClientMock              *automock.HelmClient
	InstBindDataMock            *automock.InstanceBindDataRemover
	OpIDProviderFake            func() (internal.OperationID, error)
	UpdateStateDescMethodCalled chan struct{}
}

func (ts *deprovisionServiceTestSuite) AssertExpectations(t *testing.T) {
	ts.InstStateGetterMock.AssertExpectations(t)
	ts.InstStorageMock.AssertExpectations(t)
	ts.InstBindDataMock.AssertExpectations(t)
	ts.OpStorageMock.AssertExpectations(t)
	ts.HelmClientMock.AssertExpectations(t)
}

func (ts *deprovisionServiceTestSuite) GetAllMocks() (*automock.InstanceStorage, *automock.OperationStorage, *automock.OperationStorage, *automock.InstanceBindDataRemover, *automock.HelmClient, func() (internal.OperationID, error), *automock.InstanceStateGetter) {
	return ts.InstStorageMock, ts.OpStorageMock, ts.OpStorageMock, ts.InstBindDataMock, ts.HelmClientMock, ts.OpIDProviderFake, ts.InstStateGetterMock
}

func (ts *deprovisionServiceTestSuite) SetUp() {
	ts.Exp.Populate()
}

func (ts *deprovisionServiceTestSuite) FixBundle() internal.Bundle {
	return *ts.Exp.NewBundle()
}

func (ts *deprovisionServiceTestSuite) FixInstance() internal.Instance {
	return *ts.Exp.NewInstance()
}

func (ts *deprovisionServiceTestSuite) FixInstanceOperation() internal.InstanceOperation {
	return *ts.Exp.NewInstanceOperation(internal.OperationTypeRemove, internal.OperationStateInProgress)
}

func (ts *deprovisionServiceTestSuite) FixDeprovisionRequest() osb.DeprovisionRequest {
	return osb.DeprovisionRequest{
		InstanceID:        string(ts.Exp.InstanceID),
		ServiceID:         string(ts.Exp.Service.ID),
		PlanID:            string(ts.Exp.ServicePlan.ID),
		AcceptsIncomplete: true,
	}
}
