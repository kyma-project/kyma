package servicecatalog

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
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
	externalErr := fmt.Errorf("Cannot create instance `%s` in environment `%s`", params.Name, params.Environment)

	parameters := r.instanceConverter.GQLCreateInputToInstanceCreateParameters(&params)
	item, err := r.instanceSvc.Create(*parameters)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating ServiceInstance `%s` in environment `%s`", params.Name, params.Environment))
		return nil, externalErr
	}

	// ServicePlan and ServiceClass references are empty just after the resource has been created
	// Adding these references manually, because they are needed to resolve all Service Instance fields
	serviceClass, err := r.classGetter.FindByExternalName(parameters.ExternalServiceClassName)
	if err != nil || serviceClass == nil {
		glog.Error(errors.Wrapf(err, "while getting ServiceClass for externalName `%s`", parameters.ExternalServiceClassName))
		return nil, externalErr
	}

	servicePlan, err := r.planGetter.FindByExternalNameForClass(parameters.ExternalServicePlanName, serviceClass.Name)
	if err != nil || servicePlan == nil {
		glog.Error(errors.Wrapf(err, "while getting ServicePlan for externalName `%s`", parameters.ExternalServicePlanName))
		return nil, externalErr
	}

	instanceCopy := item.DeepCopy()
	instanceCopy.Spec.ClusterServicePlanRef = &v1beta1.ClusterObjectReference{
		Name: servicePlan.Name,
	}
	instanceCopy.Spec.ClusterServiceClassRef = &v1beta1.ClusterObjectReference{
		Name: serviceClass.Name,
	}

	instance := r.instanceConverter.ToGQL(instanceCopy)
	return instance, nil
}

func (r *instanceResolver) DeleteServiceInstanceMutation(ctx context.Context, name, environment string) (*gqlschema.ServiceInstance, error) {
	externalErr := fmt.Errorf("Cannot delete instance `%s` in environment `%s`", name, environment)

	instance, err := r.instanceSvc.Find(name, environment)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while finding ServiceInstance `%s` in environment `%s`", name, environment))
		return nil, externalErr
	}

	if instance == nil {
		return nil, fmt.Errorf("Cannot find instance `%s` in environment `%s`", name, environment)
	}

	instanceCopy := instance.DeepCopy()
	err = r.instanceSvc.Delete(name, environment)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting ServiceInstance `%s` from environment `%s`", name, environment))
		return nil, externalErr
	}

	deletedInstance := r.instanceConverter.ToGQL(instanceCopy)

	return deletedInstance, nil
}

func (r *instanceResolver) ServiceInstanceQuery(ctx context.Context, name string, environment string) (*gqlschema.ServiceInstance, error) {
	externalErr := fmt.Errorf("Cannot query instance with name `%s` in environment `%s`", name, environment)

	serviceInstance, err := r.instanceSvc.Find(name, environment)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting ServiceInstance `%s` from environment `%s`", name, environment))
		return nil, externalErr
	}
	if serviceInstance == nil {
		return nil, nil
	}

	result := r.instanceConverter.ToGQL(serviceInstance)

	return result, nil
}

func (r *instanceResolver) ServiceInstancesQuery(ctx context.Context, environment string, first *int, offset *int, status *gqlschema.InstanceStatusType) ([]gqlschema.ServiceInstance, error) {
	externalErr := fmt.Errorf("Cannot query instances in environment `%s`", environment)

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
		glog.Error(errors.Wrapf(err, "while listing ServiceInstances for environment %s", environment))
		return nil, externalErr
	}

	serviceInstances := r.instanceConverter.ToGQLs(items)

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
	errMessage := "Cannot query ServicePlan for instance"

	if obj == nil {
		glog.Error(errors.New("ServiceInstance cannot be empty in order to resolve ServicePlan for instance"))
		return nil, errors.New(errMessage)
	}

	if obj.ServicePlanName == nil {
		return nil, nil
	}

	externalErr := fmt.Errorf("%s `%s` in environment `%s`", errMessage, obj.Name, obj.Environment)

	item, err := r.planGetter.Find(*obj.ServicePlanName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting ServicePlan for instance `%s` in environment `%s`", obj.Name, obj.Environment))
		return nil, externalErr
	}
	if item == nil {
		return nil, nil
	}

	plan, err := r.planConverter.ToGQL(item)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting plan %s to ServicePlan type", plan.Name))
		return nil, externalErr
	}

	return plan, nil
}

func (r *instanceResolver) ServiceInstanceServiceClassField(ctx context.Context, obj *gqlschema.ServiceInstance) (*gqlschema.ServiceClass, error) {
	errMessage := "Cannot query ServiceClass for instance"

	if obj == nil || obj.ServiceClassName == nil {
		glog.Error(errors.New("ServiceClassName cannot be empty in order to resolve ServiceClass for instance"))
		return nil, errors.New(errMessage)
	}

	externalErr := fmt.Errorf("%s `%s` in environment `%s`", errMessage, obj.Name, obj.Environment)

	serviceClass, err := r.classGetter.Find(*obj.ServiceClassName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting ServiceClass for  instance %s in environment `%s`", obj.Name, obj.Environment))
		return nil, externalErr
	}
	if serviceClass == nil {
		return nil, nil
	}

	result, err := r.classConverter.ToGQL(serviceClass)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting class %s to ServiceClass type", serviceClass.Name))
		return nil, externalErr
	}

	return result, nil
}

func (r *instanceResolver) ServiceInstanceBindableField(ctx context.Context, obj *gqlschema.ServiceInstance) (bool, error) {
	errMessage := "Cannot query `bindable` field for instance"

	// FIXME: It returns error for Create mutation, because ServiceClassName and ServicePlanName are empty
	if obj == nil || obj.ServiceClassName == nil || obj.ServicePlanName == nil {
		glog.Error(errors.New("ServiceClass or ServiceClassName or ServicePlanName cannot be empty in order to resolve ServiceClass for instance"))
		return false, errors.New(errMessage)
	}

	externalErr := fmt.Errorf("%s `%s` in environment `%s`", errMessage, obj.Name, obj.Environment)

	serviceClass, err := r.classGetter.Find(*obj.ServiceClassName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting ServiceClass for instance %s in environment `%s` in order to resolve `bindable` field", obj.Name, obj.Environment))
		return false, externalErr
	}

	servicePlan, err := r.planGetter.Find(*obj.ServicePlanName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting ServicePlan for instance %s in environment `%s` in order to resolve `bindable` field", obj.Name, obj.Environment))
		return false, externalErr
	}

	return r.instanceSvc.IsBindable(serviceClass, servicePlan), nil
}
