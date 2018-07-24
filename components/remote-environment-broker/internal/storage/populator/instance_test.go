package populator_test

import (
	"context"
	"fmt"
	"testing"

	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/storage/populator"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/storage/populator/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPopulateOnlyREBInstances(t *testing.T) {
	// GIVEN
	mockClientSet := fake.NewSimpleClientset(
		fixRedisServiceClass(),
		fixRebServiceClass(),
		fixRedisServiceInstance(),
		fixRebServiceInstanceInDefaultNs(),
		fixRebServiceInstnaceInProdNs())

	mockInserter := &automock.InstanceInserter{}
	defer mockInserter.AssertExpectations(t)

	insertExpectations := newSiInsertExpectations(expectedServiceInstanceInProdNs(), expectedServiceInstanceInDefaultNs())
	defer insertExpectations.AssertExpectations(t)

	mockInserter.On("Insert", mock.MatchedBy(insertExpectations.OnInsertArgsChecker)).
		Run(func(args mock.Arguments) {
			actualInstance := args.Get(0).(*internal.Instance)
			insertExpectations.ReportInsertingInstance(actualInstance)
		}).
		Return(nil).Twice()

	sut := populator.NewInstances(mockClientSet, mockInserter, fixREBrokerName())
	// WHEN
	actualErr := sut.Do(context.Background())
	// THEN
	assert.NoError(t, actualErr)
}

func TestPopulateErrorOnInsert(t *testing.T) {
	// GIVEN
	mockClientSet := fake.NewSimpleClientset(fixRebServiceClass(), fixRebServiceInstanceInDefaultNs())
	mockInserter := &automock.InstanceInserter{}
	defer mockInserter.AssertExpectations(t)

	mockInserter.On("Insert", mock.Anything).Return(errors.New("some error"))
	sut := populator.NewInstances(mockClientSet, mockInserter, fixREBrokerName())
	// WHEN
	actualErr := sut.Do(context.Background())
	// THEN
	assert.Error(t, actualErr)
	fmt.Println(actualErr)

}

func fixRedisServiceClass() *scv1beta1.ClusterServiceClass {
	return &scv1beta1.ClusterServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "service-class-redis",
		},
		Spec: scv1beta1.ClusterServiceClassSpec{
			ClusterServiceBrokerName: fixHelmBrokerName(),
		},
	}
}

func fixRebServiceClass() *scv1beta1.ClusterServiceClass {
	return &scv1beta1.ClusterServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "service-class-reb",
		},
		Spec: scv1beta1.ClusterServiceClassSpec{
			ClusterServiceBrokerName: fixREBrokerName(),
		},
	}
}

func fixRedisServiceInstance() *scv1beta1.ServiceInstance {
	return &scv1beta1.ServiceInstance{
		Spec: scv1beta1.ServiceInstanceSpec{
			ClusterServiceClassRef: &scv1beta1.ClusterObjectReference{
				Name: "service-class-redis",
			},
			ExternalID: "redis-external-id",
			ClusterServicePlanRef: &scv1beta1.ClusterObjectReference{
				Name: "micro",
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      "redis-instance",
		},
	}
}

func fixRebServiceInstanceInDefaultNs() *scv1beta1.ServiceInstance {
	return &scv1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "orders-instance-1",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: scv1beta1.ServiceInstanceSpec{
			ClusterServiceClassRef: &scv1beta1.ClusterObjectReference{
				Name: "service-class-reb",
			},
			ExternalID: "orders-external-id-1",
			ClusterServicePlanRef: &scv1beta1.ClusterObjectReference{
				Name: "default",
			},
		},
	}
}

func expectedServiceInstanceInDefaultNs() *internal.Instance {
	return &internal.Instance{
		ID:            internal.InstanceID("orders-external-id-1"),
		Namespace:     internal.Namespace(metav1.NamespaceDefault),
		ParamsHash:    "TODO",
		ServicePlanID: internal.ServicePlanID("default"),
		ServiceID:     internal.ServiceID("service-class-reb"),
		State:         internal.InstanceStateFailed,
	}
}

func fixRebServiceInstnaceInProdNs() *scv1beta1.ServiceInstance {
	return &scv1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "prod",
			Name:      "orders-instance-2",
		},
		Spec: scv1beta1.ServiceInstanceSpec{
			ClusterServiceClassRef: &scv1beta1.ClusterObjectReference{
				Name: "service-class-reb",
			},
			ExternalID: "orders-external-id-2",
			ClusterServicePlanRef: &scv1beta1.ClusterObjectReference{
				Name: "default",
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

func expectedServiceInstanceInProdNs() *internal.Instance {
	return &internal.Instance{
		ID:            internal.InstanceID("orders-external-id-2"),
		Namespace:     internal.Namespace("prod"),
		ParamsHash:    "TODO",
		ServicePlanID: internal.ServicePlanID("default"),
		ServiceID:     internal.ServiceID("service-class-reb"),
		State:         internal.InstanceStateSucceeded,
	}
}

func fixREBrokerName() string {
	return "remote-environment-broker"
}

func fixHelmBrokerName() string {
	return "helm-broker"
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
