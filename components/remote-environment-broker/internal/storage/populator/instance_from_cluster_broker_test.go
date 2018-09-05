package populator_test

import (
	"context"
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
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPopulateOnlyREBInstances(t *testing.T) {
	// GIVEN
	mockClientSet := fake.NewSimpleClientset(fixAllSCObjects()...)

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

	mockConverter := &automock.InstanceConverter{}
	defer mockConverter.AssertExpectations(t)

	mockConverter.On("MapServiceInstance", fixRebServiceInstanceFromClusterSCInDefaultNs()).Return(expectedServiceInstanceInDefaultNs())
	mockConverter.On("MapServiceInstance", fixRebServiceInstanceFromClusterSCInProdNs()).Return(expectedServiceInstanceInProdNs())
	sut := populator.NewInstancesFromClusterBroker(mockClientSet, mockInserter, mockConverter, fixREBrokerName())
	// WHEN
	actualErr := sut.Do(context.Background())
	// THEN
	assert.NoError(t, actualErr)
}

func TestPopulateErrorOnInsert(t *testing.T) {
	// GIVEN
	mockClientSet := fake.NewSimpleClientset(fixAllSCObjects()...)
	mockInserter := &automock.InstanceInserter{}
	defer mockInserter.AssertExpectations(t)
	mockInserter.On("Insert", mock.Anything).Return(errors.New("some error"))

	mockConverter := &automock.InstanceConverter{}
	defer mockConverter.AssertExpectations(t)
	mockConverter.On("MapServiceInstance", mock.Anything).Return(&internal.Instance{})
	sut := populator.NewInstancesFromClusterBroker(mockClientSet, mockInserter, mockConverter, fixREBrokerName())
	// WHEN
	actualErr := sut.Do(context.Background())
	// THEN
	assert.EqualError(t, actualErr, "while inserting service instance: some error")

}

func fixREBrokerName() string {
	return "remote-environment-broker"
}

func fixNsREBrokerName() string {
	return "remote-env-broker"
}

func fixHelmBrokerName() string {
	return "helm-broker"
}

func fixNsHelmBrokerName() string {
	return "ns-helm-broker"
}

func fixAllSCObjects() []runtime.Object {
	return []runtime.Object{
		fixRedisClusterServiceClass(),
		fixRebClusterServiceClass(),
		fixNsRedisServiceClass(),
		fixNsRebClusterServiceClass(),

		fixRedisServiceInstanceFromClusterSC(),
		fixRedisServiceInstanceFromNsSC(),
		fixRebServiceInstanceFromClusterSCInDefaultNs(),
		fixRebServiceInstanceFromClusterSCInProdNs(),
		fixRebServiceInstanceFromNsSC(),
	}
}

func fixRedisClusterServiceClass() *scv1beta1.ClusterServiceClass {
	return &scv1beta1.ClusterServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "service-class-redis",
		},
		Spec: scv1beta1.ClusterServiceClassSpec{
			ClusterServiceBrokerName: fixHelmBrokerName(),
		},
	}
}

func fixRebClusterServiceClass() *scv1beta1.ClusterServiceClass {
	return &scv1beta1.ClusterServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "service-class-reb",
		},
		Spec: scv1beta1.ClusterServiceClassSpec{
			ClusterServiceBrokerName: fixREBrokerName(),
		},
	}
}

func fixNsRedisServiceClass() *scv1beta1.ServiceClass {
	return &scv1beta1.ServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns-service-class-redis",
		},
		Spec: scv1beta1.ServiceClassSpec{
			ServiceBrokerName: fixNsHelmBrokerName(),
		},
	}
}

func fixNsRebClusterServiceClass() *scv1beta1.ServiceClass {
	return &scv1beta1.ServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns-service-class-reb",
		},
		Spec: scv1beta1.ServiceClassSpec{
			ServiceBrokerName: fixNsREBrokerName(),
		},
	}
}

func fixRedisServiceInstanceFromClusterSC() *scv1beta1.ServiceInstance {
	return &scv1beta1.ServiceInstance{
		Spec: scv1beta1.ServiceInstanceSpec{
			ClusterServiceClassRef: &scv1beta1.ClusterObjectReference{
				Name: "service-class-redis",
			},
			ExternalID: "redis-external-id-1",
			ClusterServicePlanRef: &scv1beta1.ClusterObjectReference{
				Name: "micro",
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      "redis-instance-1",
		},
	}
}

func fixRedisServiceInstanceFromNsSC() *scv1beta1.ServiceInstance {
	return &scv1beta1.ServiceInstance{
		Spec: scv1beta1.ServiceInstanceSpec{
			ServiceClassRef: &scv1beta1.LocalObjectReference{
				Name: "ns-service-class-redis",
			},
			ExternalID: "redis-external-id-2",
			ServicePlanRef: &scv1beta1.LocalObjectReference{
				Name: "micro",
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      "redis-instance-2",
		},
	}
}
func fixRebServiceInstanceFromClusterSCInDefaultNs() *scv1beta1.ServiceInstance {
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

func fixRebServiceInstanceFromClusterSCInProdNs() *scv1beta1.ServiceInstance {
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

func fixRebServiceInstanceFromNsSC() *scv1beta1.ServiceInstance {
	return &scv1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "stage",
			Name:      "promotions-instance-3",
		},
		Spec: scv1beta1.ServiceInstanceSpec{
			ServiceClassRef: &scv1beta1.LocalObjectReference{
				Name: "ns-service-class-reb",
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

func expectedServiceInstanceFromNsBroker() *internal.Instance {
	return &internal.Instance{
		ID:            internal.InstanceID("promotions-instance-3"),
		Namespace:     internal.Namespace("stage"),
		ParamsHash:    "TODO",
		ServicePlanID: internal.ServicePlanID("mini"),
		ServiceID:     internal.ServiceID("ns-service-class-reb"),
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
