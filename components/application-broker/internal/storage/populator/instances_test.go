package populator_test

import (
	"context"
	"errors"
	"testing"

	scv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage/populator"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage/populator/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPopulateInstances(t *testing.T) {
	// GIVEN
	mockClientSet := fake.NewSimpleClientset(fixAllSCObjects()...)

	mockInserter := &automock.InstanceInserter{}
	defer mockInserter.AssertExpectations(t)

	insertExpectations := newSiInsertExpectations(expectedServiceInstanceFromNsBroker())
	defer insertExpectations.AssertExpectations(t)

	mockInserter.On("Insert", mock.MatchedBy(insertExpectations.OnInsertArgsChecker)).
		Run(func(args mock.Arguments) {
			actualInstance := args.Get(0).(*internal.Instance)
			insertExpectations.ReportInsertingInstance(actualInstance)
		}).
		Return(nil).Once()

	mockConverter := &automock.InstanceConverter{}
	defer mockConverter.AssertExpectations(t)

	mockConverter.On("MapServiceInstance", fixABServiceInstanceFromNsSC()).Return(expectedServiceInstanceFromNsBroker())
	sut := populator.NewInstances(mockClientSet, mockInserter, mockConverter)
	// WHEN
	actualErr := sut.Do(context.Background())
	// THEN
	assert.NoError(t, actualErr)
}

func TestPopulateInstancesErrorOnInsert(t *testing.T) {
	// GIVEN
	mockClientSet := fake.NewSimpleClientset(fixAllSCObjects()...)

	mockInserter := &automock.InstanceInserter{}
	defer mockInserter.AssertExpectations(t)
	mockInserter.On("Insert", mock.Anything).Return(errors.New("some error"))

	mockConverter := &automock.InstanceConverter{}
	defer mockConverter.AssertExpectations(t)

	mockConverter.On("MapServiceInstance", mock.Anything).Return(&internal.Instance{})
	sut := populator.NewInstances(mockClientSet, mockInserter, mockConverter)
	// WHEN
	actualErr := sut.Do(context.Background())
	// THEN
	assert.EqualError(t, actualErr, "while inserting service instance: some error")

}

func fixNsAppBrokerName() string {
	return "application-broker"
}

func fixAllSCObjects() []runtime.Object {
	return []runtime.Object{
		fixABServiceClass(),
		fixABServiceInstanceFromNsSC(),
	}
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

func fixABServiceInstanceFromNsSC() *scv1beta1.ServiceInstance {
	return &scv1beta1.ServiceInstance{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "stage",
			Name:      "promotions-instance-3",
		},
		Spec: scv1beta1.ServiceInstanceSpec{
			ServiceClassRef: &scv1beta1.LocalObjectReference{
				Name: "ns-service-class-ab",
			},
			ExternalID: "promotions-external-id-3",
			ServicePlanRef: &scv1beta1.LocalObjectReference{
				Name: "mini",
			},
		},
		Status: scv1beta1.ServiceInstanceStatus{
			Conditions: []scv1beta1.ServiceInstanceCondition{
				{
					Type:   scv1beta1.ServiceInstanceConditionReady,
					Status: scv1beta1.ConditionTrue,
				},
			},
		},
	}
}

func expectedServiceInstanceFromNsBroker() *internal.Instance {
	return &internal.Instance{
		ID:            internal.InstanceID("promotions-instance-3"),
		Namespace:     internal.Namespace("stage"),
		ParamsHash:    "TODO",
		ServicePlanID: internal.ServicePlanID("mini"),
		ServiceID:     internal.ServiceID("ns-service-class-ab"),
		State:         internal.InstanceStateSucceeded,
	}
}

type siInsertExpectations struct {
	insertCount map[internal.Instance]int
}

func newSiInsertExpectations(expectedInstances ...*internal.Instance) siInsertExpectations {
	exp := siInsertExpectations{insertCount: make(map[internal.Instance]int)}
	for _, inst := range expectedInstances {
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
