package servicecatalog

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
)

type instanceResolver struct {
	instanceSvc       instanceSvc
	planGetter        planGetter
	classGetter       classGetter
	instanceConverter gqlInstanceConverter
	classConverter    gqlClassConverter
	planConverter     gqlPlanConverter
}

func newInstanceResolver(instanceSvc instanceSvc, planGetter planGetter, classGetter classGetter) *instanceResolver {
	return &instanceResolver{
		instanceSvc:       instanceSvc,
		planGetter:        planGetter,
		classGetter:       classGetter,
		instanceConverter: &instanceConverter{},
		classConverter:    &classConverter{},
		planConverter:     &planConverter{},
	}
}

func (r *instanceResolver) CreateServiceInstanceMutation(ctx context.Context, params gqlschema.ServiceInstanceCreateInput) (*gqlschema.ServiceInstance, error) {
	parameters := r.instanceConverter.GQLCreateInputToInstanceCreateParameters(&params)
	item, err := r.instanceSvc.Create(*parameters)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s` in environment `%s`", pretty.ServiceInstance, params.Name, params.Environment))
		return nil, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(params.Name), gqlerror.WithEnvironment(params.Environment))
	}

	// ServicePlan and ServiceClass references are empty just after the resource has been created
	// Adding these references manually, because they are needed to resolve all Service Instance fields
	serviceClass, err := r.classGetter.FindByExternalName(parameters.ExternalServiceClassName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for externalName `%s`", pretty.ServiceClass, parameters.ExternalServiceClassName))
		return nil, gqlerror.New(err, pretty.ServiceClass, gqlerror.WithCustomArgument("externalName", parameters.ExternalServiceClassName))
	}
	if serviceClass == nil {
		glog.Error(fmt.Errorf("cannot find %s with externalName `%s`", pretty.ServiceClass, parameters.ExternalServiceClassName))
		return nil, gqlerror.NewNotFound(pretty.ServiceClass, gqlerror.WithCustomArgument("externalName", parameters.ExternalServiceClassName))
	}

	servicePlan, err := r.planGetter.FindByExternalNameForClass(parameters.ExternalServicePlanName, serviceClass.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for externalName `%s` for %s `%s`", pretty.ServicePlan, parameters.ExternalServicePlanName, pretty.ServiceClass, serviceClass.Name))
		return nil, gqlerror.New(err, pretty.ServicePlan, gqlerror.WithCustomArgument("externalName", parameters.ExternalServicePlanName))
	}
	if servicePlan == nil {
		glog.Error(fmt.Errorf("cannot find %s with externalName `%s`", pretty.ServicePlan, parameters.ExternalServicePlanName))
		return nil, gqlerror.NewNotFound(pretty.ServicePlan, gqlerror.WithCustomArgument("externalName", parameters.ExternalServicePlanName))
	}

	instanceCopy := item.DeepCopy()
	instanceCopy.Spec.ClusterServicePlanRef = &v1beta1.ClusterObjectReference{
		Name: servicePlan.Name,
	}
	instanceCopy.Spec.ClusterServiceClassRef = &v1beta1.ClusterObjectReference{
		Name: serviceClass.Name,
	}

	instance, err := r.instanceConverter.ToGQL(instanceCopy)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServiceInstance))
		return nil, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(params.Name), gqlerror.WithEnvironment(params.Environment))
	}

	return instance, nil
}

func (r *instanceResolver) DeleteServiceInstanceMutation(ctx context.Context, name, environment string) (*gqlschema.ServiceInstance, error) {
	instance, err := r.instanceSvc.Find(name, environment)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while finding %s `%s` in environment `%s`", pretty.ServiceInstance, name, environment))
		return nil, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(name), gqlerror.WithEnvironment(environment))
	}

	if instance == nil {
		glog.Error(fmt.Errorf("cannot find %s `%s` in environment `%s`", pretty.ServiceInstance, name, environment))
		return nil, gqlerror.NewNotFound(pretty.ServiceInstance, gqlerror.WithName(name), gqlerror.WithEnvironment(environment))
	}

	instanceCopy := instance.DeepCopy()
	err = r.instanceSvc.Delete(name, environment)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s` from environment `%s`", pretty.ServiceInstance, name, environment))
		return nil, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(name), gqlerror.WithEnvironment(environment))
	}

	deletedInstance, err := r.instanceConverter.ToGQL(instanceCopy)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServiceInstance))
		return nil, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(name), gqlerror.WithEnvironment(environment))
	}

	return deletedInstance, nil
}

func (r *instanceResolver) ServiceInstanceQuery(ctx context.Context, name string, environment string) (*gqlschema.ServiceInstance, error) {
	serviceInstance, err := r.instanceSvc.Find(name, environment)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s `%s` from environment `%s`", pretty.ServiceInstance, name, environment))
		return nil, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(name), gqlerror.WithEnvironment(environment))
	}
	if serviceInstance == nil {
		return nil, nil
	}

	result, err := r.instanceConverter.ToGQL(serviceInstance)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServiceInstance))
		return nil, gqlerror.New(err, pretty.ServiceInstance, gqlerror.WithName(name), gqlerror.WithEnvironment(environment))
	}

	return result, nil
}

func (r *instanceResolver) ServiceInstancesQuery(ctx context.Context, environment string, first *int, offset *int, status *gqlschema.InstanceStatusType) ([]gqlschema.ServiceInstance, error) {
	var items []*v1beta1.ServiceInstance
	var err error

	if status != nil {
		statusType := r.instanceConverter.GQLStatusTypeToServiceStatusType(*status)
		items, err = r.instanceSvc.ListForStatus(environment, pager.PagingParams{
			First:  first,
			Offset: offset,
		}, &statusType)
	} else {
		items, err = r.instanceSvc.List(environment, pager.PagingParams{
			First:  first,
			Offset: offset,
		})
	}

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s for environment %s", pretty.ServiceInstances, environment))
		return nil, gqlerror.New(err, pretty.ServiceInstances, gqlerror.WithEnvironment(environment))
	}

	serviceInstances, err := r.instanceConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServiceInstances))
		return nil, gqlerror.New(err, pretty.ServiceInstances, gqlerror.WithEnvironment(environment))
	}

	return serviceInstances, nil
}

func (r *instanceResolver) ServiceInstanceEventSubscription(ctx context.Context, environment string) (<-chan gqlschema.ServiceInstanceEvent, error) {
	channel := make(chan gqlschema.ServiceInstanceEvent, 1)
	filter := func(object interface{}) bool {
		instance, ok := object.(*v1beta1.ServiceInstance)
		if !ok {
			return false
		}

		return instance.Namespace == environment
	}

	listener := newInstanceListener(channel, filter, r.instanceConverter)

	r.instanceSvc.Subscribe(listener)
	go func() {
		defer close(channel)
		defer r.instanceSvc.Unsubscribe(listener)
		<-ctx.Done()
	}()

	return channel, nil
}

func (r *instanceResolver) ServiceInstanceServicePlanField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServicePlan, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("%s cannot be empty in order to resolve %s for instance", pretty.ServiceInstance, pretty.ServicePlan))
		return nil, gqlerror.NewInternal()
	}
	if obj.ServicePlanName == nil {
		glog.Warning(fmt.Sprintf("ServicePlanName is empty during resolving %s for %s", pretty.ServicePlan, pretty.ServiceInstance))
		return nil, nil
	}

	item, err := r.planGetter.Find(*obj.ServicePlanName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for %s `%s` in environment `%s`", pretty.ServicePlan, pretty.ServiceInstance, obj.Name, obj.Environment))
		return nil, gqlerror.New(err, pretty.ServicePlan, gqlerror.WithName(*obj.ServicePlanName))
	}
	if item == nil {
		return nil, nil
	}

	plan, err := r.planConverter.ToGQL(item)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s %s to %s type", pretty.ServicePlan, plan.Name, pretty.ServicePlan))
		return nil, gqlerror.New(err, pretty.ServicePlan, gqlerror.WithName(*obj.ServicePlanName))
	}

	return plan, nil
}

func (r *instanceResolver) ServiceInstanceServiceClassField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServiceClass, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve %s for instance"), pretty.ServiceInstance, pretty.ServiceClass)
		return nil, gqlerror.NewInternal()
	}
	if obj.ServiceClassName == nil {
		glog.Warning(fmt.Sprintf("ServiceClassName is empty during resolving %s for instance", pretty.ServiceClass))
		return nil, nil
	}

	serviceClass, err := r.classGetter.Find(*obj.ServiceClassName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for %s %s in environment `%s`", pretty.ServiceClass, pretty.ServiceInstance, obj.Name, obj.Environment))
		return nil, gqlerror.New(err, pretty.ServiceClass, gqlerror.WithName(*obj.ServiceClassName))
	}
	if serviceClass == nil {
		return nil, nil
	}

	result, err := r.classConverter.ToGQL(serviceClass)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s %s to %s type", pretty.ServiceClass, serviceClass.Name, pretty.ServiceClass))
		return nil, gqlerror.New(err, pretty.ServiceClass, gqlerror.WithName(*obj.ServiceClassName))
	}

	return result, nil
}

func (r *instanceResolver) ServiceInstanceBindableField(ctx context.Context, obj *gqlschema.ServiceInstance) (bool, error) {
	if obj == nil {
		glog.Error(errors.New("ServiceInstance cannot be empty in order to resolve `bindable` field for instance"))
		return false, gqlerror.NewInternal()
	}
	if obj.ServiceClassName == nil {
		glog.Warning(errors.New("ServiceClassName is empty during resolving `bindable` field for instance"))
		return false, nil
	}
	if obj.ServicePlanName == nil {
		glog.Warning(errors.New("ServicePlanName is empty during resolving `bindable` field for instance"))
		return false, nil
	}

	serviceClass, err := r.classGetter.Find(*obj.ServiceClassName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting ServiceClass for instance %s in environment `%s` in order to resolve `bindable` field", obj.Name, obj.Environment))
		return false, gqlerror.New(err, pretty.ServiceClass, gqlerror.WithName(*obj.ServiceClassName))
	}

	servicePlan, err := r.planGetter.Find(*obj.ServicePlanName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting ServicePlan for instance %s in environment `%s` in order to resolve `bindable` field", obj.Name, obj.Environment))
		return false, gqlerror.New(err, pretty.ServicePlan, gqlerror.WithName(*obj.ServicePlanName))
	}

	return r.instanceSvc.IsBindable(serviceClass, servicePlan), nil
}
