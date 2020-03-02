// Code generated by failery v1.0.0. DO NOT EDIT.

package disabled

import context "context"
import gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

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

// ClusterServiceClassActivatedField provides a failing mock function with given fields: ctx, obj, namespace
func (_m *Resolver) ClusterServiceClassActivatedField(ctx context.Context, obj *gqlschema.ClusterServiceClass, namespace *string) (bool, error) {
	var r0 bool
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ClusterServiceClassClusterAssetGroupField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ClusterServiceClassClusterAssetGroupField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.ClusterAssetGroup, error) {
	var r0 *gqlschema.ClusterAssetGroup
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ClusterServiceClassInstancesField provides a failing mock function with given fields: ctx, obj, namespace
func (_m *Resolver) ClusterServiceClassInstancesField(ctx context.Context, obj *gqlschema.ClusterServiceClass, namespace *string) ([]gqlschema.ServiceInstance, error) {
	var r0 []gqlschema.ServiceInstance
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

// ClusterServicePlanClusterAssetGroupField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ClusterServicePlanClusterAssetGroupField(ctx context.Context, obj *gqlschema.ClusterServicePlan) (*gqlschema.ClusterAssetGroup, error) {
	var r0 *gqlschema.ClusterAssetGroup
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

// CreateServiceInstanceMutation provides a failing mock function with given fields: ctx, namespace, params
func (_m *Resolver) CreateServiceInstanceMutation(ctx context.Context, namespace string, params gqlschema.ServiceInstanceCreateInput) (*gqlschema.ServiceInstance, error) {
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

// DeleteServiceInstanceMutation provides a failing mock function with given fields: ctx, name, namespace
func (_m *Resolver) DeleteServiceInstanceMutation(ctx context.Context, name string, namespace string) (*gqlschema.ServiceInstance, error) {
	var r0 *gqlschema.ServiceInstance
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceBindingEventSubscription provides a failing mock function with given fields: ctx, namespace
func (_m *Resolver) ServiceBindingEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.ServiceBindingEvent, error) {
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

// ServiceBindingsToInstanceQuery provides a failing mock function with given fields: ctx, instanceName, namespace
func (_m *Resolver) ServiceBindingsToInstanceQuery(ctx context.Context, instanceName string, namespace string) (*gqlschema.ServiceBindings, error) {
	var r0 *gqlschema.ServiceBindings
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceBrokerEventSubscription provides a failing mock function with given fields: ctx, namespace
func (_m *Resolver) ServiceBrokerEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.ServiceBrokerEvent, error) {
	var r0 <-chan gqlschema.ServiceBrokerEvent
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceBrokerQuery provides a failing mock function with given fields: ctx, name, namespace
func (_m *Resolver) ServiceBrokerQuery(ctx context.Context, name string, namespace string) (*gqlschema.ServiceBroker, error) {
	var r0 *gqlschema.ServiceBroker
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceBrokersQuery provides a failing mock function with given fields: ctx, namespace, first, offset
func (_m *Resolver) ServiceBrokersQuery(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.ServiceBroker, error) {
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

// ServiceClassAssetGroupField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ServiceClassAssetGroupField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.AssetGroup, error) {
	var r0 *gqlschema.AssetGroup
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceClassClusterAssetGroupField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ServiceClassClusterAssetGroupField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.ClusterAssetGroup, error) {
	var r0 *gqlschema.ClusterAssetGroup
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceClassInstancesField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ServiceClassInstancesField(ctx context.Context, obj *gqlschema.ServiceClass) ([]gqlschema.ServiceInstance, error) {
	var r0 []gqlschema.ServiceInstance
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

// ServiceClassQuery provides a failing mock function with given fields: ctx, name, namespace
func (_m *Resolver) ServiceClassQuery(ctx context.Context, name string, namespace string) (*gqlschema.ServiceClass, error) {
	var r0 *gqlschema.ServiceClass
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceClassesQuery provides a failing mock function with given fields: ctx, namespace, first, offset
func (_m *Resolver) ServiceClassesQuery(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.ServiceClass, error) {
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

// ServiceInstanceEventSubscription provides a failing mock function with given fields: ctx, namespace
func (_m *Resolver) ServiceInstanceEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.ServiceInstanceEvent, error) {
	var r0 <-chan gqlschema.ServiceInstanceEvent
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServiceInstanceQuery provides a failing mock function with given fields: ctx, name, namespace
func (_m *Resolver) ServiceInstanceQuery(ctx context.Context, name string, namespace string) (*gqlschema.ServiceInstance, error) {
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

// ServiceInstancesQuery provides a failing mock function with given fields: ctx, namespace, first, offset, status
func (_m *Resolver) ServiceInstancesQuery(ctx context.Context, namespace string, first *int, offset *int, status *gqlschema.InstanceStatusType) ([]gqlschema.ServiceInstance, error) {
	var r0 []gqlschema.ServiceInstance
	var r1 error
	r1 = _m.err

	return r0, r1
}

// ServicePlanClusterAssetGroupField provides a failing mock function with given fields: ctx, obj
func (_m *Resolver) ServicePlanClusterAssetGroupField(ctx context.Context, obj *gqlschema.ServicePlan) (*gqlschema.ClusterAssetGroup, error) {
	var r0 *gqlschema.ClusterAssetGroup
	var r1 error
	r1 = _m.err

	return r0, r1
}
