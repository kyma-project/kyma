// Code generated by failery v1.0.0. DO NOT EDIT.

package disabled

import context "context"
import gqlschema "github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"

// Resolver is an autogenerated failing mock type for the Resolver type
type Resolver struct {
	err error
}

// NewResolver creates a new Resolver type instance
func NewResolver(err error) *Resolver {
	return &Resolver{err: err}
}

// ClusterServiceBrokerEventSubscription provides a failing mock function with given fields: ctx
func (_m *Resolver) ClusterServiceBrokerEventSubscription(ctx context.Context) (<-chan gqlschema.ClusterServiceBrokerEvent, error) {
	var r0 <-chan gqlschema.ClusterServiceBrokerEvent
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ClusterServiceBrokerQuery provides a failing mock function with given fields: ctx, name
func (_m *Resolver) ClusterServiceBrokerQuery(ctx context.Context, name string) (*gqlschema.ClusterServiceBroker, error) {
	var r0 *gqlschema.ClusterServiceBroker
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ClusterServiceBrokersQuery provides a failing mock function with given fields: ctx, first, offset
func (_m *Resolver) ClusterServiceBrokersQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.ClusterServiceBroker, error) {
	var r0 []gqlschema.ClusterServiceBroker
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ClusterServiceClassActivatedField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ClusterServiceClassActivatedField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (bool, error) {
	var r0 bool
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ClusterServiceClassApiSpecField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ClusterServiceClassApiSpecField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error) {
	var r0 *gqlschema.JSON
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ClusterServiceClassAsyncApiSpecField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ClusterServiceClassAsyncApiSpecField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error) {
	var r0 *gqlschema.JSON
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ClusterServiceClassContentField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ClusterServiceClassContentField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error) {
	var r0 *gqlschema.JSON
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ClusterServiceClassPlansField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ClusterServiceClassPlansField(ctx context.Context, obj *gqlschema.ClusterServiceClass) ([]gqlschema.ClusterServicePlan, error) {
	var r0 []gqlschema.ClusterServicePlan
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ClusterServiceClassQuery provides a failing mock function with given fields: ctx, name
func (_m *Resolver) ClusterServiceClassQuery(ctx context.Context, name string) (*gqlschema.ClusterServiceClass, error) {
	var r0 *gqlschema.ClusterServiceClass
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ClusterServiceClassesQuery provides a failing mock function with given fields: ctx, first, offset
func (_m *Resolver) ClusterServiceClassesQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.ClusterServiceClass, error) {
	var r0 []gqlschema.ClusterServiceClass
	var r1 error
	r1 = _m.err

	return r0, r1
}

// CreateServiceBindingMutation provides a failing mock function with given fields: ctx, serviceBindingName, serviceInstanceName, env, parameters
func (_m *Resolver) CreateServiceBindingMutation(ctx context.Context, serviceBindingName *string, serviceInstanceName string, env string, parameters *gqlschema.JSON) (*gqlschema.CreateServiceBindingOutput, error) {
	var r0 *gqlschema.CreateServiceBindingOutput
	var r1 error
	r1 = _m.err

	return r0, r1
}

// CreateServiceBindingUsageMutation provides a failing mock function with given fields: ctx, input
func (_m *Resolver) CreateServiceBindingUsageMutation(ctx context.Context, input *gqlschema.CreateServiceBindingUsageInput) (*gqlschema.ServiceBindingUsage, error) {
	var r0 *gqlschema.ServiceBindingUsage
	var r1 error
	r1 = _m.err

	return r0, r1
}

// CreateServiceInstanceMutation provides a failing mock function with given fields: ctx, params
func (_m *Resolver) CreateServiceInstanceMutation(ctx context.Context, params gqlschema.ServiceInstanceCreateInput) (*gqlschema.ServiceInstance, error) {
	var r0 *gqlschema.ServiceInstance
	var r1 error
	r1 = _m.err

	return r0, r1
}

// DeleteServiceBindingMutation provides a failing mock function with given fields: ctx, serviceBindingName, env
func (_m *Resolver) DeleteServiceBindingMutation(ctx context.Context, serviceBindingName string, env string) (*gqlschema.DeleteServiceBindingOutput, error) {
	var r0 *gqlschema.DeleteServiceBindingOutput
	var r1 error
	r1 = _m.err

	return r0, r1
}

// DeleteServiceBindingUsageMutation provides a failing mock function with given fields: ctx, serviceBindingUsageName, namespace
func (_m *Resolver) DeleteServiceBindingUsageMutation(ctx context.Context, serviceBindingUsageName string, namespace string) (*gqlschema.DeleteServiceBindingUsageOutput, error) {
	var r0 *gqlschema.DeleteServiceBindingUsageOutput
	var r1 error
	r1 = _m.err

	return r0, r1
}

// DeleteServiceInstanceMutation provides a failing mock function with given fields: ctx, name, environment
func (_m *Resolver) DeleteServiceInstanceMutation(ctx context.Context, name string, environment string) (*gqlschema.ServiceInstance, error) {
	var r0 *gqlschema.ServiceInstance
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ListBindableResources provides a failing mock function with given fields: ctx, environment
func (_m *Resolver) ListBindableResources(ctx context.Context, environment string) ([]gqlschema.BindableResourcesOutputItem, error) {
	var r0 []gqlschema.BindableResourcesOutputItem
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ListServiceUsageKindResources provides a failing mock function with given fields: ctx, usageKind, environment
func (_m *Resolver) ListServiceUsageKindResources(ctx context.Context, usageKind string, environment string) ([]gqlschema.UsageKindResource, error) {
	var r0 []gqlschema.UsageKindResource
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ListUsageKinds provides a failing mock function with given fields: ctx, first, offset
func (_m *Resolver) ListUsageKinds(ctx context.Context, first *int, offset *int) ([]gqlschema.UsageKind, error) {
	var r0 []gqlschema.UsageKind
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceBindingEventSubscription provides a failing mock function with given fields: ctx, environment
func (_m *Resolver) ServiceBindingEventSubscription(ctx context.Context, environment string) (<-chan gqlschema.ServiceBindingEvent, error) {
	var r0 <-chan gqlschema.ServiceBindingEvent
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceBindingQuery provides a failing mock function with given fields: ctx, name, env
func (_m *Resolver) ServiceBindingQuery(ctx context.Context, name string, env string) (*gqlschema.ServiceBinding, error) {
	var r0 *gqlschema.ServiceBinding
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceBindingUsageEventSubscription provides a failing mock function with given fields: ctx, environment
func (_m *Resolver) ServiceBindingUsageEventSubscription(ctx context.Context, environment string) (<-chan gqlschema.ServiceBindingUsageEvent, error) {
	var r0 <-chan gqlschema.ServiceBindingUsageEvent
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceBindingUsageQuery provides a failing mock function with given fields: ctx, name, environment
func (_m *Resolver) ServiceBindingUsageQuery(ctx context.Context, name string, environment string) (*gqlschema.ServiceBindingUsage, error) {
	var r0 *gqlschema.ServiceBindingUsage
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceBindingUsagesOfInstanceQuery provides a failing mock function with given fields: ctx, instanceName, env
func (_m *Resolver) ServiceBindingUsagesOfInstanceQuery(ctx context.Context, instanceName string, env string) ([]gqlschema.ServiceBindingUsage, error) {
	var r0 []gqlschema.ServiceBindingUsage
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceBindingsToInstanceQuery provides a failing mock function with given fields: ctx, instanceName, environment
func (_m *Resolver) ServiceBindingsToInstanceQuery(ctx context.Context, instanceName string, environment string) (gqlschema.ServiceBindings, error) {
	var r0 gqlschema.ServiceBindings
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceBrokerEventSubscription provides a failing mock function with given fields: ctx, environment
func (_m *Resolver) ServiceBrokerEventSubscription(ctx context.Context, environment string) (<-chan gqlschema.ServiceBrokerEvent, error) {
	var r0 <-chan gqlschema.ServiceBrokerEvent
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceBrokerQuery provides a failing mock function with given fields: ctx, name, environment
func (_m *Resolver) ServiceBrokerQuery(ctx context.Context, name string, environment string) (*gqlschema.ServiceBroker, error) {
	var r0 *gqlschema.ServiceBroker
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceBrokersQuery provides a failing mock function with given fields: ctx, environment, first, offset
func (_m *Resolver) ServiceBrokersQuery(ctx context.Context, environment string, first *int, offset *int) ([]gqlschema.ServiceBroker, error) {
	var r0 []gqlschema.ServiceBroker
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceClassActivatedField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ServiceClassActivatedField(ctx context.Context, obj *gqlschema.ServiceClass) (bool, error) {
	var r0 bool
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceClassApiSpecField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ServiceClassApiSpecField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	var r0 *gqlschema.JSON
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceClassAsyncApiSpecField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ServiceClassAsyncApiSpecField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	var r0 *gqlschema.JSON
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceClassContentField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ServiceClassContentField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	var r0 *gqlschema.JSON
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceClassPlansField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ServiceClassPlansField(ctx context.Context, obj *gqlschema.ServiceClass) ([]gqlschema.ServicePlan, error) {
	var r0 []gqlschema.ServicePlan
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceClassQuery provides a failing mock function with given fields: ctx, name, environment
func (_m *Resolver) ServiceClassQuery(ctx context.Context, name string, environment string) (*gqlschema.ServiceClass, error) {
	var r0 *gqlschema.ServiceClass
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceClassesQuery provides a failing mock function with given fields: ctx, environment, first, offset
func (_m *Resolver) ServiceClassesQuery(ctx context.Context, environment string, first *int, offset *int) ([]gqlschema.ServiceClass, error) {
	var r0 []gqlschema.ServiceClass
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceInstanceBindableField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ServiceInstanceBindableField(ctx context.Context, obj *gqlschema.ServiceInstance) (bool, error) {
	var r0 bool
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceInstanceClusterServiceClassField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ServiceInstanceClusterServiceClassField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ClusterServiceClass, error) {
	var r0 *gqlschema.ClusterServiceClass
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceInstanceClusterServicePlanField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ServiceInstanceClusterServicePlanField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ClusterServicePlan, error) {
	var r0 *gqlschema.ClusterServicePlan
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceInstanceEventSubscription provides a failing mock function with given fields: ctx, environment
func (_m *Resolver) ServiceInstanceEventSubscription(ctx context.Context, environment string) (<-chan gqlschema.ServiceInstanceEvent, error) {
	var r0 <-chan gqlschema.ServiceInstanceEvent
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceInstanceQuery provides a failing mock function with given fields: ctx, name, environment
func (_m *Resolver) ServiceInstanceQuery(ctx context.Context, name string, environment string) (*gqlschema.ServiceInstance, error) {
	var r0 *gqlschema.ServiceInstance
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceInstanceServiceClassField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ServiceInstanceServiceClassField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServiceClass, error) {
	var r0 *gqlschema.ServiceClass
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceInstanceServicePlanField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ServiceInstanceServicePlanField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServicePlan, error) {
	var r0 *gqlschema.ServicePlan
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceInstancesQuery provides a failing mock function with given fields: ctx, environment, first, offset, status
func (_m *Resolver) ServiceInstancesQuery(ctx context.Context, environment string, first *int, offset *int, status *gqlschema.InstanceStatusType) ([]gqlschema.ServiceInstance, error) {
	var r0 []gqlschema.ServiceInstance
	var r1 error
	r1 = _m.err

	return r0, r1
}
