// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	handlers "github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/nats/core"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

// JetStreamBackend is an autogenerated mock type for the JetStreamBackend type
type JetStreamBackend struct {
	mock.Mock
}

// DeleteSubscription provides a mock function with given fields: subscription
func (_m *JetStreamBackend) DeleteSubscription(subscription *v1alpha1.Subscription) error {
	ret := _m.Called(subscription)

	var r0 error
	if rf, ok := ret.Get(0).(func(*v1alpha1.Subscription) error); ok {
		r0 = rf(subscription)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetJetStreamSubjects provides a mock function with given fields: subjects
func (_m *JetStreamBackend) GetJetStreamSubjects(subjects []string) []string {
	ret := _m.Called(subjects)

	var r0 []string
	if rf, ok := ret.Get(0).(func([]string) []string); ok {
		r0 = rf(subjects)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// Initialize provides a mock function with given fields: connCloseHandler
func (_m *JetStreamBackend) Initialize(connCloseHandler handlers.ConnClosedHandler) error {
	ret := _m.Called(connCloseHandler)

	var r0 error
	if rf, ok := ret.Get(0).(func(handlers.ConnClosedHandler) error); ok {
		r0 = rf(connCloseHandler)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SyncSubscription provides a mock function with given fields: subscription
func (_m *JetStreamBackend) SyncSubscription(subscription *v1alpha1.Subscription) error {
	ret := _m.Called(subscription)

	var r0 error
	if rf, ok := ret.Get(0).(func(*v1alpha1.Subscription) error); ok {
		r0 = rf(subscription)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
