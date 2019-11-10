package servicecatalog

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"

	"github.com/golang/glog"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	cmsPretty "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/pretty"
	rafterPretty "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/pretty"
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
	cmsRetriever      shared.CmsRetriever
	rafterRetriever      shared.RafterRetriever
	classConverter    gqlClusterServiceClassConverter
	instanceConverter gqlServiceInstanceConverter
	planConverter     gqlClusterServicePlanConverter
}

func newClusterServiceClassResolver(classLister clusterServiceClassListGetter, planLister clusterServicePlanLister, instanceLister instanceListerByClusterServiceClass, cmsRetriever shared.CmsRetriever, rafterRetriever shared.RafterRetriever) *clusterServiceClassResolver {
	return &clusterServiceClassResolver{
		classLister:       classLister,
		planLister:        planLister,
		instanceLister:    instanceLister,
		cmsRetriever:      cmsRetriever,
		rafterRetriever:   rafterRetriever,
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

func (r *clusterServiceClassResolver) ClusterServiceClassClusterAssetGroupField(ctx context.Context, obj *gqlschema.ClusterServiceClass) (*gqlschema.ClusterAssetGroup, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve `clusterAssetGroup` field"), pretty.ClusterServiceClass)
		return nil, gqlerror.NewInternal()
	}

	item, err := r.rafterRetriever.ClusterAssetGroup().Find(obj.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", rafterPretty.ClusterAssetGroup, pretty.ClusterServiceClass, obj.Name))
		return nil, gqlerror.New(err, rafterPretty.ClusterAssetGroup)
	}

	clusterAssetGroup, err := r.rafterRetriever.ClusterAssetGroupConverter().ToGQL(item)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", rafterPretty.ClusterAssetGroup))
		return nil, gqlerror.New(err, rafterPretty.ClusterAssetGroup)
	}

	return clusterAssetGroup, nil
}
