package broker

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	accessAutomock "github.com/kyma-project/kyma/components/application-broker/internal/access/automock"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker/automock"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"

	"fmt"

	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger/spy"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8testing "k8s.io/client-go/testing"
)

func TestProvisionAsync(t *testing.T) {

	type testCase struct {
		name                           string
		givenCanProvisionOutput        access.CanProvisionOutput
		givenCanProvisionError         error
		expectedOpState                internal.OperationState
		expectedOpDesc                 string
		expectedEventActivationCreated bool
		expectedInstanceState          internal.InstanceState
	}

	for _, tc := range []testCase{
		{
			name:                           "success",
			givenCanProvisionOutput:        access.CanProvisionOutput{Allowed: true},
			expectedOpState:                internal.OperationStateSucceeded,
			expectedOpDesc:                 "provisioning succeeded",
			expectedEventActivationCreated: true,
			expectedInstanceState:          internal.InstanceStateSucceeded,
		},
		{
			name:                           "cannot provision",
			givenCanProvisionOutput:        access.CanProvisionOutput{Allowed: false, Reason: "very important reason"},
			expectedOpState:                internal.OperationStateFailed,
			expectedOpDesc:                 "Forbidden provisioning instance [inst-123] for application [name: ec-prod, id: service-id] in namespace: [example-namesapce]. Reason: [very important reason]",
			expectedEventActivationCreated: false,
			expectedInstanceState:          internal.InstanceStateFailed,
		},
		{
			name:                           "error on access checking",
			givenCanProvisionError:         errors.New("some error"),
			expectedOpState:                internal.OperationStateFailed,
			expectedOpDesc:                 "provisioning failed on error: some error",
			expectedEventActivationCreated: false,
			expectedInstanceState:          internal.InstanceStateFailed,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// GIVEN
			mockInstanceStorage := &automock.InstanceStorage{}
			defer mockInstanceStorage.AssertExpectations(t)
			mockStateGetter := &automock.InstanceStateGetter{}
			defer mockStateGetter.AssertExpectations(t)
			mockOperationStorage := &automock.OperationStorage{}
			defer mockOperationStorage.AssertExpectations(t)
			mockAccessChecker := &accessAutomock.ProvisionChecker{}
			defer mockAccessChecker.AssertExpectations(t)
			mockAppFinder := &automock.AppFinder{}
			defer mockAppFinder.AssertExpectations(t)
			mockServiceInstanceGetter := &automock.ServiceInstanceGetter{}
			defer mockServiceInstanceGetter.AssertExpectations(t)

			clientset := fake.NewSimpleClientset()

			defaultWaitTime := time.Minute

			mockStateGetter.On("IsProvisioned", fixInstanceID()).
				Return(false, nil).Once()

			mockStateGetter.On("IsProvisioningInProgress", fixInstanceID()).
				Return(internal.OperationID(""), false, nil)

			mockOperationIDProvider := func() (internal.OperationID, error) {
				return fixOperationID(), nil
			}

			instanceOperation := fixNewCreateInstanceOperation()
			mockOperationStorage.On("Insert", instanceOperation).
				Return(nil)

			mockOperationStorage.On("UpdateStateDesc", fixInstanceID(), fixOperationID(), tc.expectedOpState, &tc.expectedOpDesc).
				Return(nil)

			instance := fixNewInstance()
			mockInstanceStorage.On("Insert", instance).
				Return(nil)

			mockAccessChecker.On("CanProvision", fixInstanceID(), internal.ApplicationServiceID(fixServiceID()), internal.Namespace(fixNs()), defaultWaitTime).
				Return(tc.givenCanProvisionOutput, tc.givenCanProvisionError)

			mockAppFinder.On("FindOneByServiceID", internal.ApplicationServiceID(fixServiceID())).
				Return(fixApp(), nil).
				Once()

			mockInstanceStorage.On("UpdateState", fixInstanceID(), tc.expectedInstanceState).
				Return(nil).
				Once()

			if tc.expectedEventActivationCreated {
				mockServiceInstanceGetter.On("GetByNamespaceAndExternalID", fixNs(), string(fixInstanceID())).Return(FixServiceInstance(), nil)
			}

			sut := NewProvisioner(mockInstanceStorage, mockInstanceStorage,
				mockStateGetter,
				mockOperationStorage,
				mockOperationStorage,
				mockAccessChecker,
				mockAppFinder,
				mockServiceInstanceGetter,
				clientset.ApplicationconnectorV1alpha1(),
				mockInstanceStorage,
				mockOperationIDProvider, spy.NewLogDummy())

			asyncFinished := make(chan struct{}, 0)
			sut.asyncHook = func() {
				asyncFinished <- struct{}{}
			}

			// WHEN
			actResp, err := sut.Provision(context.Background(), osbContext{}, fixProvisionRequest())

			// THEN
			assert.Nil(t, err)
			assert.NotNil(t, actResp)
			assert.True(t, actResp.Async)
			expOpID := osb.OperationKey(fixOperationID())
			assert.Equal(t, &expOpID, actResp.OperationKey)

			select {
			case <-asyncFinished:
				if tc.expectedEventActivationCreated == true {
					eventActivation, err := sut.eaClient.EventActivations(fixNs()).Get(fixServiceID(), v1.GetOptions{})
					assert.Nil(t, err)
					assert.Equal(t, fixEventActivation(), eventActivation)
				}
			case <-time.After(time.Second):
				assert.Fail(t, "Async processing not finished")
			}
		})
	}
}

func TestProvisionWhenAlreadyProvisioned(t *testing.T) {
	// GIVEN
	mockStateGetter := &automock.InstanceStateGetter{}
	defer mockStateGetter.AssertExpectations(t)
	mockStateGetter.On("IsProvisioned", fixInstanceID()).Return(true, nil)

	sut := NewProvisioner(nil, nil, mockStateGetter, nil, nil, nil, nil, nil, nil, nil, nil, spy.NewLogDummy())
	// WHEN
	actResp, err := sut.Provision(context.Background(), osbContext{}, fixProvisionRequest())

	// THEN
	assert.Nil(t, err)
	assert.NotNil(t, actResp)
	assert.False(t, actResp.Async)
}

func TestProvisionWhenProvisioningInProgress(t *testing.T) {
	// GIVEN
	mockStateGetter := &automock.InstanceStateGetter{}
	defer mockStateGetter.AssertExpectations(t)
	mockStateGetter.On("IsProvisioned", fixInstanceID()).Return(false, nil)
	mockStateGetter.On("IsProvisioningInProgress", fixInstanceID()).Return(fixOperationID(), true, nil)

	sut := NewProvisioner(nil, nil, mockStateGetter, nil, nil, nil, nil, nil, nil, nil, nil, spy.NewLogDummy()) // WHEN
	actResp, err := sut.Provision(context.Background(), osbContext{}, fixProvisionRequest())

	// THEN
	assert.Nil(t, err)
	assert.NotNil(t, actResp)
	assert.True(t, actResp.Async)

	expOpKey := osb.OperationKey(fixOperationID())
	assert.Equal(t, &expOpKey, actResp.OperationKey)
}

func TestProvisionCreatingEventActivation(t *testing.T) {
	// GIVEN
	defaultWaitTime := time.Minute

	tests := map[string]func(cli *fake.Clientset, instStorage *automock.InstanceStorage, optStorage *automock.OperationStorage){
		"generic error when creating EA": func(cli *fake.Clientset, instStorage *automock.InstanceStorage, optStorage *automock.OperationStorage) {
			cli.PrependReactor("create", "eventactivations", failingReactor)
			optStorage.On("UpdateStateDesc", fixInstanceID(), fixOperationID(), internal.OperationStateFailed, fixErrWhileCreatingEA()).
				Return(nil)
			instStorage.On("UpdateState", fixInstanceID(), internal.InstanceStateFailed).
				Return(nil).
				Once()
		},
		"EA already exist error": func(cli *fake.Clientset, instStorage *automock.InstanceStorage, optStorage *automock.OperationStorage) {
			cli.PrependReactor("create", "eventactivations", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, apiErrors.NewAlreadyExists(schema.GroupResource{}, "fix")
			})
			optStorage.On("UpdateStateDesc", fixInstanceID(), fixOperationID(), internal.OperationStateSucceeded, ptrStr("provisioning succeeded")).
				Return(nil)
			instStorage.On("UpdateState", fixInstanceID(), internal.InstanceStateSucceeded).
				Return(nil).Once()
		},
		"generic error when updating EA after already exist error": func(cli *fake.Clientset, instStorage *automock.InstanceStorage, optStorage *automock.OperationStorage) {
			cli.PrependReactor("create", "eventactivations", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, apiErrors.NewAlreadyExists(schema.GroupResource{}, "fix")
			})
			cli.PrependReactor("update", "eventactivations", failingReactor)
			optStorage.On("UpdateStateDesc", fixInstanceID(), fixOperationID(), internal.OperationStateFailed, fixErrWhileUpdatingEA()).
				Return(nil)
			instStorage.On("UpdateState", fixInstanceID(), internal.InstanceStateFailed).
				Return(nil).
				Once()
		},
		"generic error when getting EA after already exist error": func(cli *fake.Clientset, instStorage *automock.InstanceStorage, optStorage *automock.OperationStorage) {
			cli.PrependReactor("create", "eventactivations", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, apiErrors.NewAlreadyExists(schema.GroupResource{}, "fix")
			})
			cli.PrependReactor("get", "eventactivations", failingReactor)
			optStorage.On("UpdateStateDesc", fixInstanceID(), fixOperationID(), internal.OperationStateFailed, fixErrWhileGettingEA()).
				Return(nil)
			instStorage.On("UpdateState", fixInstanceID(), internal.InstanceStateFailed).
				Return(nil).
				Once()
		},
	}
	for tn, setupMocks := range tests {
		t.Run(tn, func(t *testing.T) {
			// GIVEN
			mockStateGetter := &automock.InstanceStateGetter{}
			defer mockStateGetter.AssertExpectations(t)
			mockOperationStorage := &automock.OperationStorage{}
			defer mockOperationStorage.AssertExpectations(t)
			mockAccessChecker := &accessAutomock.ProvisionChecker{}
			defer mockAccessChecker.AssertExpectations(t)
			mockAppFinder := &automock.AppFinder{}
			defer mockAppFinder.AssertExpectations(t)
			mockServiceInstanceGetter := &automock.ServiceInstanceGetter{}
			defer mockServiceInstanceGetter.AssertExpectations(t)
			clientset := fake.NewSimpleClientset(fixEventActivation())

			mockStateGetter.On("IsProvisioned", fixInstanceID()).
				Return(false, nil).
				Once()

			mockStateGetter.On("IsProvisioningInProgress", fixInstanceID()).
				Return(internal.OperationID(""), false, nil)

			mockOperationIDProvider := func() (internal.OperationID, error) {
				return fixOperationID(), nil
			}

			instanceOperation := fixNewCreateInstanceOperation()
			mockOperationStorage.On("Insert", instanceOperation).
				Return(nil)

			instance := fixNewInstance()
			mockInstanceStorage := &automock.InstanceStorage{}
			mockInstanceStorage.On("Insert", instance).Return(nil)
			defer mockInstanceStorage.AssertExpectations(t)

			mockAppFinder.On("FindOneByServiceID", internal.ApplicationServiceID(fixServiceID())).
				Return(fixApp(), nil).
				Once()

			mockAccessChecker.On("CanProvision", fixInstanceID(), internal.ApplicationServiceID(fixServiceID()), internal.Namespace(fixNs()), defaultWaitTime).
				Return(access.CanProvisionOutput{Allowed: true}, nil)

			mockServiceInstanceGetter.On("GetByNamespaceAndExternalID", fixNs(), string(fixInstanceID())).Return(FixServiceInstance(), nil)

			setupMocks(clientset, mockInstanceStorage, mockOperationStorage)
			sut := NewProvisioner(mockInstanceStorage, mockInstanceStorage,
				mockStateGetter,
				mockOperationStorage,
				mockOperationStorage,
				mockAccessChecker,
				mockAppFinder,
				mockServiceInstanceGetter,
				clientset.ApplicationconnectorV1alpha1(),
				mockInstanceStorage,
				mockOperationIDProvider, spy.NewLogDummy())

			asyncFinished := make(chan struct{}, 0)
			sut.asyncHook = func() {
				asyncFinished <- struct{}{}
			}

			// WHEN
			_, err := sut.Provision(context.Background(), osbContext{}, fixProvisionRequest())
			assert.Nil(t, err)

			// THEN
			select {
			case <-asyncFinished:
			case <-time.After(time.Second):
				assert.Fail(t, "Async processing not finished")
			}
		})
	}
}

func TestProvisionErrorOnGettingServiceInstance(t *testing.T) {
	// GIVEN
	mockInstanceStorage := &automock.InstanceStorage{}
	defer mockInstanceStorage.AssertExpectations(t)
	mockStateGetter := &automock.InstanceStateGetter{}
	defer mockStateGetter.AssertExpectations(t)
	mockOperationStorage := &automock.OperationStorage{}
	defer mockOperationStorage.AssertExpectations(t)
	mockAccessChecker := &accessAutomock.ProvisionChecker{}
	defer mockAccessChecker.AssertExpectations(t)
	mockAppFinder := &automock.AppFinder{}
	defer mockAppFinder.AssertExpectations(t)
	mockServiceInstanceGetter := &automock.ServiceInstanceGetter{}
	defer mockServiceInstanceGetter.AssertExpectations(t)

	clientset := fake.NewSimpleClientset()

	defaultWaitTime := time.Minute

	mockStateGetter.On("IsProvisioned", fixInstanceID()).
		Return(false, nil).Once()

	mockStateGetter.On("IsProvisioningInProgress", fixInstanceID()).
		Return(internal.OperationID(""), false, nil)

	mockOperationIDProvider := func() (internal.OperationID, error) {
		return fixOperationID(), nil
	}

	instanceOperation := fixNewCreateInstanceOperation()
	mockOperationStorage.On("Insert", instanceOperation).
		Return(nil)

	instance := fixNewInstance()
	mockInstanceStorage.On("Insert", instance).
		Return(nil)

	mockAppFinder.On("FindOneByServiceID", internal.ApplicationServiceID(fixServiceID())).
		Return(fixApp(), nil).
		Once()

	mockAccessChecker.On("CanProvision", fixInstanceID(), internal.ApplicationServiceID(fixServiceID()), internal.Namespace(fixNs()), defaultWaitTime).
		Return(access.CanProvisionOutput{Allowed: true}, nil)

	mockServiceInstanceGetter.On("GetByNamespaceAndExternalID", fixNs(), string(fixInstanceID())).Return(nil, errors.New("custom error"))

	mockInstanceStorage.On("UpdateState", fixInstanceID(), internal.InstanceStateFailed).
		Return(nil).
		Once()

	mockOperationStorage.On("UpdateStateDesc", fixInstanceID(), fixOperationID(), internal.OperationStateFailed, fixErrWhileGettingServiceInstance()).
		Return(nil)

	sut := NewProvisioner(mockInstanceStorage, mockInstanceStorage,
		mockStateGetter,
		mockOperationStorage,
		mockOperationStorage,
		mockAccessChecker,
		mockAppFinder,
		mockServiceInstanceGetter,
		clientset.ApplicationconnectorV1alpha1(),
		mockInstanceStorage,
		mockOperationIDProvider, spy.NewLogDummy())

	asyncFinished := make(chan struct{}, 0)
	sut.asyncHook = func() {
		asyncFinished <- struct{}{}
	}

	// WHEN
	_, err := sut.Provision(context.Background(), osbContext{}, fixProvisionRequest())
	assert.Nil(t, err)

	// THEN
	select {
	case <-asyncFinished:
	case <-time.After(time.Second):
		assert.Fail(t, "Async processing not finished")
	}
}

func TestProvisionErrorOnCheckingIfProvisioned(t *testing.T) {
	// GIVEN
	mockStateGetter := &automock.InstanceStateGetter{}
	defer mockStateGetter.AssertExpectations(t)
	mockStateGetter.On("IsProvisioned", fixInstanceID()).Return(false, fixError())

	sut := NewProvisioner(nil, nil, mockStateGetter, nil, nil, nil, nil, nil, nil, nil, nil, spy.NewLogDummy())
	// WHEN
	_, err := sut.Provision(context.Background(), osbContext{}, fixProvisionRequest())

	// THEN
	assert.Error(t, err)
}

func TestProvisionErrorOnCheckingIfProvisionInProgress(t *testing.T) {
	// GIVEN
	mockStateGetter := &automock.InstanceStateGetter{}
	defer mockStateGetter.AssertExpectations(t)
	mockStateGetter.On("IsProvisioned", fixInstanceID()).Return(false, nil)
	mockStateGetter.On("IsProvisioningInProgress", fixInstanceID()).Return(internal.OperationID(""), false, fixError())

	sut := NewProvisioner(nil, nil, mockStateGetter, nil, nil, nil, nil, nil, nil, nil, nil, spy.NewLogDummy())
	// WHEN
	_, err := sut.Provision(context.Background(), osbContext{}, fixProvisionRequest())

	// THEN
	assert.Error(t, err)
}

func TestProvisionErrorOnIDGeneration(t *testing.T) {
	// GIVEN
	mockStateGetter := &automock.InstanceStateGetter{}
	defer mockStateGetter.AssertExpectations(t)

	mockStateGetter.On("IsProvisioned", fixInstanceID()).
		Return(false, nil).Once()

	mockStateGetter.On("IsProvisioningInProgress", fixInstanceID()).
		Return(internal.OperationID(""), false, nil)

	mockOperationIDProvider := func() (internal.OperationID, error) {
		return "", fixError()
	}
	sut := NewProvisioner(nil, nil, mockStateGetter, nil, nil, nil, nil, nil, nil, nil, mockOperationIDProvider, spy.NewLogDummy())
	// WHEN
	_, err := sut.Provision(context.Background(), osbContext{}, fixProvisionRequest())
	// THEN
	assert.Error(t, err)
}

func TestProvisionErrorOnInsertingOperation(t *testing.T) {
	// GIVEN
	mockStateGetter := &automock.InstanceStateGetter{}
	defer mockStateGetter.AssertExpectations(t)
	mockOperationStorage := &automock.OperationStorage{}
	defer mockOperationStorage.AssertExpectations(t)

	mockStateGetter.On("IsProvisioned", fixInstanceID()).
		Return(false, nil).Once()

	mockStateGetter.On("IsProvisioningInProgress", fixInstanceID()).
		Return(internal.OperationID(""), false, nil)

	mockOperationIDProvider := func() (internal.OperationID, error) {
		return fixOperationID(), nil
	}

	instanceOperation := fixNewCreateInstanceOperation()
	mockOperationStorage.On("Insert", instanceOperation).
		Return(fixError())

	sut := NewProvisioner(nil, nil,
		mockStateGetter,
		mockOperationStorage,
		mockOperationStorage,
		nil,
		nil,
		nil,
		nil,
		nil,
		mockOperationIDProvider, spy.NewLogDummy())

	// WHEN
	_, err := sut.Provision(context.Background(), osbContext{}, fixProvisionRequest())
	// THEN
	assert.Error(t, err)
}

func TestProvisionErrorOnInsertingInstance(t *testing.T) {
	// GIVEN
	mockInstanceStorage := &automock.InstanceStorage{}
	defer mockInstanceStorage.AssertExpectations(t)
	mockStateGetter := &automock.InstanceStateGetter{}
	defer mockStateGetter.AssertExpectations(t)
	mockOperationStorage := &automock.OperationStorage{}
	defer mockOperationStorage.AssertExpectations(t)
	mockAppFinder := &automock.AppFinder{}
	defer mockAppFinder.AssertExpectations(t)

	mockStateGetter.On("IsProvisioned", fixInstanceID()).
		Return(false, nil).Once()

	mockStateGetter.On("IsProvisioningInProgress", fixInstanceID()).
		Return(internal.OperationID(""), false, nil)

	mockOperationIDProvider := func() (internal.OperationID, error) {
		return fixOperationID(), nil
	}

	instanceOperation := fixNewCreateInstanceOperation()
	mockOperationStorage.On("Insert", instanceOperation).
		Return(nil)

	instance := fixNewInstance()
	mockInstanceStorage.On("Insert", instance).Return(fixError())

	mockAppFinder.On("FindOneByServiceID", internal.ApplicationServiceID(fixServiceID())).
		Return(fixApp(), nil).
		Once()

	sut := NewProvisioner(mockInstanceStorage, mockInstanceStorage,
		mockStateGetter,
		mockOperationStorage,
		mockOperationStorage,
		nil,
		mockAppFinder,
		nil,
		nil,
		nil,
		mockOperationIDProvider, spy.NewLogDummy())

	// WHEN
	_, err := sut.Provision(context.Background(), osbContext{}, fixProvisionRequest())
	// THEN
	assert.Error(t, err)

}

func failingReactor(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
	return true, nil, errors.New("custom error")
}

func fixErrWhileCreatingEA() *string {
	err := fmt.Sprintf("provisioning failed while creating EventActivation on error: while creating EventActivation with name: %q in namespace: %q: custom error", fixServiceID(), fixNs())
	return &err
}

func fixErrWhileUpdatingEA() *string {
	err := fmt.Sprintf("provisioning failed while creating EventActivation on error: while ensuring update on EventActivation: while updating EventActivation with name: %q in namespace: %q: custom error", fixServiceID(), fixNs())
	return &err
}

func fixErrWhileGettingEA() *string {
	err := fmt.Sprintf("provisioning failed while creating EventActivation on error: while ensuring update on EventActivation: while getting EventActivation with name: %q from namespace: %q: custom error", fixServiceID(), fixNs())
	return &err
}

func fixErrWhileGettingServiceInstance() *string {
	err := fmt.Sprintf("provisioning failed while creating EventActivation on error: while getting service instance with external id: %q in namespace: %q: custom error", fixInstanceID(), fixNs())
	return &err
}

func ptrStr(s string) *string {
	return &s
}
