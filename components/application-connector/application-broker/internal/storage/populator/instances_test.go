package populator_test

import (
	"errors"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage/populator"
	"testing"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage/populator/automock"

	scv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	instanceID              = "ns-service-class-ab"
	instanceNamespace       = "stage"
	instanceApplicationID   = "ns-service-class-ab"
	instanceApplicationPlan = "mini"
)

func TestPopulateInstances(t *testing.T) {
	type tcOperations struct {
		operationID    internal.OperationID
		operationState internal.OperationState
		operationType  internal.OperationType
	}
	for name, tc := range map[string]struct {
		operations                 []tcOperations
		instanceLastConditionState string
		instanceStatus             scv1beta1.ConditionStatus
		instanceCurrentOperation   scv1beta1.ServiceInstanceOperation
		apiPackagesSupport         bool
		executeProvisionProcess    bool
		executeInstanceInsert      bool
		executeNewOperationID      bool
		executeDeprovisionProcess  bool
		executeMapServiceInstance  bool
	}{
		"ready ServiceInstance": {
			operations: []tcOperations{
				{
					operationID:    "ABCD1234",
					operationState: internal.OperationStateSucceeded,
					operationType:  internal.OperationTypeCreate,
				},
			},
			instanceStatus:            scv1beta1.ConditionTrue,
			instanceCurrentOperation:  scv1beta1.ServiceInstanceOperationProvision,
			executeNewOperationID:     true,
			executeInstanceInsert:     true,
			executeMapServiceInstance: true,
		},
		"provision ServiceInstance with PollingLastOperation error": {
			operations: []tcOperations{
				{
					operationID:    "ABCD1234",
					operationState: internal.OperationStateInProgress,
					operationType:  internal.OperationTypeCreate,
				},
			},
			instanceLastConditionState: "ErrorPollingLastOperation",
			instanceStatus:             scv1beta1.ConditionFalse,
			instanceCurrentOperation:   scv1beta1.ServiceInstanceOperationProvision,
			executeProvisionProcess:    true,
			executeInstanceInsert:      true,
			executeMapServiceInstance:  true,
		},
		"deprovision ServiceInstance with CallFailedMode error": {
			operations: []tcOperations{
				{
					operationID:    "ABCD1234",
					operationState: internal.OperationStateSucceeded,
					operationType:  internal.OperationTypeCreate,
				},
				{
					operationID:    "1234ABCD",
					operationState: internal.OperationStateInProgress,
					operationType:  internal.OperationTypeRemove,
				},
			},
			instanceLastConditionState: "DeprovisionCallFailed",
			instanceStatus:             scv1beta1.ConditionFalse,
			instanceCurrentOperation:   scv1beta1.ServiceInstanceOperationDeprovision,
			apiPackagesSupport:         true,
			executeNewOperationID:      true,
			executeDeprovisionProcess:  true,
			executeMapServiceInstance:  true,
		},
		"deprovision ServiceInstance with PollingLastOperation error": {
			operations: []tcOperations{
				{
					operationID:    "ABCD1234",
					operationState: internal.OperationStateSucceeded,
					operationType:  internal.OperationTypeCreate,
				},
				{
					operationID:    "1234ABCD",
					operationState: internal.OperationStateInProgress,
					operationType:  internal.OperationTypeRemove,
				},
			},
			instanceLastConditionState: "ErrorPollingLastOperation",
			instanceStatus:             scv1beta1.ConditionFalse,
			instanceCurrentOperation:   scv1beta1.ServiceInstanceOperationDeprovision,
			executeNewOperationID:      true,
			executeDeprovisionProcess:  true,
			executeMapServiceInstance:  true,
		},
		"provision failed ServiceInstance": {
			instanceStatus: scv1beta1.ConditionFalse,
		},
	} {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			var (
				lastOperation = "CDEF5678"
				instance      = fixABServiceInstanceFromNsSC(lastOperation, tc.instanceLastConditionState, tc.instanceStatus, tc.instanceCurrentOperation)
			)

			mockClientSet := fake.NewSimpleClientset(fixABServiceClass(), instance)

			mockInserter := &automock.InstanceInserter{}
			defer mockInserter.AssertExpectations(t)

			mockConverter := &automock.InstanceConverter{}
			defer mockConverter.AssertExpectations(t)

			mockOperationInserter := &automock.OperationInserter{}
			defer mockOperationInserter.AssertExpectations(t)

			var insertExpectations siInsertExpectations
			if tc.executeInstanceInsert {
				insertExpectations = newSiInsertExpectations(expectedServiceInstanceFromNsBroker(instance))
			} else {
				insertExpectations = newSiInsertExpectations(nil)
			}
			defer insertExpectations.AssertExpectations(t)

			mockBroker := &automock.BrokerProcesses{}
			defer mockBroker.AssertExpectations(t)

			mockIDSelector := &automock.ApplicationServiceIDSelector{}
			defer mockIDSelector.AssertExpectations(t)

			if tc.executeMapServiceInstance {
				mockConverter.On("MapServiceInstance", instance).Return(expectedServiceInstanceFromNsBroker(instance))
			}

			if tc.executeNewOperationID {
				for _, tco := range tc.operations {
					mockBroker.On("NewOperationID").Return(tco.operationID, nil).Once()
				}
			}

			var expectedApplicationID internal.ApplicationServiceID
			if tc.executeProvisionProcess || tc.executeDeprovisionProcess {
				if tc.apiPackagesSupport {
					expectedApplicationID = instanceApplicationPlan
				} else {
					expectedApplicationID = instanceApplicationID
				}
				mockIDSelector.On("SelectApplicationServiceID", instanceApplicationID, instanceApplicationPlan).Return(expectedApplicationID).Once()
			}

			if tc.executeProvisionProcess {
				mockBroker.On("ProvisionProcess", broker.RestoreProvisionRequest{
					Parameters:           nil,
					InstanceID:           instanceApplicationID,
					OperationID:          internal.OperationID(lastOperation),
					Namespace:            instanceNamespace,
					ApplicationServiceID: expectedApplicationID,
				}).Return(nil).Once()

				tc.operations[0].operationID = internal.OperationID(lastOperation)
			}

			for _, tco := range tc.operations {
				mockOperationInserter.On("Insert",
					&internal.InstanceOperation{
						InstanceID:  instanceID,
						OperationID: tco.operationID,
						Type:        tco.operationType,
						State:       tco.operationState},
				).Return(nil).Once()
			}

			if tc.executeDeprovisionProcess {
				mockBroker.On("DeprovisionProcess", broker.DeprovisionProcessRequest{
					Instance: &internal.Instance{
						ID:            instanceID,
						ServiceID:     instanceApplicationID,
						ServicePlanID: instanceApplicationPlan,
						Namespace:     instanceNamespace,
						State:         internal.InstanceStateFailed,
					},
					OperationID:          tc.operations[len(tc.operations)-1].operationID,
					ApplicationServiceID: expectedApplicationID,
				}).Once()
			}

			if tc.executeInstanceInsert {
				mockInserter.On("Insert", mock.MatchedBy(insertExpectations.OnInsertArgsChecker)).
					Run(func(args mock.Arguments) {
						actualInstance := args.Get(0).(*internal.Instance)
						insertExpectations.ReportInsertingInstance(actualInstance)
					}).
					Return(nil).Once()
			}

			sut := populator.NewInstances(mockClientSet, mockInserter, mockConverter, mockOperationInserter, mockBroker, mockIDSelector, logrus.New())

			// WHEN
			actualErr := sut.Do()

			// THEN
			assert.NoError(t, actualErr)
		})
	}
}

func TestPopulateInstancesErrorOnInsert(t *testing.T) {
	// GIVEN
	mockClientSet := fake.NewSimpleClientset(
		fixABServiceClass(),
		fixABServiceInstanceFromNsSC("", "", scv1beta1.ConditionTrue, ""))

	mockInserter := &automock.InstanceInserter{}
	defer mockInserter.AssertExpectations(t)
	mockInserter.On("Insert", mock.Anything).Return(errors.New("some error"))

	mockOperationInserter := &automock.OperationInserter{}
	defer mockOperationInserter.AssertExpectations(t)

	mockConverter := &automock.InstanceConverter{}
	defer mockConverter.AssertExpectations(t)

	mockBroker := &automock.BrokerProcesses{}
	defer mockBroker.AssertExpectations(t)

	mockIDSelector := &automock.ApplicationServiceIDSelector{}
	defer mockIDSelector.AssertExpectations(t)

	mockConverter.On("MapServiceInstance", mock.Anything).Return(&internal.Instance{})
	sut := populator.NewInstances(mockClientSet, mockInserter, mockConverter, mockOperationInserter, mockBroker, mockIDSelector, logrus.New())

	mockBroker.On("NewOperationID").Return(internal.OperationID(""), nil).Once()
	mockOperationInserter.On("Insert", mock.AnythingOfType("*internal.InstanceOperation")).Return(nil).Once()

	// WHEN
	actualErr := sut.Do()

	// THEN
	assert.EqualError(t, actualErr, "while saving service instance data: while inserting service instance: some error")
}

func fixNsAppBrokerName() string {
	return "application-broker"
}

func fixABServiceClass() *scv1beta1.ServiceClass {
	return &scv1beta1.ServiceClass{
		ObjectMeta: v1.ObjectMeta{
			Name: "ns-service-class-ab",
		},
		Spec: scv1beta1.ServiceClassSpec{
			ServiceBrokerName: fixNsAppBrokerName(),
		},
	}
}

func fixABServiceInstanceFromNsSC(lastOperation, lastConditionState string, status scv1beta1.ConditionStatus, currentOperation scv1beta1.ServiceInstanceOperation) *scv1beta1.ServiceInstance {
	return &scv1beta1.ServiceInstance{
		ObjectMeta: v1.ObjectMeta{
			Namespace: instanceNamespace,
			Name:      "promotions-instance-3",
		},
		Spec: scv1beta1.ServiceInstanceSpec{
			ServiceClassRef: &scv1beta1.LocalObjectReference{
				Name: instanceApplicationID,
			},
			ExternalID: instanceID,
			ServicePlanRef: &scv1beta1.LocalObjectReference{
				Name: instanceApplicationPlan,
			},
		},
		Status: scv1beta1.ServiceInstanceStatus{
			Conditions: []scv1beta1.ServiceInstanceCondition{
				{
					Type:   scv1beta1.ServiceInstanceConditionReady,
					Status: status,
				},
			},
			CurrentOperation:   currentOperation,
			LastConditionState: lastConditionState,
			LastOperation:      &lastOperation,
		},
	}
}

func expectedServiceInstanceFromNsBroker(si *scv1beta1.ServiceInstance) *internal.Instance {
	var state internal.InstanceState
	if si.Status.Conditions[0].Status == scv1beta1.ConditionTrue {
		state = internal.InstanceStateSucceeded
	} else {
		state = internal.InstanceStateFailed
	}

	return &internal.Instance{
		ID:            internal.InstanceID(si.Spec.ExternalID),
		Namespace:     instanceNamespace,
		ServicePlanID: instanceApplicationPlan,
		ServiceID:     instanceApplicationID,
		State:         state,
	}
}

type siInsertExpectations struct {
	insertCount map[internal.Instance]int
}

func newSiInsertExpectations(expectedInstances ...*internal.Instance) siInsertExpectations {
	exp := siInsertExpectations{insertCount: make(map[internal.Instance]int)}
	for _, inst := range expectedInstances {
		if inst == nil {
			continue
		}
		exp.insertCount[*inst] = 0
	}
	return exp
}

func (e *siInsertExpectations) OnInsertArgsChecker(actual *internal.Instance) bool {
	_, ok := e.insertCount[*actual]
	return ok
}

func (e *siInsertExpectations) ReportInsertingInstance(actual *internal.Instance) {
	e.insertCount[*actual] = e.insertCount[*actual] + 1
}

func (e *siInsertExpectations) AssertExpectations(t *testing.T) {
	for k, v := range e.insertCount {
		assert.Equal(t, 1, v, "Incorrect number of inserts for [%+v]", k)
	}
}
