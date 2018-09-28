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

//go:generate mockery -name=clusterServiceClassListGetter -output=automock -outpkg=automock -case=underscore
type clusterServiceClassListGetter interface {
	clusterServiceClassGetter
	List(pagingParams pager.PagingParams) ([]*v1beta1.ClusterServiceClass, error)
}

//go:generate mockery -name=gqlClusterServiceClassConverter -output=automock -outpkg=automock -case=underscore
type gqlClusterServiceClassConverter interface {
	ToGQL(in *v1beta1.ClusterServiceClass) (*gqlschema.ClusterServiceClass, error)
	ToGQLs(in []*v1beta1.ClusterServiceClass) ([]gqlschema.ClusterServiceClass, error)
}

//go:generate mockery -name=clusterServicePlanLister  -output=automock -outpkg=automock -case=underscore
type clusterServicePlanLister interface {
	ListForClusterServiceClass(name string) ([]*v1beta1.ClusterServicePlan, error)
}

//go:generate mockery -name=instanceListerByClusterServiceClass -output=automock -outpkg=automock -case=underscore
type instanceListerByClusterServiceClass interface {
	ListForClusterServiceClass(className, externalClassName string) ([]*v1beta1.ServiceInstance, error)
}

type clusterServiceClassResolver struct {
	classLister        clusterServiceClassListGetter
	planLister         clusterServicePlanLister
	instanceLister     instanceListerByClusterServiceClass
	asyncApiSpecGetter AsyncApiSpecGetter
	apiSpecGetter      ApiSpecGetter
	contentGetter      ContentGetter
	classConverter     gqlClusterServiceClassConverter
	planConverter      gqlClusterServicePlanConverter
}

func newClusterServiceClassResolver(classLister clusterServiceClassListGetter, planLister clusterServicePlanLister, instanceLister instanceListerByClusterServiceClass, asyncApiSpecGetter AsyncApiSpecGetter, apiSpecGetter ApiSpecGetter, contentGetter ContentGetter) *clusterServiceClassResolver {
	return &clusterServiceClassResolver{
		classLister:        classLister,
		planLister:         planLister,
		instanceLister:     instanceLister,
		asyncApiSpecGetter: asyncApiSpecGetter,
		apiSpecGetter:      apiSpecGetter,
		contentGetter:      contentGetter,
		classConverter:     &clusterServiceClassConverter{},
		planConverter:      &clusterServicePlanConverter{},
	}
}

func (r *clusterServiceClassResolver) ClusterServiceClassQuery(ctx context.Context, name string) (*gqlschema.ClusterServiceClass, error) {
	serviceClass, err := r.classLister.Find(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s with name %s", pretty.ClusterServiceClass, name))
		return nil, gqlerror.New(err, pretty.ClusterServiceClass, gqlerror.WithName(name))
	}
	if serviceClass == nil {
		return nil, nil
	}

	result, err := r.classConverter.ToGQL(serviceClass)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting to %s type", pretty.ClusterServiceClass))
		return nil, gqlerror.New(err, pretty.ClusterServiceClass, gqlerror.WithName(name))
	}

	return result, nil
}

func (r *clusterServiceClassResolver) ClusterServiceClassesQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.ClusterServiceClass, error) {
	items, err := r.classLister.List(pager.PagingParams{
		First:  first,
		Offset: offset,
	})

	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.ClusterServiceClasses))
		return nil, gqlerror.New(err, pretty.ClusterServiceClasses)
	}

	serviceClasses, err := r.classConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ClusterServiceClasses))
		return nil, gqlerror.New(err, pretty.ClusterServiceClasses)
	}

	return serviceClasses, nil
}

func (r *clusterServiceClassResolver) ClusterServiceClassPlansField(ctx context.Context, obj *gqlschema.ClusterServiceClass) ([]gqlschema.ClusterServicePlan, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve %s for class"), pretty.ClusterServiceClass, pretty.ClusterServicePlans)
		return nil, gqlerror.NewInternal()
	}

	items, err := r.planLister.ListForClusterServiceClass(obj.Name)
	if err != nil {
		glog.Error(errors.Wrap(err, "while getting %s"), pretty.ClusterServicePlans)
		return nil, gqlerror.New(err, pretty.ClusterServicePlans)
	}

	convertedPlans, err := r.planConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ClusterServicePlans))
		return nil, gqlerror.New(err, pretty.ClusterServicePlans)
	}

	return convertedPlans, nil
}

func (r *clusterServiceClassResolver) ClusterServiceClassActivatedField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (bool, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("%s cannot be empty in order to resolve activated field", pretty.ClusterServiceClass))
		return false, gqlerror.NewInternal()
	}

	items, err := r.instanceLister.ListForClusterServiceClass(obj.Name, obj.ExternalName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for %s %s", pretty.ServiceInstances, pretty.ClusterServiceClass, obj.Name))
		return false, gqlerror.New(err, pretty.ServiceInstances)
	}

	return len(items) > 0, nil
}

func (r *clusterServiceClassResolver) ClusterServiceClassApiSpecField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve apiSpec field"), pretty.ClusterServiceClass)
		return nil, gqlerror.NewInternal()
	}

	apiSpec, err := r.apiSpecGetter.Find("service-class", obj.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", contentPretty.ApiSpec, pretty.ClusterServiceClass, obj.ExternalName))
		return nil, gqlerror.New(err, contentPretty.ApiSpec)
	}

	if apiSpec == nil {
		return nil, nil
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(apiSpec.Raw)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s for %s %s", contentPretty.ApiSpec, pretty.ClusterServiceClass, obj.ExternalName))
		return nil, gqlerror.New(err, contentPretty.ApiSpec)
	}

	return &result, nil
}

func (r *clusterServiceClassResolver) ClusterServiceClassAsyncApiSpecField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve asyncApiSpec field"), pretty.ClusterServiceClass)
		return nil, gqlerror.NewInternal()
	}

	asyncApiSpec, err := r.asyncApiSpecGetter.Find("service-class", obj.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", contentPretty.AsyncApiSpec, pretty.ClusterServiceClass, obj.ExternalName))
		return nil, gqlerror.New(err, contentPretty.AsyncApiSpec)
	}

	if asyncApiSpec == nil {
		return nil, nil
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(asyncApiSpec.Raw)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s for %s %s", contentPretty.AsyncApiSpec, pretty.ClusterServiceClass, obj.ExternalName))
		return nil, gqlerror.New(err, contentPretty.AsyncApiSpec)
	}

	return &result, nil
}

func (r *clusterServiceClassResolver) ClusterServiceClassContentField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve `content` field"), pretty.ClusterServiceClass)
		return nil, gqlerror.NewInternal()
	}

	content, err := r.contentGetter.Find("service-class", obj.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", contentPretty.Content, pretty.ClusterServiceClass, obj.ExternalName))
		return nil, gqlerror.New(err, contentPretty.Content)
	}

	if content == nil {
		return nil, nil
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(content.Raw)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s for %s %s", contentPretty.Content, pretty.ClusterServiceClass, obj.ExternalName))
		return nil, gqlerror.New(err, contentPretty.Content)
	}

	return &result, nil
}
