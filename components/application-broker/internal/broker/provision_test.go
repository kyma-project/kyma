package broker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	fakeistioclientset "istio.io/client-go/pkg/clientset/versioned/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	eventingfake "knative.dev/eventing/pkg/client/clientset/versioned/fake"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	accessAutomock "github.com/kyma-project/kyma/components/application-broker/internal/access/automock"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker/automock"
	bt "github.com/kyma-project/kyma/components/application-broker/internal/broker/testing"
	"github.com/kyma-project/kyma/components/application-broker/internal/knative"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8testing "k8s.io/client-go/testing"

	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger/spy"
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
		expectedAPIPackageCredsCreated bool
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
			expectedAPIPackageCredsCreated: true,
			expectedInstanceState:          internal.InstanceStateSucceeded,
		},
		{
			name:                           "cannot provision",
			givenCanProvisionOutput:        access.CanProvisionOutput{Allowed: false, Reason: "very important reason"},
			expectedOpState:                internal.OperationStateFailed,
			expectedOpDesc:                 "Forbidden provisioning instance [inst-123] for application [name: ec-prod, id: service-id] in namespace: [" + appNs + "]. Reason: [very important reason]",
			expectedEventActivationCreated: false,
			expectedAPIPackageCredsCreated: false,
			expectedInstanceState:          internal.InstanceStateFailed,
		},
		{
			name:                           "error on access checking",
			givenCanProvisionError:         errors.New("some error"),
			expectedOpState:                internal.OperationStateFailed,
			expectedOpDesc:                 "provisioning failed on error: some error",
			expectedEventActivationCreated: false,
			expectedAPIPackageCredsCreated: false,
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
			apiPkgCredsCreatorMock := &automock.APIPackageCredentialsCreator{}

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

			mockAccessChecker.On("CanProvision", fixInstanceID(), fixAppServiceID(), fixNs(), defaultWaitTime).
				Return(tc.givenCanProvisionOutput, tc.givenCanProvisionError)

			mockAppFinder.On("FindOneByServiceID", fixAppServiceID()).
				Return(fixApp(), nil).
				Once()

			mockInstanceStorage.On("UpdateState", fixInstanceID(), tc.expectedInstanceState).
				Return(nil).
				Once()

			apiPkgCredsCreatorMock.On("EnsureAPIPackageCredentials", context.Background(), fixAppID().ApplicationID, string(fixAppServiceID()), string(fixInstanceID()), fixProvisionRequest().Parameters).Return(nil)

			knCli, k8sCli, istioCli, eaCli := bt.NewFakeClients(tc.initialObjs...)

			sut := NewProvisioner(mockInstanceStorage, mockInstanceStorage, mockStateGetter, mockOperationStorage, mockOperationStorage, mockAccessChecker, mockAppFinder, eaCli.ApplicationconnectorV1alpha1(), knative.NewClient(knCli, k8sCli), istioCli, mockInstanceStorage, mockOperationIDProvider, spy.NewLogDummy(), &IDSelector{false}, apiPkgCredsCreatorMock, validateProvisionRequestV2)

			asyncFinished := make(chan struct{}, 0)
			sut.asyncHook = func() {
				asyncFinished <- struct{}{}
			}

			// WHEN
			actResp, err := sut.Provision(context.Background(), osbContext{BrokerNamespace: string(fixNs())}, fixProvisionRequest())

			// THEN
			assert.Nil(t, err)
			assert.NotNil(t, actResp)
			assert.True(t, actResp.Async)
			expOpID := osb.OperationKey(fixOperationID())
			assert.Equal(t, &expOpID, actResp.OperationKey)

			select {
			case <-asyncFinished:
				if tc.expectedEventActivationCreated {
					eventActivation, err := sut.eaClient.EventActivations(string(fixNs())).
						Get(string(fixServiceID()), v1.GetOptions{})
					assert.Nil(t, err)
					assert.Equal(t, fixEventActivation(), eventActivation)
				}
				if tc.expectedAPIPackageCredsCreated {
					apiPkgCredsCreatorMock.AssertExpectations(t)
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

	sut := NewProvisioner(nil, nil, mockStateGetter, nil, nil, nil, nil, nil, nil, nil, nil, nil, spy.NewLogDummy(), nil, nil, validateProvisionRequestV2)
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

	sut := NewProvisioner(nil, nil, mockStateGetter, nil, nil, nil, nil, nil, nil, nil, nil, nil, spy.NewLogDummy(), nil, nil, validateProvisionRequestV2)

	// WHEN
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
			knCli, k8sCli, _, _ := bt.NewFakeClients(initialObjs...)
			return knCli, k8sCli
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
			knCli, k8sCli, _, _ := bt.NewFakeClients(initialObjs...)
			return knCli, k8sCli
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
			knCli, k8sCli, _, _ := bt.NewFakeClients(initialObjs...)
			return knCli, k8sCli
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
			knCli, k8sCli, _, _ := bt.NewFakeClients(initialObjs...)
			return knCli, k8sCli
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
			apiPkgCredsCreatorMock := &automock.APIPackageCredentialsCreator{}
			defer apiPkgCredsCreatorMock.AssertExpectations(t)

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

			apiPkgCredsCreatorMock.On("EnsureAPIPackageCredentials", context.Background(), fixAppID().ApplicationID, string(fixAppServiceID()), string(fixInstanceID()), fixProvisionRequest().Parameters).Return(nil).
				Once()

			knCli, k8sCli := setupMocks(clientset, mockInstanceStorage, mockOperationStorage)
			sut := NewProvisioner(mockInstanceStorage, mockInstanceStorage, mockStateGetter, mockOperationStorage, mockOperationStorage, mockAccessChecker, mockAppFinder, clientset.ApplicationconnectorV1alpha1(), knative.NewClient(knCli, k8sCli), fakeistioclientset.NewSimpleClientset(), mockInstanceStorage, mockOperationIDProvider, spy.NewLogDummy(), &IDSelector{false}, apiPkgCredsCreatorMock, validateProvisionRequestV2)

			asyncFinished := make(chan struct{}, 0)
			sut.asyncHook = func() {
				asyncFinished <- struct{}{}
			}

			// WHEN
			_, err := sut.Provision(context.Background(), osbContext{BrokerNamespace: string(fixNs())}, fixProvisionRequest())
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

func TestProvisionErrorOnCheckingIfProvisioned(t *testing.T) {
	// GIVEN
	mockStateGetter := &automock.InstanceStateGetter{}
	defer mockStateGetter.AssertExpectations(t)
	mockStateGetter.On("IsProvisioned", fixInstanceID()).Return(false, fixError())

	sut := NewProvisioner(nil, nil, mockStateGetter, nil, nil, nil, nil, nil, nil, nil, nil, nil, spy.NewLogDummy(), nil, nil, validateProvisionRequestV2)
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

	sut := NewProvisioner(nil, nil, mockStateGetter, nil, nil, nil, nil, nil, nil, nil, nil, nil, spy.NewLogDummy(), nil, nil, validateProvisionRequestV2)
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
	sut := NewProvisioner(nil, nil, mockStateGetter, nil, nil, nil, nil, nil, nil, nil, nil, mockOperationIDProvider, spy.NewLogDummy(), nil, nil, validateProvisionRequestV2)
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

	sut := NewProvisioner(nil, nil, mockStateGetter, mockOperationStorage, mockOperationStorage, nil, nil, nil, nil, nil, nil, mockOperationIDProvider, spy.NewLogDummy(), nil, nil, validateProvisionRequestV2)

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

	sut := NewProvisioner(mockInstanceStorage, mockInstanceStorage, mockStateGetter, mockOperationStorage, mockOperationStorage, nil, mockAppFinder, nil, nil, nil, nil, mockOperationIDProvider, spy.NewLogDummy(), &IDSelector{false}, nil, validateProvisionRequestV2)

	// WHEN
	_, err := sut.Provision(context.Background(), osbContext{BrokerNamespace: string(fixNs())}, fixProvisionRequest())
	// THEN
	assert.Error(t, err)

}

func TestDoKnativeResourceProvision(t *testing.T) {
	var (
		appNs         = fixNs()
		appName       = fixAppName()
		iID           = fixInstanceID()
		opID          = fixOperationID()
		appID         = fixAppID().ApplicationID
		appSvcID      = fixAppServiceID()
		eventProvider = fixEventProvider()
		apiProvider   = false
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
			expectedOpDesc:                 `provisioning failed while enabling default Knative Broker for namespace: example-namespace on error: namespace not found: namespaces "example-namespace" not found`,
			expectedEventActivationCreated: true,
			expectedInstanceState:          internal.InstanceStateFailed,
			initialObjs: []runtime.Object{
				bt.NewAppChannel(string(appName)),
			},
			expectCreates: []runtime.Object{
				bt.NewAppSubscription(string(appNs), string(appName), bt.WithSpec(t, knative.GetDefaultBrokerURI(appNs))),
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
			expectCreates: []runtime.Object{
				bt.NewIstioPolicy(string(appNs), fmt.Sprintf("%s%s", appNs, policyNameSuffix)),
			},
			expectUpdates: []runtime.Object{
				bt.NewAppSubscription(string(appNs), string(appName), bt.WithSpec(t, knative.GetDefaultBrokerURI(appNs))),
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
				bt.NewAppSubscription(string(appNs), string(appName), bt.WithSpec(t, knative.GetDefaultBrokerURI(appNs))),
				bt.NewIstioPolicy(string(appNs), fmt.Sprintf("%s%s", appNs, policyNameSuffix)),
			},
			expectUpdates: []runtime.Object{
				bt.NewAppNamespace(string(appNs), true),
			},
		},
		{
			name:                           "provision success no istio policy created before",
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
				bt.NewAppSubscription(string(appNs), string(appName), bt.WithSpec(t, knative.GetDefaultBrokerURI(appNs))),
				bt.NewIstioPolicy(string(appNs), fmt.Sprintf("%s%s", appNs, policyNameSuffix)),
			},
			expectUpdates: []runtime.Object{
				bt.NewAppNamespace(string(appNs), true),
			},
		},
		{
			name:                           "provision success an istio policy created before",
			givenCanProvisionOutput:        access.CanProvisionOutput{Allowed: true},
			expectedOpState:                internal.OperationStateSucceeded,
			expectedOpDesc:                 internal.OperationDescriptionProvisioningSucceeded,
			expectedEventActivationCreated: true,
			expectedInstanceState:          internal.InstanceStateSucceeded,
			initialObjs: []runtime.Object{
				bt.NewAppChannel(string(appName)),
				bt.NewAppNamespace(string(appNs), false),
				bt.NewIstioPolicy(string(appNs), fmt.Sprintf("%s%s", appNs, policyNameSuffix)),
			},
			expectCreates: []runtime.Object{
				bt.NewAppSubscription(string(appNs), string(appName), bt.WithSpec(t, knative.GetDefaultBrokerURI(appNs))),
				bt.NewIstioPolicy(string(appNs), fmt.Sprintf("%s%s", appNs, policyNameSuffix)),
			},
			expectUpdates: []runtime.Object{
				bt.NewAppNamespace(string(appNs), true),
				bt.NewIstioPolicy(string(appNs), fmt.Sprintf("%s%s", appNs, policyNameSuffix)),
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

			mockOperationStorage.On("UpdateStateDesc", iID, opID, tc.expectedOpState, &tc.expectedOpDesc).Return(nil).Once()
			mockAccessChecker.On("CanProvision", fixInstanceID(), fixAppServiceID(), fixNs(), time.Minute).Return(tc.givenCanProvisionOutput, tc.givenCanProvisionError)
			mockInstanceStorage.On("UpdateState", fixInstanceID(), tc.expectedInstanceState).Return(nil).Once()

			knCli, k8sCli, istioCli, eaClient := bt.NewFakeClients(tc.initialObjs...)

			provisioner := NewProvisioner(nil, nil, nil, mockOperationStorage, mockOperationStorage, mockAccessChecker, nil, eaClient.ApplicationconnectorV1alpha1(), knative.NewClient(knCli, k8sCli), istioCli, mockInstanceStorage, nil, spy.NewLogDummy(), nil, nil, validateProvisionRequestV2)

			// WHEN
			provisioner.do(fixProvisionRequest().Parameters, iID, opID, appName, appID, appSvcID, appNs, eventProvider, apiProvider, displayName)

			// THEN
			if tc.expectedEventActivationCreated == true {
				eventActivation, err := provisioner.eaClient.EventActivations(string(fixNs())).Get(string(fixServiceID()), v1.GetOptions{})
				assert.Nil(t, err)
				assert.Equal(t, fixEventActivation(), eventActivation)
			}
			actionsAsserter := bt.NewActionsAsserter(t, knCli, k8sCli, istioCli)
			actionsAsserter.AssertCreates(t, tc.expectCreates)
			actionsAsserter.AssertUpdates(t, tc.expectUpdates)
			mockOperationStorage.AssertExpectations(t)
		})
	}
}

func TestDoAPIPackageCredentialsProvision(t *testing.T) {
	var (
		appNs         = fixNs()
		appName       = fixAppName()
		opID          = fixOperationID()
		displayName   = fixDisplayName()
		eventProvider = false

		iID      = fixInstanceID()
		appID    = fixAppID().ApplicationID
		appSvcID = fixAppServiceID()
	)

	type testCase struct {
		givenCanProvisionOutput     access.CanProvisionOutput
		givenCanProvisionError      error
		expectedOpState             internal.OperationState
		expectedOpDesc              string
		expectedInstanceState       internal.InstanceState
		apiProvider                 bool
		setupAPIPkgCredsCreatorMock func(mock *automock.APIPackageCredentialsCreator)
	}
	for tN, tC := range map[string]testCase{
		"operation should succeed if creating credentials ended successfully": {
			apiProvider: true,

			givenCanProvisionOutput: access.CanProvisionOutput{Allowed: true},
			expectedOpState:         internal.OperationStateSucceeded,
			expectedOpDesc:          internal.OperationDescriptionProvisioningSucceeded,
			expectedInstanceState:   internal.InstanceStateSucceeded,
			setupAPIPkgCredsCreatorMock: func(mock *automock.APIPackageCredentialsCreator) {
				mock.On("EnsureAPIPackageCredentials", context.Background(), fixAppID().ApplicationID, string(fixAppServiceID()), string(fixInstanceID()), fixProvisionRequest().Parameters).Return(nil).
					Once()
			},
		},
		"operation should failed if there was an error while creating credentials": {
			apiProvider: true,

			givenCanProvisionOutput: access.CanProvisionOutput{Allowed: true},
			expectedOpState:         internal.OperationStateFailed,
			expectedOpDesc:          `provisioning failed while ensuring API Package credentials: generic error`,
			expectedInstanceState:   internal.InstanceStateFailed,
			setupAPIPkgCredsCreatorMock: func(mock *automock.APIPackageCredentialsCreator) {
				mock.On("EnsureAPIPackageCredentials", context.Background(), fixAppID().ApplicationID, string(fixAppServiceID()), string(fixInstanceID()), fixProvisionRequest().Parameters).Return(errors.New("generic error")).
					Once()
			},
		},
		"operation should succeed if creating credentials was skipped": {
			apiProvider: false, // we are not api provider, so do not create a creds

			givenCanProvisionOutput: access.CanProvisionOutput{Allowed: true},
			expectedOpState:         internal.OperationStateSucceeded,
			expectedOpDesc:          internal.OperationDescriptionProvisioningSucceeded,
			expectedInstanceState:   internal.InstanceStateSucceeded,
			setupAPIPkgCredsCreatorMock: func(mock *automock.APIPackageCredentialsCreator) {
				mock.ExpectedCalls = nil
			},
		},
	} {
		t.Run(tN, func(t *testing.T) {
			// GIVEN
			mockInstanceStorage := &automock.InstanceStorage{}
			mockOperationStorage := &automock.OperationStorage{}
			mockAccessChecker := &accessAutomock.ProvisionChecker{}
			apiPkgCredsCreatorMock := &automock.APIPackageCredentialsCreator{}

			mockOperationStorage.On("UpdateStateDesc", iID, opID, tC.expectedOpState, &tC.expectedOpDesc).Return(nil).Once()
			mockAccessChecker.On("CanProvision", fixInstanceID(), fixAppServiceID(), fixNs(), time.Minute).Return(tC.givenCanProvisionOutput, tC.givenCanProvisionError)
			mockInstanceStorage.On("UpdateState", fixInstanceID(), tC.expectedInstanceState).Return(nil).Once()

			tC.setupAPIPkgCredsCreatorMock(apiPkgCredsCreatorMock)

			provisioner := NewProvisioner(nil, nil, nil, mockOperationStorage, mockOperationStorage, mockAccessChecker, nil, nil, nil, nil, mockInstanceStorage, nil, spy.NewLogDummy(), nil, apiPkgCredsCreatorMock, validateProvisionRequestV2)

			// WHEN
			provisioner.do(fixProvisionRequest().Parameters, iID, opID, appName, appID, appSvcID, appNs, eventProvider, tC.apiProvider, displayName)

			// THEN
			mockInstanceStorage.AssertExpectations(t)
			mockOperationStorage.AssertExpectations(t)
			mockAccessChecker.AssertExpectations(t)
			apiPkgCredsCreatorMock.AssertExpectations(t)
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

func ptrStr(s string) *string {
	return &s
}
