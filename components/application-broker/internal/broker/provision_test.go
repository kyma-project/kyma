package broker

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	accessAutomock "github.com/kyma-project/kyma/components/application-broker/internal/access/automock"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker/automock"
	bt "github.com/kyma-project/kyma/components/application-broker/internal/broker/testing"
	"github.com/kyma-project/kyma/components/application-broker/internal/knative"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	eventingfake "knative.dev/eventing/pkg/client/clientset/versioned/fake"

	"fmt"

	"net/http"

	"github.com/komkom/go-jsonhash"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger/spy"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8testing "k8s.io/client-go/testing"
)

func TestProvisionAsync(t *testing.T) {
	var (
		appNs   = string(fixNs())
		appName = string(fixAppName())
	)

	type testCase struct {
		name                           string
		initialObjs                    []runtime.Object
		givenCanProvisionOutput        access.CanProvisionOutput
		givenCanProvisionError         error
		expectedOpState                internal.OperationState
		expectedOpDesc                 string
		expectedEventActivationCreated bool
		expectedInstanceState          internal.InstanceState
	}

	for _, tc := range []testCase{
		{
			name: "success",
			initialObjs: []runtime.Object{
				bt.NewAppNamespace(appNs, false),
				bt.NewAppChannel(appName),
			},
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
			expectedOpDesc:                 "Forbidden provisioning instance [inst-123] for application [name: ec-prod, id: service-id] in namespace: [" + appNs + "]. Reason: [very important reason]",
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
			instanceOperation.ParamsHash = jsonhash.HashS(map[string]interface{}{})
			mockOperationStorage.On("Insert", instanceOperation).
				Return(nil)

			mockOperationStorage.On("UpdateStateDesc", fixInstanceID(), fixOperationID(), tc.expectedOpState, &tc.expectedOpDesc).
				Return(nil)

			instance := fixNewInstance()
			instance.ParamsHash = jsonhash.HashS(map[string]interface{}{})
			mockInstanceStorage.On("Insert", instance).
				Return(nil)

			mockAccessChecker.On("CanProvision", fixInstanceID(), fixAppServiceID(), fixNs(), defaultWaitTime).
				Return(tc.givenCanProvisionOutput, tc.givenCanProvisionError)

			mockAppFinder.On("FindOneByServiceID", fixAppServiceID()).
				Return(fixApp(), nil).
				Once()

			mockInstanceStorage.On("UpdateState", fixInstanceID(), tc.expectedInstanceState).
				Return(nil).
				Once()

			if tc.expectedEventActivationCreated {
				mockServiceInstanceGetter.On("GetByNamespaceAndExternalID", string(fixNs()), string(fixInstanceID())).Return(FixServiceInstance(), nil)
			}
			knCli, k8sCli := bt.NewFakeClients(tc.initialObjs...)

			sut := NewProvisioner(mockInstanceStorage, mockInstanceStorage,
				mockStateGetter,
				mockOperationStorage,
				mockOperationStorage,
				mockAccessChecker,
				mockAppFinder,
				mockServiceInstanceGetter,
				clientset.ApplicationconnectorV1alpha1(),
				knative.NewClient(knCli, k8sCli),
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
					eventActivation, err := sut.eaClient.EventActivations(string(fixNs())).
						Get(string(fixServiceID()), v1.GetOptions{})
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

	instance := fixNewInstance()
	instance.ParamsHash = jsonhash.HashS(map[string]interface{}{})

	mockInstanceStorage := &automock.InstanceStorage{}
	mockInstanceStorage.On("Get", fixInstanceID()).Return(instance, nil)
	defer mockInstanceStorage.AssertExpectations(t)

	sut := NewProvisioner(nil, mockInstanceStorage, mockStateGetter, nil, nil, nil, nil, nil, nil, nil, nil, nil, spy.NewLogDummy())
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

	instance := fixNewInstance()
	instance.ParamsHash = jsonhash.HashS(map[string]interface{}{})

	mockInstanceStorage := &automock.InstanceStorage{}
	mockInstanceStorage.On("Get", fixInstanceID()).Return(instance, nil)
	defer mockInstanceStorage.AssertExpectations(t)

	sut := NewProvisioner(nil, mockInstanceStorage, mockStateGetter, nil, nil, nil, nil, nil, nil, nil, nil, nil, spy.NewLogDummy()) // WHEN
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
	var (
		defaultWaitTime = time.Minute
		appNs           = string(fixNs())
		appName         = string(fixAppName())
	)

	type setupMocksFunc = func(cli *fake.Clientset, instStorage *automock.InstanceStorage, optStorage *automock.OperationStorage) (*eventingfake.Clientset, *k8sfake.Clientset)

	tests := map[string]setupMocksFunc{
		"generic error when creating EA": func(cli *fake.Clientset, instStorage *automock.InstanceStorage, optStorage *automock.OperationStorage) (*eventingfake.Clientset, *k8sfake.Clientset) {
			cli.PrependReactor("create", "eventactivations", failingReactor)
			optStorage.On("UpdateStateDesc", fixInstanceID(), fixOperationID(), internal.OperationStateFailed, fixErrWhileCreatingEA()).
				Return(nil)
			instStorage.On("UpdateState", fixInstanceID(), internal.InstanceStateFailed).
				Return(nil).
				Once()
			initialObjs := []runtime.Object{
				bt.NewAppNamespace(appNs, false),
				bt.NewAppChannel(appName),
			}
			return bt.NewFakeClients(initialObjs...)
		},
		"EA already exist error": func(cli *fake.Clientset, instStorage *automock.InstanceStorage, optStorage *automock.OperationStorage) (*eventingfake.Clientset, *k8sfake.Clientset) {
			cli.PrependReactor("create", "eventactivations", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, apiErrors.NewAlreadyExists(schema.GroupResource{}, "fix")
			})
			optStorage.On("UpdateStateDesc", fixInstanceID(), fixOperationID(), internal.OperationStateSucceeded, ptrStr(internal.OperationDescriptionProvisioningSucceeded)).
				Return(nil)
			instStorage.On("UpdateState", fixInstanceID(), internal.InstanceStateSucceeded).
				Return(nil).Once()
			initialObjs := []runtime.Object{
				bt.NewAppNamespace(appNs, false),
				bt.NewAppChannel(appName),
			}
			return bt.NewFakeClients(initialObjs...)
		},
		"generic error when updating EA after already exist error": func(cli *fake.Clientset, instStorage *automock.InstanceStorage, optStorage *automock.OperationStorage) (*eventingfake.Clientset, *k8sfake.Clientset) {
			cli.PrependReactor("create", "eventactivations", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, apiErrors.NewAlreadyExists(schema.GroupResource{}, "fix")
			})
			cli.PrependReactor("update", "eventactivations", failingReactor)
			optStorage.On("UpdateStateDesc", fixInstanceID(), fixOperationID(), internal.OperationStateFailed, fixErrWhileUpdatingEA()).
				Return(nil)
			instStorage.On("UpdateState", fixInstanceID(), internal.InstanceStateFailed).
				Return(nil).
				Once()
			initialObjs := []runtime.Object{
				bt.NewAppNamespace(appNs, false),
				bt.NewAppChannel(appName),
			}
			return bt.NewFakeClients(initialObjs...)
		},
		"generic error when getting EA after already exist error": func(cli *fake.Clientset, instStorage *automock.InstanceStorage, optStorage *automock.OperationStorage) (*eventingfake.Clientset, *k8sfake.Clientset) {
			cli.PrependReactor("create", "eventactivations", func(action k8testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, apiErrors.NewAlreadyExists(schema.GroupResource{}, "fix")
			})
			cli.PrependReactor("get", "eventactivations", failingReactor)
			optStorage.On("UpdateStateDesc", fixInstanceID(), fixOperationID(), internal.OperationStateFailed, fixErrWhileGettingEA()).
				Return(nil)
			instStorage.On("UpdateState", fixInstanceID(), internal.InstanceStateFailed).
				Return(nil).
				Once()
			initialObjs := []runtime.Object{
				bt.NewAppNamespace(appNs, false),
				bt.NewAppChannel(appName),
			}
			return bt.NewFakeClients(initialObjs...)
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
			instanceOperation.ParamsHash = jsonhash.HashS(map[string]interface{}{})
			mockOperationStorage.On("Insert", instanceOperation).
				Return(nil)

			instance := fixNewInstance()
			instance.ParamsHash = jsonhash.HashS(map[string]interface{}{})
			mockInstanceStorage := &automock.InstanceStorage{}
			mockInstanceStorage.On("Insert", instance).Return(nil)
			defer mockInstanceStorage.AssertExpectations(t)

			mockAppFinder.On("FindOneByServiceID", internal.ApplicationServiceID(fixServiceID())).
				Return(fixApp(), nil).
				Once()

			mockAccessChecker.On("CanProvision", fixInstanceID(), internal.ApplicationServiceID(fixServiceID()), internal.Namespace(fixNs()), defaultWaitTime).
				Return(access.CanProvisionOutput{Allowed: true}, nil)

			mockServiceInstanceGetter.On("GetByNamespaceAndExternalID", string(fixNs()), string(fixInstanceID())).Return(FixServiceInstance(), nil)

			knCli, k8sCli := setupMocks(clientset, mockInstanceStorage, mockOperationStorage)
			sut := NewProvisioner(
				mockInstanceStorage,
				mockInstanceStorage,
				mockStateGetter,
				mockOperationStorage,
				mockOperationStorage,
				mockAccessChecker,
				mockAppFinder,
				mockServiceInstanceGetter,
				clientset.ApplicationconnectorV1alpha1(),
				knative.NewClient(knCli, k8sCli),
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
	appNs := string(fixNs())
	appName := string(fixAppName())
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
	instanceOperation.ParamsHash = jsonhash.HashS(map[string]interface{}{})
	mockOperationStorage.On("Insert", instanceOperation).
		Return(nil)

	instance := fixNewInstance()
	instance.ParamsHash = jsonhash.HashS(map[string]interface{}{})
	mockInstanceStorage.On("Insert", instance).
		Return(nil)

	mockAppFinder.On("FindOneByServiceID", fixAppServiceID()).
		Return(fixApp(), nil).
		Once()

	mockAccessChecker.On("CanProvision", fixInstanceID(), fixAppServiceID(), fixNs(), defaultWaitTime).
		Return(access.CanProvisionOutput{Allowed: true}, nil)

	mockServiceInstanceGetter.On("GetByNamespaceAndExternalID", string(fixNs()), string(fixInstanceID())).Return(nil, errors.New("custom error"))

	mockInstanceStorage.On("UpdateState", fixInstanceID(), internal.InstanceStateFailed).
		Return(nil).
		Once()

	mockOperationStorage.On("UpdateStateDesc", fixInstanceID(), fixOperationID(), internal.OperationStateFailed, fixErrWhileGettingServiceInstance()).
		Return(nil)

	initialObjs := []runtime.Object{
		bt.NewAppNamespace(appNs, false),
		bt.NewAppChannel(appName),
	}

	knCli, k8sCli := bt.NewFakeClients(initialObjs...)

	sut := NewProvisioner(
		mockInstanceStorage,
		mockInstanceStorage,
		mockStateGetter,
		mockOperationStorage,
		mockOperationStorage,
		mockAccessChecker,
		mockAppFinder,
		mockServiceInstanceGetter,
		clientset.ApplicationconnectorV1alpha1(),
		knative.NewClient(knCli, k8sCli),
		mockInstanceStorage,
		mockOperationIDProvider,
		spy.NewLogDummy(),
	)

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

	sut := NewProvisioner(nil, nil, mockStateGetter, nil, nil, nil, nil, nil, nil, nil, nil, nil, spy.NewLogDummy())
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

	sut := NewProvisioner(nil, nil, mockStateGetter, nil, nil, nil, nil, nil, nil, nil, nil, nil, spy.NewLogDummy())
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
	sut := NewProvisioner(nil, nil, mockStateGetter, nil, nil, nil, nil, nil, nil, nil, nil, mockOperationIDProvider, spy.NewLogDummy())
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
	instanceOperation.ParamsHash = jsonhash.HashS(map[string]interface{}{})
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
	instanceOperation.ParamsHash = jsonhash.HashS(map[string]interface{}{})
	mockOperationStorage.On("Insert", instanceOperation).
		Return(nil)

	instance := fixNewInstance()
	instance.ParamsHash = jsonhash.HashS(map[string]interface{}{})
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
		nil,
		mockOperationIDProvider, spy.NewLogDummy())

	// WHEN
	_, err := sut.Provision(context.Background(), osbContext{}, fixProvisionRequest())
	// THEN
	assert.Error(t, err)

}

func TestProvisionConflictWhenInstanceIsProvisioned(t *testing.T) {
	// GIVEN
	mockInstanceStorage := &automock.InstanceStorage{}
	mockInstanceStorage.On("Get", fixInstanceID()).Return(fixNewInstance(), nil).Once()
	defer mockInstanceStorage.AssertExpectations(t)

	mockStateGetter := &automock.InstanceStateGetter{}
	defer mockStateGetter.AssertExpectations(t)
	mockOperationStorage := &automock.OperationStorage{}
	defer mockOperationStorage.AssertExpectations(t)

	mockStateGetter.On("IsProvisioned", fixInstanceID()).
		Return(true, nil).Once()

	mockOperationIDProvider := func() (internal.OperationID, error) {
		return fixOperationID(), nil
	}

	sut := NewProvisioner(
		nil,
		mockInstanceStorage,
		mockStateGetter,
		mockOperationStorage,
		mockOperationStorage,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		mockOperationIDProvider,
		spy.NewLogDummy(),
	)

	// WHEN
	_, err := sut.Provision(context.Background(), osbContext{}, fixProvisionRequest())
	// THEN
	assert.Error(t, err)
	assert.Equal(t, err.StatusCode, http.StatusConflict)
}

func TestProvisionConflictWhenInstanceIsBeingProvisioned(t *testing.T) {
	// GIVEN
	mockInstanceStorage := &automock.InstanceStorage{}
	mockInstanceStorage.On("Get", fixInstanceID()).Return(fixNewInstance(), nil).Once()
	defer mockInstanceStorage.AssertExpectations(t)

	mockStateGetter := &automock.InstanceStateGetter{}
	defer mockStateGetter.AssertExpectations(t)
	mockOperationStorage := &automock.OperationStorage{}
	defer mockOperationStorage.AssertExpectations(t)

	mockStateGetter.On("IsProvisioned", fixInstanceID()).
		Return(false, nil).Once()

	mockStateGetter.On("IsProvisioningInProgress", fixInstanceID()).
		Return(internal.OperationID(""), true, nil)

	mockOperationIDProvider := func() (internal.OperationID, error) {
		return fixOperationID(), nil
	}

	sut := NewProvisioner(
		mockInstanceStorage,
		mockInstanceStorage,
		mockStateGetter,
		mockOperationStorage,
		mockOperationStorage,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		mockOperationIDProvider,
		spy.NewLogDummy(),
	)

	// WHEN
	_, err := sut.Provision(context.Background(), osbContext{}, fixProvisionRequest())
	// THEN
	assert.Error(t, err)
	assert.Equal(t, err.StatusCode, http.StatusConflict)
}

func TestDoProvision(t *testing.T) {
	var (
		appNs         = fixNs()
		appName       = fixAppName()
		iID           = fixInstanceID()
		opID          = fixOperationID()
		appID         = fixAppServiceID()
		eventProvider = fixEventProvider()
		displayName   = fixDisplayName()
	)

	type testCase struct {
		name                           string
		givenCanProvisionOutput        access.CanProvisionOutput
		givenCanProvisionError         error
		expectedOpState                internal.OperationState
		expectedOpDesc                 string
		expectedEventActivationCreated bool
		expectedInstanceState          internal.InstanceState
		initialObjs                    []runtime.Object
		expectCreates                  []runtime.Object
		expectUpdates                  []runtime.Object
	}

	for _, tc := range []testCase{
		{
			name:                           "provision fail namespace not found",
			givenCanProvisionOutput:        access.CanProvisionOutput{Allowed: true},
			expectedOpState:                internal.OperationStateFailed,
			expectedOpDesc:                 fmt.Sprintf("provisioning failed while enabling default Knative Broker for namespace: example-namespace on error: namespaces %q not found", appNs),
			expectedEventActivationCreated: true,
			expectedInstanceState:          internal.InstanceStateFailed,
			initialObjs: []runtime.Object{
				bt.NewAppChannel(string(appName)),
			},
			expectCreates: []runtime.Object{
				bt.NewAppSubscription(string(appNs), string(appName), bt.WithSpec(knative.GetDefaultBrokerURI(appNs))),
			},
		},
		{
			name:                           "provision fail channel not found",
			givenCanProvisionOutput:        access.CanProvisionOutput{Allowed: true},
			expectedOpState:                internal.OperationStateFailed,
			expectedOpDesc:                 `provisioning failed while persisting Knative Subscription for application: ec-prod namespace: example-namespace on error: getting the Knative channel for the application [ec-prod]: channels.messaging.knative.dev "" not found`,
			expectedEventActivationCreated: true,
			expectedInstanceState:          internal.InstanceStateFailed,
			initialObjs: []runtime.Object{
				bt.NewAppNamespace(string(appNs), false),
			},
		},
		{
			name:                           "provision success subscription created before",
			givenCanProvisionOutput:        access.CanProvisionOutput{Allowed: true},
			expectedOpState:                internal.OperationStateSucceeded,
			expectedOpDesc:                 internal.OperationDescriptionProvisioningSucceeded,
			expectedEventActivationCreated: true,
			expectedInstanceState:          internal.InstanceStateSucceeded,
			initialObjs: []runtime.Object{
				bt.NewAppChannel(string(appName)),
				bt.NewAppNamespace(string(appNs), false),
				bt.NewAppSubscription(string(appNs), string(appName)),
			},
			expectUpdates: []runtime.Object{
				bt.NewAppSubscription(string(appNs), string(appName), bt.WithSpec(knative.GetDefaultBrokerURI(appNs))),
				bt.NewAppNamespace(string(appNs), true),
			},
		},
		{
			name:                           "provision success no subscription created before",
			givenCanProvisionOutput:        access.CanProvisionOutput{Allowed: true},
			expectedOpState:                internal.OperationStateSucceeded,
			expectedOpDesc:                 internal.OperationDescriptionProvisioningSucceeded,
			expectedEventActivationCreated: true,
			expectedInstanceState:          internal.InstanceStateSucceeded,
			initialObjs: []runtime.Object{
				bt.NewAppChannel(string(appName)),
				bt.NewAppNamespace(string(appNs), false),
			},
			expectCreates: []runtime.Object{
				bt.NewAppSubscription(string(appNs), string(appName), bt.WithSpec(knative.GetDefaultBrokerURI(appNs))),
			},
			expectUpdates: []runtime.Object{
				bt.NewAppNamespace(string(appNs), true),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// GIVEN
			mockInstanceStorage := &automock.InstanceStorage{}
			defer mockInstanceStorage.AssertExpectations(t)
			mockOperationStorage := &automock.OperationStorage{}
			defer mockOperationStorage.AssertExpectations(t)
			mockAccessChecker := &accessAutomock.ProvisionChecker{}
			defer mockAccessChecker.AssertExpectations(t)
			mockServiceInstanceGetter := &automock.ServiceInstanceGetter{}
			defer mockServiceInstanceGetter.AssertExpectations(t)

			mockOperationStorage.On("UpdateStateDesc", iID, opID, tc.expectedOpState, &tc.expectedOpDesc).Return(nil).Once()
			mockAccessChecker.On("CanProvision", fixInstanceID(), fixAppServiceID(), fixNs(), time.Minute).Return(tc.givenCanProvisionOutput, tc.givenCanProvisionError)
			mockInstanceStorage.On("UpdateState", fixInstanceID(), tc.expectedInstanceState).Return(nil).Once()
			if tc.expectedEventActivationCreated {
				mockServiceInstanceGetter.On("GetByNamespaceAndExternalID", string(fixNs()), string(fixInstanceID())).Return(FixServiceInstance(), nil)
			}

			knCli, k8sCli := bt.NewFakeClients(tc.initialObjs...)

			provisioner := NewProvisioner(
				nil,
				nil,
				nil,
				mockOperationStorage,
				mockOperationStorage,
				mockAccessChecker,
				nil,
				mockServiceInstanceGetter,
				fake.NewSimpleClientset().ApplicationconnectorV1alpha1(),
				knative.NewClient(knCli, k8sCli),
				mockInstanceStorage,
				nil,
				spy.NewLogDummy(),
			)

			// WHEN
			provisioner.do(iID, opID, appName, appID, appNs, eventProvider, displayName)

			// THEN
			if tc.expectedEventActivationCreated == true {
				eventActivation, err := provisioner.eaClient.EventActivations(string(fixNs())).Get(string(fixServiceID()), v1.GetOptions{})
				assert.Nil(t, err)
				assert.Equal(t, fixEventActivation(), eventActivation)
			}
			actionsAsserter := bt.NewActionsAsserter(t, knCli, k8sCli)
			actionsAsserter.AssertCreates(t, tc.expectCreates)
			actionsAsserter.AssertUpdates(t, tc.expectUpdates)
			mockOperationStorage.AssertExpectations(t)
		})
	}
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
