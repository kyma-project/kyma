// Code generated by mockery v1.0.0. DO NOT EDIT.
package automock

import context "context"

import gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
import mock "github.com/stretchr/testify/mock"

// Resolver is an autogenerated mock type for the Resolver type
type Resolver struct {
	mock.Mock
}

// CreateManyTriggers provides a mock function with given fields: ctx, namespace, triggers, ownerRef
func (_m *Resolver) CreateManyTriggers(ctx context.Context, namespace string, triggers []gqlschema.TriggerCreateInput, ownerRef []gqlschema.OwnerReference) ([]gqlschema.Trigger, error) {
	ret := _m.Called(ctx, namespace, triggers, ownerRef)

	var r0 []gqlschema.Trigger
	if rf, ok := ret.Get(0).(func(context.Context, string, []gqlschema.TriggerCreateInput, []gqlschema.OwnerReference) []gqlschema.Trigger); ok {
		r0 = rf(ctx, namespace, triggers, ownerRef)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gqlschema.Trigger)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []gqlschema.TriggerCreateInput, []gqlschema.OwnerReference) error); ok {
		r1 = rf(ctx, namespace, triggers, ownerRef)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateTrigger provides a mock function with given fields: ctx, namespace, trigger, ownerRef
func (_m *Resolver) CreateTrigger(ctx context.Context, namespace string, trigger gqlschema.TriggerCreateInput, ownerRef []gqlschema.OwnerReference) (*gqlschema.Trigger, error) {
	ret := _m.Called(ctx, namespace, trigger, ownerRef)

	var r0 *gqlschema.Trigger
	if rf, ok := ret.Get(0).(func(context.Context, string, gqlschema.TriggerCreateInput, []gqlschema.OwnerReference) *gqlschema.Trigger); ok {
		r0 = rf(ctx, namespace, trigger, ownerRef)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gqlschema.Trigger)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, gqlschema.TriggerCreateInput, []gqlschema.OwnerReference) error); ok {
		r1 = rf(ctx, namespace, trigger, ownerRef)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteManyTriggers provides a mock function with given fields: ctx, namespace, triggers
func (_m *Resolver) DeleteManyTriggers(ctx context.Context, namespace string, triggers []gqlschema.TriggerMetadataInput) ([]gqlschema.TriggerMetadata, error) {
	ret := _m.Called(ctx, namespace, triggers)

	var r0 []gqlschema.TriggerMetadata
	if rf, ok := ret.Get(0).(func(context.Context, string, []gqlschema.TriggerMetadataInput) []gqlschema.TriggerMetadata); ok {
		r0 = rf(ctx, namespace, triggers)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gqlschema.TriggerMetadata)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []gqlschema.TriggerMetadataInput) error); ok {
		r1 = rf(ctx, namespace, triggers)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteTrigger provides a mock function with given fields: ctx, namespace, trigger
func (_m *Resolver) DeleteTrigger(ctx context.Context, namespace string, trigger gqlschema.TriggerMetadataInput) (*gqlschema.TriggerMetadata, error) {
	ret := _m.Called(ctx, namespace, trigger)

	var r0 *gqlschema.TriggerMetadata
	if rf, ok := ret.Get(0).(func(context.Context, string, gqlschema.TriggerMetadataInput) *gqlschema.TriggerMetadata); ok {
		r0 = rf(ctx, namespace, trigger)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gqlschema.TriggerMetadata)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, gqlschema.TriggerMetadataInput) error); ok {
		r1 = rf(ctx, namespace, trigger)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TriggerEventSubscription provides a mock function with given fields: ctx, namespace, subscriber
func (_m *Resolver) TriggerEventSubscription(ctx context.Context, namespace string, subscriber *gqlschema.SubscriberInput) (<-chan gqlschema.TriggerEvent, error) {
	ret := _m.Called(ctx, namespace, subscriber)

	var r0 <-chan gqlschema.TriggerEvent
	if rf, ok := ret.Get(0).(func(context.Context, string, *gqlschema.SubscriberInput) <-chan gqlschema.TriggerEvent); ok {
		r0 = rf(ctx, namespace, subscriber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan gqlschema.TriggerEvent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, *gqlschema.SubscriberInput) error); ok {
		r1 = rf(ctx, namespace, subscriber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TriggersQuery provides a mock function with given fields: ctx, namespace, subscriber
func (_m *Resolver) TriggersQuery(ctx context.Context, namespace string, subscriber *gqlschema.SubscriberInput) ([]gqlschema.Trigger, error) {
	ret := _m.Called(ctx, namespace, subscriber)

	var r0 []gqlschema.Trigger
	if rf, ok := ret.Get(0).(func(context.Context, string, *gqlschema.SubscriberInput) []gqlschema.Trigger); ok {
		r0 = rf(ctx, namespace, subscriber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gqlschema.Trigger)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, *gqlschema.SubscriberInput) error); ok {
		r1 = rf(ctx, namespace, subscriber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
