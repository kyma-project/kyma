package controllers

import (
	"testing"

	"github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma.cx/v1alpha1"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newFakeSubWithStatus(name string, status v1alpha1.SubscriptionStatus) *v1alpha1.Subscription {
	return &v1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status:     status,
	}
}

type MockSupervisior struct {
	mock.Mock
}

func (m *MockSupervisior) PoisonPill() {}

func (m *MockSupervisior) IsRunning() bool { return true }

func (m *MockSupervisior) IsNATSConnected() bool { return true }

func (m *MockSupervisior) StartSubscriptionReq(sub *v1alpha1.Subscription) {
	m.Called(sub)
}

func (m *MockSupervisior) StopSubscriptionReq(sub *v1alpha1.Subscription) {
	m.Called(sub)
}

func Test_UpdateFunction(t *testing.T) {
	mockSupervisor := &MockSupervisior{}
	updateFunction := getUpdateFnWithEventActivationCheck(mockSupervisor)

	orgSub := newFakeSubWithStatus("test-update", v1alpha1.SubscriptionStatus{})

	subWithEventActivation := verifyStartSubCalledOnEventActivation(orgSub, mockSupervisor, updateFunction, t)

	verifyStopSubCalledOnEventDeactivation(subWithEventActivation, mockSupervisor, updateFunction, t)

}
func verifyStartSubCalledOnEventActivation(orgSub *v1alpha1.Subscription, mockSupervisor *MockSupervisior, updateFunction func(oldObj, newObj interface{}), t *testing.T) *v1alpha1.Subscription {
	activeStatus := v1alpha1.SubscriptionStatus{
		Conditions: []v1alpha1.SubscriptionCondition{
			{
				Type:   v1alpha1.EventsActivated,
				Status: v1alpha1.ConditionTrue,
			},
		},
	}
	subWithEventsActivation := orgSub.DeepCopy()
	subWithEventsActivation.Status = activeStatus
	mockSupervisor.On("StartSubscriptionReq", subWithEventsActivation).Return()
	updateFunction(orgSub, subWithEventsActivation)
	mockSupervisor.AssertCalled(t, "StartSubscriptionReq", subWithEventsActivation)
	return subWithEventsActivation
}

func verifyStopSubCalledOnEventDeactivation(orgSub *v1alpha1.Subscription, mockSupervisor *MockSupervisior, updateFunction func(oldObj, newObj interface{}), t *testing.T) {
	deactiveStatus := v1alpha1.SubscriptionStatus{
		Conditions: []v1alpha1.SubscriptionCondition{
			{
				Type:   v1alpha1.EventsActivated,
				Status: v1alpha1.ConditionFalse,
			},
		},
	}
	SubWithEventsDeactivation := orgSub.DeepCopy()
	SubWithEventsDeactivation.Status = deactiveStatus
	mockSupervisor.On("StopSubscriptionReq", SubWithEventsDeactivation).Return()
	updateFunction(orgSub, SubWithEventsDeactivation)
	mockSupervisor.AssertCalled(t, "StopSubscriptionReq", SubWithEventsDeactivation)
}
