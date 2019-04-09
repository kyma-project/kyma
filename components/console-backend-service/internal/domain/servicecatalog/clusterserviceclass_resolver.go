package servicecatalog

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	cmsPretty "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/pretty"
	contentPretty "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/content/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
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
	ListForClusterServiceClass(className, externalClassName string, namespace *string) ([]*v1beta1.ServiceInstance, error)
}

type clusterServiceClassResolver struct {
	classLister       clusterServiceClassListGetter
	planLister        clusterServicePlanLister
	instanceLister    instanceListerByClusterServiceClass
	contentRetriever  shared.ContentRetriever
	cmsRetriever      shared.CmsRetriever
	classConverter    gqlClusterServiceClassConverter
	instanceConverter gqlServiceInstanceConverter
	planConverter     gqlClusterServicePlanConverter
}

func newClusterServiceClassResolver(classLister clusterServiceClassListGetter, planLister clusterServicePlanLister, instanceLister instanceListerByClusterServiceClass, contentRetriever shared.ContentRetriever, cmsRetriever shared.CmsRetriever) *clusterServiceClassResolver {
	return &clusterServiceClassResolver{
		classLister:       classLister,
		planLister:        planLister,
		instanceLister:    instanceLister,
		contentRetriever:  contentRetriever,
		cmsRetriever:      cmsRetriever,
		classConverter:    &clusterServiceClassConverter{},
		planConverter:     &clusterServicePlanConverter{},
		instanceConverter: &serviceInstanceConverter{},
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

func (r *clusterServiceClassResolver) ClusterServiceClassInstancesField(ctx context.Context, obj *gqlschema.ClusterServiceClass, namespace *string) ([]gqlschema.ServiceInstance, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("%s cannot be empty in order to resolve activated field", pretty.ClusterServiceClass))
		return nil, gqlerror.NewInternal()
	}

	items, err := r.instanceLister.ListForClusterServiceClass(obj.Name, obj.ExternalName, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for %s %s", pretty.ServiceInstances, pretty.ClusterServiceClass, obj.Name))
		return nil, gqlerror.New(err, pretty.ServiceInstances)
	}

	instances, err := r.instanceConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServiceInstance))
		return nil, gqlerror.New(err, pretty.ServiceInstance)
	}

	return instances, nil
}

func (r *clusterServiceClassResolver) ClusterServiceClassActivatedField(ctx context.Context, obj *gqlschema.ClusterServiceClass, namespace *string) (bool, error) {
	instances, err := r.ClusterServiceClassInstancesField(ctx, obj, namespace)
	if err != nil {
		return false, err
	}

	return len(instances) > 0, nil
}

func (r *clusterServiceClassResolver) ClusterServiceClassApiSpecField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve apiSpec field"), pretty.ClusterServiceClass)
		return nil, gqlerror.NewInternal()
	}

	apiSpec, err := r.contentRetriever.ApiSpec().Find("service-class", obj.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}

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

func (r *clusterServiceClassResolver) ClusterServiceClassOpenApiSpecField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve openApiSpec field"), pretty.ClusterServiceClass)
		return nil, gqlerror.NewInternal()
	}

	openApiSpec, err := r.contentRetriever.OpenApiSpec().Find("service-class", obj.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}

		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", contentPretty.OpenApiSpec, pretty.ClusterServiceClass, obj.ExternalName))
		return nil, gqlerror.New(err, contentPretty.OpenApiSpec)
	}

	if openApiSpec == nil {
		return nil, nil
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(openApiSpec.Raw)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s for %s %s", contentPretty.OpenApiSpec, pretty.ClusterServiceClass, obj.ExternalName))
		return nil, gqlerror.New(err, contentPretty.OpenApiSpec)
	}

	return &result, nil
}

func (r *clusterServiceClassResolver) ClusterServiceClassODataSpecField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*string, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve odataSpec field"), pretty.ClusterServiceClass)
		return nil, gqlerror.NewInternal()
	}

	odataSpec, err := r.contentRetriever.ODataSpec().Find("service-class", obj.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}

		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", contentPretty.ODataSpec, pretty.ClusterServiceClass, obj.ExternalName))
		return nil, gqlerror.New(err, contentPretty.ODataSpec)
	}

	if odataSpec == nil || odataSpec.Raw == "" {
		return nil, nil
	}

	return &odataSpec.Raw, nil
}

func (r *clusterServiceClassResolver) ClusterServiceClassAsyncApiSpecField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.JSON, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve asyncApiSpec field"), pretty.ClusterServiceClass)
		return nil, gqlerror.NewInternal()
	}

	asyncApiSpec, err := r.contentRetriever.AsyncApiSpec().Find("service-class", obj.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}

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

	content, err := r.contentRetriever.Content().Find("service-class", obj.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}

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

func (r *clusterServiceClassResolver) ClusterServiceClassClusterDocsTopicField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.ClusterDocsTopic, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve `clusterDocsTopic` field"), pretty.ClusterServiceClass)
		return nil, gqlerror.NewInternal()
	}

	item, err := r.cmsRetriever.ClusterDocsTopic().Find(obj.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", cmsPretty.ClusterDocsTopic, pretty.ClusterServiceClass, obj.Name))
		return nil, gqlerror.New(err, cmsPretty.ClusterDocsTopic)
	}

	clusterDocsTopic, err := r.cmsRetriever.ClusterDocsTopicConverter().ToGQL(item)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", cmsPretty.ClusterDocsTopic))
		return nil, gqlerror.New(err, cmsPretty.ClusterDocsTopic)
	}

	return clusterDocsTopic, nil
}
