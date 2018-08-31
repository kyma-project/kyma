package servicecatalog

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	contentPretty "github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
)

type classResolver struct {
	classLister        classListGetter
	planLister         planLister
	instanceLister     classInstanceLister
	asyncApiSpecGetter AsyncApiSpecGetter
	apiSpecGetter      ApiSpecGetter
	contentGetter      ContentGetter
	classConverter     gqlClassConverter
	planConverter      gqlPlanConverter
}

func newClassResolver(classLister classListGetter, planLister planLister, instanceLister classInstanceLister, asyncApiSpecGetter AsyncApiSpecGetter, apiSpecGetter ApiSpecGetter, contentGetter ContentGetter) *classResolver {
	return &classResolver{
		classLister:        classLister,
		planLister:         planLister,
		instanceLister:     instanceLister,
		asyncApiSpecGetter: asyncApiSpecGetter,
		apiSpecGetter:      apiSpecGetter,
		contentGetter:      contentGetter,
		classConverter:     &classConverter{},
		planConverter:      &planConverter{},
	}
}

func (r *classResolver) ServiceClassQuery(ctx context.Context, name string) (*gqlschema.ServiceClass, error) {
	serviceClass, err := r.classLister.Find(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s with name %s", pretty.ServiceClass, name))
		return nil, gqlerror.New(err, pretty.ServiceClass, gqlerror.WithName(name))
	}
	if serviceClass == nil {
		return nil, nil
	}

	result, err := r.classConverter.ToGQL(serviceClass)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting to %s type"), pretty.ServiceClass)
		return nil, gqlerror.New(err, pretty.ServiceClass, gqlerror.WithName(name))
	}

	return result, nil
}

func (r *classResolver) ServiceClassesQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.ServiceClass, error) {
	items, err := r.classLister.List(pager.PagingParams{
		First:  first,
		Offset: offset,
	})

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s"), pretty.ServiceClasses)
		return nil, gqlerror.New(err, pretty.ServiceClasses)
	}

	serviceClasses, err := r.classConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s"), pretty.ServiceClasses)
		return nil, gqlerror.New(err, pretty.ServiceClasses)
	}

	return serviceClasses, nil
}

func (r *classResolver) ServiceClassPlansField(ctx context.Context, obj *gqlschema.ServiceClass) ([]gqlschema.ServicePlan, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve %s for class"), pretty.ServiceClass, pretty.ServicePlans)
		return nil, gqlerror.NewInternal()
	}

	items, err := r.planLister.ListForClass(obj.Name)
	if err != nil {
		glog.Error(errors.Wrap(err, "while getting %s"), pretty.ServicePlans)
		return nil, gqlerror.New(err, pretty.ServicePlans)
	}

	convertedPlans, err := r.planConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s"), pretty.ServicePlans)
		return nil, gqlerror.New(err, pretty.ServicePlans)
	}

	return convertedPlans, nil
}

func (r *classResolver) ServiceClassActivatedField(ctx context.Context, obj *gqlschema.ServiceClass) (bool, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("%s cannot be empty in order to resolve activated field", pretty.ServiceClass))
		return false, gqlerror.NewInternal()
	}

	items, err := r.instanceLister.ListForClass(obj.Name, obj.ExternalName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for %s %s", pretty.ServiceInstances, pretty.ServiceClass, obj.Name))
		return false, gqlerror.New(err, pretty.ServiceInstances)
	}

	return len(items) > 0, nil
}

func (r *classResolver) ServiceClassApiSpecField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve apiSpec field"), pretty.ServiceClass)
		return nil, gqlerror.NewInternal()
	}

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

func (r *classResolver) ServiceClassAsyncApiSpecField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve asyncApiSpec field"), pretty.ServiceClass)
		return nil, gqlerror.NewInternal()
	}

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

func (r *classResolver) ServiceClassContentField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve `content` field"), pretty.ServiceClass)
		return nil, gqlerror.NewInternal()
	}

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

func (r *classResolver) getContentId(obj *gqlschema.ServiceClass) string {
	return fmt.Sprintf("service-class/%s", obj.Name)
}
