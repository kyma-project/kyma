package servicecatalog

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	contentPretty "github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
)

//go:generate mockery -name=serviceClassListGetter -output=automock -outpkg=automock -case=underscore
type serviceClassListGetter interface {
	serviceClassGetter
	List(environment string, pagingParams pager.PagingParams) ([]*v1beta1.ServiceClass, error)
}

//go:generate mockery -name=gqlServiceClassConverter -output=automock -outpkg=automock -case=underscore
type gqlServiceClassConverter interface {
	ToGQL(in *v1beta1.ServiceClass) (*gqlschema.ServiceClass, error)
	ToGQLs(in []*v1beta1.ServiceClass) ([]gqlschema.ServiceClass, error)
}

//go:generate mockery -name=servicePlanLister  -output=automock -outpkg=automock -case=underscore
type servicePlanLister interface {
	ListForServiceClass(name string, environment string) ([]*v1beta1.ServicePlan, error)
}

//go:generate mockery -name=instanceListerByServiceClass -output=automock -outpkg=automock -case=underscore
type instanceListerByServiceClass interface {
	ListForServiceClass(className, externalClassName string, environment string) ([]*v1beta1.ServiceInstance, error)
}

type serviceClassResolver struct {
	classLister        serviceClassListGetter
	planLister         servicePlanLister
	instanceLister     instanceListerByServiceClass
	asyncApiSpecGetter AsyncApiSpecGetter
	apiSpecGetter      ApiSpecGetter
	contentGetter      ContentGetter
	classConverter     gqlServiceClassConverter
	planConverter      gqlServicePlanConverter
}

func newServiceClassResolver(classLister serviceClassListGetter, planLister servicePlanLister, instanceLister instanceListerByServiceClass, asyncApiSpecGetter AsyncApiSpecGetter, apiSpecGetter ApiSpecGetter, contentGetter ContentGetter) *serviceClassResolver {
	return &serviceClassResolver{
		classLister:        classLister,
		planLister:         planLister,
		instanceLister:     instanceLister,
		asyncApiSpecGetter: asyncApiSpecGetter,
		apiSpecGetter:      apiSpecGetter,
		contentGetter:      contentGetter,
		classConverter:     &serviceClassConverter{},
		planConverter:      &servicePlanConverter{},
	}
}
func (r *serviceClassResolver) ServiceClassQuery(ctx context.Context, name, environment string) (*gqlschema.ServiceClass, error) {
	serviceClass, err := r.classLister.Find(name, environment)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s with name %s", pretty.ServiceClass, name))
		return nil, gqlerror.New(err, pretty.ServiceClass, gqlerror.WithName(name))
	}
	if serviceClass == nil {
		return nil, nil
	}

	result, err := r.classConverter.ToGQL(serviceClass)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting to %s type", pretty.ServiceClass))
		return nil, gqlerror.New(err, pretty.ServiceClass, gqlerror.WithName(name))
	}

	return result, nil
}

func (r *serviceClassResolver) ServiceClassesQuery(ctx context.Context, environment string, first *int, offset *int) ([]gqlschema.ServiceClass, error) {
	items, err := r.classLister.List(environment, pager.PagingParams{
		First:  first,
		Offset: offset,
	})

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.ServiceClasses))
		return nil, gqlerror.New(err, pretty.ServiceClasses)
	}

	serviceClasses, err := r.classConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServiceClasses))
		return nil, gqlerror.New(err, pretty.ServiceClasses)
	}

	return serviceClasses, nil
}

func (r *serviceClassResolver) ServiceClassPlansField(ctx context.Context, obj *gqlschema.ServiceClass) ([]gqlschema.ServicePlan, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve %s for class"), pretty.ServiceClass, pretty.ServicePlans)
		return nil, gqlerror.NewInternal()
	}

	items, err := r.planLister.ListForServiceClass(obj.Name, obj.Environment)
	if err != nil {
		glog.Error(errors.Wrap(err, "while getting %s"), pretty.ServicePlans)
		return nil, gqlerror.New(err, pretty.ServicePlans)
	}

	convertedPlans, err := r.planConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServicePlans))
		return nil, gqlerror.New(err, pretty.ServicePlans)
	}

	return convertedPlans, nil
}

func (r *serviceClassResolver) ServiceClassActivatedField(ctx context.Context, obj *gqlschema.ServiceClass) (bool, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("%s cannot be empty in order to resolve activated field", pretty.ServiceClass))
		return false, gqlerror.NewInternal()
	}

	items, err := r.instanceLister.ListForServiceClass(obj.Name, obj.ExternalName, obj.Environment)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for %s %s", pretty.ServiceInstances, pretty.ServiceClass, obj.Name))
		return false, gqlerror.New(err, pretty.ServiceInstances)
	}

	return len(items) > 0, nil
}

func (r *serviceClassResolver) ServiceClassApiSpecField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve apiSpec field"), pretty.ServiceClass)
		return nil, gqlerror.NewInternal()
	}

	//TODO: Fix getting docs for local ServiceClasses
	apiSpec, err := r.apiSpecGetter.Find("service-class", obj.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", contentPretty.ApiSpec, pretty.ServiceClass, obj.ExternalName))
		return nil, gqlerror.New(err, contentPretty.ApiSpec)
	}

	if apiSpec == nil {
		return nil, nil
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(apiSpec.Raw)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s for %s %s", contentPretty.ApiSpec, pretty.ServiceClass, obj.ExternalName))
		return nil, gqlerror.New(err, contentPretty.ApiSpec)
	}

	return &result, nil
}

func (r *serviceClassResolver) ServiceClassAsyncApiSpecField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve asyncApiSpec field"), pretty.ServiceClass)
		return nil, gqlerror.NewInternal()
	}

	//TODO: Fix getting docs for local ServiceClasses
	asyncApiSpec, err := r.asyncApiSpecGetter.Find("service-class", obj.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", contentPretty.AsyncApiSpec, pretty.ServiceClass, obj.ExternalName))
		return nil, gqlerror.New(err, contentPretty.AsyncApiSpec)
	}

	if asyncApiSpec == nil {
		return nil, nil
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(asyncApiSpec.Raw)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s for %s %s", contentPretty.AsyncApiSpec, pretty.ServiceClass, obj.ExternalName))
		return nil, gqlerror.New(err, contentPretty.AsyncApiSpec)
	}

	return &result, nil
}

func (r *serviceClassResolver) ServiceClassContentField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve `content` field"), pretty.ServiceClass)
		return nil, gqlerror.NewInternal()
	}

	//TODO: Fix getting docs for local ServiceClasses
	content, err := r.contentGetter.Find("service-class", obj.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", contentPretty.Content, pretty.ServiceClass, obj.ExternalName))
		return nil, gqlerror.New(err, contentPretty.Content)
	}

	if content == nil {
		return nil, nil
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(content.Raw)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s for %s %s", contentPretty.Content, pretty.ServiceClass, obj.ExternalName))
		return nil, gqlerror.New(err, contentPretty.Content)
	}

	return &result, nil
}
