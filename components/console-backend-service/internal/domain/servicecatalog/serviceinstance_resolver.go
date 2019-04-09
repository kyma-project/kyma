package servicecatalog

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/pkg/errors"
)

//go:generate mockery -name=clusterServicePlanGetter -output=automock -outpkg=automock -case=underscore
type clusterServicePlanGetter interface {
	Find(name string) (*v1beta1.ClusterServicePlan, error)
	FindByExternalName(planExternalName, className string) (*v1beta1.ClusterServicePlan, error)
}

//go:generate mockery -name=servicePlanGetter -output=automock -outpkg=automock -case=underscore
type servicePlanGetter interface {
	Find(name, namespace string) (*v1beta1.ServicePlan, error)
	FindByExternalName(planExternalName, className, namespace string) (*v1beta1.ServicePlan, error)
}

type serviceInstanceResolver struct {
	instanceSvc                  serviceInstanceSvc
	clusterServicePlanGetter     clusterServicePlanGetter
	clusterServiceClassGetter    clusterServiceClassGetter
	servicePlanGetter            servicePlanGetter
	serviceClassGetter           serviceClassGetter
	instanceConverter            gqlServiceInstanceConverter
	clusterServiceClassConverter gqlClusterServiceClassConverter
	clusterServicePlanConverter  gqlClusterServicePlanConverter
	serviceClassConverter        gqlServiceClassConverter
	servicePlanConverter         gqlServicePlanConverter
}

func newServiceInstanceResolver(instanceSvc serviceInstanceSvc, clusterServicePlanGetter clusterServicePlanGetter, clusterServiceClassGetter clusterServiceClassGetter, servicePlanGetter servicePlanGetter, serviceClassGetter serviceClassGetter) *serviceInstanceResolver {
	return &serviceInstanceResolver{
		instanceSvc:                  instanceSvc,
		clusterServicePlanGetter:     clusterServicePlanGetter,
		clusterServiceClassGetter:    clusterServiceClassGetter,
		servicePlanGetter:            servicePlanGetter,
		serviceClassGetter:           serviceClassGetter,
		instanceConverter:            &serviceInstanceConverter{},
		clusterServiceClassConverter: &clusterServiceClassConverter{},
		clusterServicePlanConverter:  &clusterServicePlanConverter{},
		serviceClassConverter:        &serviceClassConverter{},
		servicePlanConverter:         &servicePlanConverter{},
	}
}

func (r *serviceInstanceResolver) CreateServiceInstanceMutation(ctx context.Context, namespace string, params gqlschema.ServiceInstanceCreateInput) (*gqlschema.ServiceInstance, error) {
	parameters := r.instanceConverter.GQLCreateInputToInstanceCreateParameters(&params, namespace)
	item, err := r.instanceSvc.Create(*parameters)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s` in namespace `%s`", pretty.ServiceInstance, params.Name, namespace))
		return nil, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(params.Name), gqlerror.WithNamespace(namespace))
	}

	// ServicePlan and ServiceClass references are empty just after the resource has been created
	// Adding these references manually, because they are needed to resolve all Service Instance fields

	var clusterServiceClass *v1beta1.ClusterServiceClass
	var serviceClass *v1beta1.ServiceClass

	if parameters.ClassRef.ClusterWide {
		clusterServiceClass, err = r.clusterServiceClassGetter.FindByExternalName(parameters.ClassRef.ExternalName)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while getting %s for externalName `%s`", pretty.ClusterServiceClass, parameters.ClassRef.ExternalName))
			return nil, gqlerror.New(err, pretty.ClusterServiceClass, gqlerror.WithCustomArgument("externalName", parameters.ClassRef.ExternalName))
		}
		if clusterServiceClass == nil {
			glog.Error(fmt.Errorf("cannot find %s with externalName `%s`", pretty.ClusterServiceClass, parameters.ClassRef.ExternalName))
			return nil, gqlerror.NewNotFound(pretty.ClusterServiceClass, gqlerror.WithCustomArgument("externalName", parameters.ClassRef.ExternalName))
		}
	} else {
		serviceClass, err = r.serviceClassGetter.FindByExternalName(parameters.ClassRef.ExternalName, parameters.Namespace)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while getting %s for externalName `%s`", pretty.ServiceClass, parameters.ClassRef.ExternalName))
			return nil, gqlerror.New(err, pretty.ServiceClass, gqlerror.WithCustomArgument("externalName", parameters.ClassRef.ExternalName))
		}
		if serviceClass == nil {
			glog.Error(fmt.Errorf("cannot find %s with externalName `%s`", pretty.ServiceClass, parameters.ClassRef.ExternalName))
			return nil, gqlerror.NewNotFound(pretty.ServiceClass, gqlerror.WithCustomArgument("externalName", parameters.ClassRef.ExternalName))
		}
	}

	var clusterServicePlan *v1beta1.ClusterServicePlan
	var servicePlan *v1beta1.ServicePlan

	if parameters.PlanRef.ClusterWide {
		clusterServicePlan, err = r.clusterServicePlanGetter.FindByExternalName(parameters.PlanRef.ExternalName, clusterServiceClass.Name)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while getting %s for externalName `%s` for %s `%s`", pretty.ClusterServicePlan, parameters.PlanRef.ExternalName, pretty.ClusterServiceClass, clusterServiceClass.Name))
			return nil, gqlerror.New(err, pretty.ClusterServicePlan, gqlerror.WithCustomArgument("externalName", parameters.PlanRef.ExternalName))
		}
		if clusterServicePlan == nil {
			glog.Error(fmt.Errorf("cannot find %s with externalName `%s`", pretty.ClusterServicePlan, parameters.PlanRef.ExternalName))
			return nil, gqlerror.NewNotFound(pretty.ClusterServicePlan, gqlerror.WithCustomArgument("externalName", parameters.PlanRef.ExternalName))
		}
	} else {
		servicePlan, err = r.servicePlanGetter.FindByExternalName(parameters.PlanRef.ExternalName, serviceClass.Name, parameters.Namespace)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while getting %s for externalName `%s` for %s `%s`", pretty.ServicePlan, parameters.PlanRef.ExternalName, pretty.ServiceClass, serviceClass.Name))
			return nil, gqlerror.New(err, pretty.ServicePlan, gqlerror.WithCustomArgument("externalName", parameters.PlanRef.ExternalName))
		}
		if servicePlan == nil {
			glog.Error(fmt.Errorf("cannot find %s with externalName `%s`", pretty.ServicePlan, parameters.PlanRef.ExternalName))
			return nil, gqlerror.NewNotFound(pretty.ServicePlan, gqlerror.WithCustomArgument("externalName", parameters.PlanRef.ExternalName))
		}
	}

	instanceCopy := item.DeepCopy()

	if clusterServiceClass != nil {
		instanceCopy.Spec.ClusterServiceClassRef = &v1beta1.ClusterObjectReference{
			Name: clusterServiceClass.Name,
		}
	} else if serviceClass != nil {
		instanceCopy.Spec.ServiceClassRef = &v1beta1.LocalObjectReference{
			Name: serviceClass.Name,
		}
	}

	if clusterServicePlan != nil {
		instanceCopy.Spec.ClusterServicePlanRef = &v1beta1.ClusterObjectReference{
			Name: clusterServicePlan.Name,
		}
	} else if servicePlan != nil {
		instanceCopy.Spec.ServicePlanRef = &v1beta1.LocalObjectReference{
			Name: servicePlan.Name,
		}
	}

	instance, err := r.instanceConverter.ToGQL(instanceCopy)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServiceInstance))
		return nil, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(params.Name), gqlerror.WithNamespace(namespace))
	}

	return instance, nil
}

func (r *serviceInstanceResolver) DeleteServiceInstanceMutation(ctx context.Context, name, namespace string) (*gqlschema.ServiceInstance, error) {
	instance, err := r.instanceSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while finding %s `%s` in namespace `%s`", pretty.ServiceInstance, name, namespace))
		return nil, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	if instance == nil {
		glog.Error(fmt.Errorf("cannot find %s `%s` in namespace `%s`", pretty.ServiceInstance, name, namespace))
		return nil, gqlerror.NewNotFound(pretty.ServiceInstance, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	instanceCopy := instance.DeepCopy()
	err = r.instanceSvc.Delete(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s` from namespace `%s`", pretty.ServiceInstance, name, namespace))
		return nil, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	deletedInstance, err := r.instanceConverter.ToGQL(instanceCopy)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServiceInstance))
		return nil, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return deletedInstance, nil
}

func (r *serviceInstanceResolver) ServiceInstanceQuery(ctx context.Context, name string, namespace string) (*gqlschema.ServiceInstance, error) {
	serviceInstance, err := r.instanceSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s `%s` from namespace `%s`", pretty.ServiceInstance, name, namespace))
		return nil, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}
	if serviceInstance == nil {
		return nil, nil
	}

	result, err := r.instanceConverter.ToGQL(serviceInstance)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServiceInstance))
		return nil, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return result, nil
}

func (r *serviceInstanceResolver) ServiceInstancesQuery(ctx context.Context, namespace string, first *int, offset *int, status *gqlschema.InstanceStatusType) ([]gqlschema.ServiceInstance, error) {
	var items []*v1beta1.ServiceInstance
	var err error

	if status != nil {
		statusType := r.instanceConverter.GQLStatusTypeToServiceStatusType(*status)
		items, err = r.instanceSvc.ListForStatus(namespace, pager.PagingParams{
			First:  first,
			Offset: offset,
		}, &statusType)
	} else {
		items, err = r.instanceSvc.List(namespace, pager.PagingParams{
			First:  first,
			Offset: offset,
		})
	}

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s for namespace %s", pretty.ServiceInstances, namespace))
		return nil, gqlerror.New(err, pretty.ServiceInstances, gqlerror.WithNamespace(namespace))
	}

	serviceInstances, err := r.instanceConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServiceInstances))
		return nil, gqlerror.New(err, pretty.ServiceInstances, gqlerror.WithNamespace(namespace))
	}

	return serviceInstances, nil
}

func (r *serviceInstanceResolver) ServiceInstanceEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.ServiceInstanceEvent, error) {
	channel := make(chan gqlschema.ServiceInstanceEvent, 1)
	filter := func(instance *v1beta1.ServiceInstance) bool {
		return instance != nil && instance.Namespace == namespace
	}

	instanceListener := listener.NewInstance(channel, filter, r.instanceConverter)

	r.instanceSvc.Subscribe(instanceListener)
	go func() {
		defer close(channel)
		defer r.instanceSvc.Unsubscribe(instanceListener)
		<-ctx.Done()
	}()

	return channel, nil
}

func (r *serviceInstanceResolver) ServiceInstanceClusterServicePlanField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ClusterServicePlan, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("%s cannot be empty in order to resolve %s for instance", pretty.ServiceInstance, pretty.ClusterServicePlan))
		return nil, gqlerror.NewInternal()
	}

	if obj.PlanReference == nil {
		glog.Warning(fmt.Sprintf("PlanReference is empty during resolving %s for instance", pretty.ClusterServicePlan))
		return nil, nil
	}

	if !obj.PlanReference.ClusterWide {
		return nil, nil
	}

	item, err := r.clusterServicePlanGetter.Find(obj.PlanReference.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for %s `%s` in namespace `%s`", pretty.ClusterServicePlan, pretty.ServiceInstance, obj.Name, obj.Namespace))
		return nil, gqlerror.New(err, pretty.ClusterServicePlan, gqlerror.WithName(obj.PlanReference.Name))
	}
	if item == nil {
		return nil, nil
	}

	plan, err := r.clusterServicePlanConverter.ToGQL(item)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s %s to %s type", pretty.ClusterServicePlan, plan.Name, pretty.ClusterServicePlan))
		return nil, gqlerror.New(err, pretty.ClusterServicePlan, gqlerror.WithName(obj.PlanReference.Name))
	}

	return plan, nil
}

func (r *serviceInstanceResolver) ServiceInstanceClusterServiceClassField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ClusterServiceClass, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve %s for instance"), pretty.ServiceInstance, pretty.ClusterServiceClass)
		return nil, gqlerror.NewInternal()
	}

	if obj.ClassReference == nil {
		glog.Warning(fmt.Sprintf("ClassReference is empty during resolving %s for instance", pretty.ClusterServiceClass))
		return nil, nil
	}

	if !obj.ClassReference.ClusterWide {
		return nil, nil
	}

	serviceClass, err := r.clusterServiceClassGetter.Find(obj.ClassReference.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for %s %s in namespace `%s`", pretty.ClusterServiceClass, pretty.ServiceInstance, obj.Name, obj.Namespace))
		return nil, gqlerror.New(err, pretty.ClusterServiceClass, gqlerror.WithName(obj.ClassReference.Name))
	}
	if serviceClass == nil {
		return nil, nil
	}

	result, err := r.clusterServiceClassConverter.ToGQL(serviceClass)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s %s to %s type", pretty.ClusterServiceClass, serviceClass.Name, pretty.ClusterServiceClass))
		return nil, gqlerror.New(err, pretty.ClusterServiceClass, gqlerror.WithName(obj.ClassReference.Name))
	}

	return result, nil
}

func (r *serviceInstanceResolver) ServiceInstanceServicePlanField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServicePlan, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("%s cannot be empty in order to resolve %s for instance", pretty.ServiceInstance, pretty.ServicePlan))
		return nil, gqlerror.NewInternal()
	}

	if obj.PlanReference == nil {
		glog.Warning(fmt.Sprintf("PlanReference is empty during resolving %s for instance", pretty.ServicePlan))
		return nil, nil
	}

	if obj.PlanReference.ClusterWide {
		return nil, nil
	}

	item, err := r.servicePlanGetter.Find(obj.PlanReference.Name, obj.Namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for %s `%s` in namespace `%s`", pretty.ServicePlan, pretty.ServiceInstance, obj.Name, obj.Namespace))
		return nil, gqlerror.New(err, pretty.ServicePlan, gqlerror.WithName(obj.PlanReference.Name))
	}
	if item == nil {
		return nil, nil
	}

	plan, err := r.servicePlanConverter.ToGQL(item)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s %s to %s type", pretty.ServicePlan, plan.Name, pretty.ServicePlan))
		return nil, gqlerror.New(err, pretty.ServicePlan, gqlerror.WithName(obj.PlanReference.Name))
	}

	return plan, nil
}

func (r *serviceInstanceResolver) ServiceInstanceServiceClassField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServiceClass, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve %s for instance"), pretty.ServiceInstance, pretty.ServiceClass)
		return nil, gqlerror.NewInternal()
	}

	if obj.ClassReference == nil {
		glog.Warning(fmt.Sprintf("ClassReference is empty during resolving %s for instance", pretty.ServiceClass))
		return nil, nil
	}

	if obj.ClassReference.ClusterWide {
		return nil, nil
	}

	serviceClass, err := r.serviceClassGetter.Find(obj.ClassReference.Name, obj.Namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for %s %s in namespace `%s`", pretty.ServiceClass, pretty.ServiceInstance, obj.Name, obj.Namespace))
		return nil, gqlerror.New(err, pretty.ServiceClass, gqlerror.WithName(obj.ClassReference.Name))
	}
	if serviceClass == nil {
		return nil, nil
	}

	result, err := r.serviceClassConverter.ToGQL(serviceClass)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s %s to %s type", pretty.ServiceClass, serviceClass.Name, pretty.ServiceClass))
		return nil, gqlerror.New(err, pretty.ServiceClass, gqlerror.WithName(obj.ClassReference.Name))
	}

	return result, nil
}

func (r *serviceInstanceResolver) ServiceInstanceBindableField(ctx context.Context, obj *gqlschema.ServiceInstance) (bool, error) {
	if obj == nil {
		glog.Error(errors.New("ServiceInstance cannot be empty in order to resolve `bindable` field for instance"))
		return false, gqlerror.NewInternal()
	}

	if obj.ClassReference == nil {
		glog.Warning("ClassReference is empty during resolving `bindable` field for instance")
		return false, nil
	}

	if obj.PlanReference == nil {
		glog.Warning("PlanReference is empty during resolving `bindable` field for instance")
		return false, nil
	}

	if !obj.ClassReference.ClusterWide {
		serviceClass, err := r.serviceClassGetter.Find(obj.ClassReference.Name, obj.Namespace)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while getting %s for instance %s in namespace `%s` in order to resolve `bindable` field", pretty.ServiceClass, obj.Name, obj.Namespace))
			return false, gqlerror.New(err, pretty.ServiceClass, gqlerror.WithName(obj.ClassReference.Name))
		}

		servicePlan, err := r.servicePlanGetter.Find(obj.PlanReference.Name, obj.Namespace)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while getting %s for instance %s in namespace `%s` in order to resolve `bindable` field", pretty.ServicePlan, obj.Name, obj.Namespace))
			return false, gqlerror.New(err, pretty.ServicePlan, gqlerror.WithName(obj.PlanReference.Name))
		}

		return r.instanceSvc.IsBindableWithLocalRefs(serviceClass, servicePlan), nil
	}

	serviceClass, err := r.clusterServiceClassGetter.Find(obj.ClassReference.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for instance %s in namespace `%s` in order to resolve `bindable` field", pretty.ClusterServiceClass, obj.Name, obj.Namespace))
		return false, gqlerror.New(err, pretty.ClusterServiceClass, gqlerror.WithName(obj.ClassReference.Name))
	}

	servicePlan, err := r.clusterServicePlanGetter.Find(obj.PlanReference.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for instance %s in namespace `%s` in order to resolve `bindable` field", pretty.ClusterServicePlan, obj.Name, obj.Namespace))
		return false, gqlerror.New(err, pretty.ClusterServicePlan, gqlerror.WithName(obj.PlanReference.Name))
	}

	return r.instanceSvc.IsBindableWithClusterRefs(serviceClass, servicePlan), nil
}
