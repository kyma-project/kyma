// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	subscribed "github.com/kyma-project/kyma/components/event-service/internal/events/subscribed"
	mock "github.com/stretchr/testify/mock"
)

// EventsClient is an autogenerated mock type for the EventsClient type
type EventsClient struct {
	mock.Mock
}

// GetSubscribedEvents provides a mock function with given fields: appName
func (_m *EventsClient) GetSubscribedEvents(appName string) (subscribed.Events, error) {
	ret := _m.Called(appName)

	var r0 subscribed.Events
	if rf, ok := ret.Get(0).(func(string) subscribed.Events); ok {
		r0 = rf(appName)
	} else {
		r0 = ret.Get(0).(subscribed.Events)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(appName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
