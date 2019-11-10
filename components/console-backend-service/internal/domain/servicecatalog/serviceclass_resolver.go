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

//go:generate mockery -name=serviceClassListGetter -output=automock -outpkg=automock -case=underscore
type serviceClassListGetter interface {
	serviceClassGetter
	List(namespace string, pagingParams pager.PagingParams) ([]*v1beta1.ServiceClass, error)
}

//go:generate mockery -name=gqlServiceClassConverter -output=automock -outpkg=automock -case=underscore
type gqlServiceClassConverter interface {
	ToGQL(in *v1beta1.ServiceClass) (*gqlschema.ServiceClass, error)
	ToGQLs(in []*v1beta1.ServiceClass) ([]gqlschema.ServiceClass, error)
}

//go:generate mockery -name=servicePlanLister  -output=automock -outpkg=automock -case=underscore
type servicePlanLister interface {
	ListForServiceClass(name string, namespace string) ([]*v1beta1.ServicePlan, error)
}

//go:generate mockery -name=instanceListerByServiceClass -output=automock -outpkg=automock -case=underscore
type instanceListerByServiceClass interface {
	ListForServiceClass(className, externalClassName string, namespace string) ([]*v1beta1.ServiceInstance, error)
}

type serviceClassResolver struct {
	classLister       serviceClassListGetter
	planLister        servicePlanLister
	instanceLister    instanceListerByServiceClass
	cmsRetriever      shared.CmsRetriever
	rafterRetriever   shared.RafterRetriever
	classConverter    gqlServiceClassConverter
	instanceConverter gqlServiceInstanceConverter
	planConverter     gqlServicePlanConverter
}

func newServiceClassResolver(classLister serviceClassListGetter, planLister servicePlanLister, instanceLister instanceListerByServiceClass, cmsRetriever shared.CmsRetriever, rafterRetriever shared.RafterRetriever) *serviceClassResolver {
	return &serviceClassResolver{
		classLister:       classLister,
		planLister:        planLister,
		instanceLister:    instanceLister,
		cmsRetriever:      cmsRetriever,
		rafterRetriever:      rafterRetriever,
		classConverter:    &serviceClassConverter{},
		planConverter:     &servicePlanConverter{},
		instanceConverter: &serviceInstanceConverter{},
	}
}
func (r *serviceClassResolver) ServiceClassQuery(ctx context.Context, name, namespace string) (*gqlschema.ServiceClass, error) {
	serviceClass, err := r.classLister.Find(name, namespace)
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

func (r *serviceClassResolver) ServiceClassesQuery(ctx context.Context, namespace string, first *int, offset *int) ([]gqlschema.ServiceClass, error) {
	items, err := r.classLister.List(namespace, pager.PagingParams{
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

	items, err := r.planLister.ListForServiceClass(obj.Name, obj.Namespace)
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

func (r *serviceClassResolver) ServiceClassInstancesField(ctx context.Context, obj *gqlschema.ServiceClass) ([]gqlschema.ServiceInstance, error) {

	if obj == nil {
		glog.Error(fmt.Errorf("%s cannot be empty in order to resolve activated field", pretty.ServiceClass))
		return nil, gqlerror.NewInternal()
	}

	items, err := r.instanceLister.ListForServiceClass(obj.Name, obj.ExternalName, obj.Namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for %s %s", pretty.ServiceInstances, pretty.ServiceClass, obj.Name))
		return nil, gqlerror.New(err, pretty.ServiceInstances)
	}

	instances, err := r.instanceConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", pretty.ServiceInstance))
		return nil, gqlerror.New(err, pretty.ServiceInstance)
	}

	return instances, nil
}

func (r *serviceClassResolver) ServiceClassActivatedField(ctx context.Context, obj *gqlschema.ServiceClass) (bool, error) {
	instances, err := r.ServiceClassInstancesField(ctx, obj)
	if err != nil {
		return false, err
	}

	return len(instances) > 0, nil
}

func (r *serviceClassResolver) ServiceClassClusterDocsTopicField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.ClusterDocsTopic, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve `clusterDocsTopic` field"), pretty.ServiceClass)
		return nil, gqlerror.NewInternal()
	}

	item, err := r.cmsRetriever.ClusterDocsTopic().Find(obj.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", cmsPretty.ClusterDocsTopic, pretty.ServiceClass, obj.Name))
		return nil, gqlerror.New(err, cmsPretty.ClusterDocsTopic)
	}

	clusterDocsTopic, err := r.cmsRetriever.ClusterDocsTopicConverter().ToGQL(item)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", cmsPretty.ClusterDocsTopic))
		return nil, gqlerror.New(err, cmsPretty.ClusterDocsTopic)
	}

	return clusterDocsTopic, nil
}

func (r *serviceClassResolver) ServiceClassDocsTopicField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.DocsTopic, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve `docsTopic` field"), pretty.ServiceClass)
		return nil, gqlerror.NewInternal()
	}

	item, err := r.cmsRetriever.DocsTopic().Find(obj.Namespace, obj.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", cmsPretty.DocsTopic, pretty.ServiceClass, obj.Name))
		return nil, gqlerror.New(err, cmsPretty.DocsTopic)
	}

	docsTopic, err := r.cmsRetriever.DocsTopicConverter().ToGQL(item)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", cmsPretty.DocsTopic))
		return nil, gqlerror.New(err, cmsPretty.DocsTopic)
	}

	return docsTopic, nil
}

func (r *serviceClassResolver) ServiceClassClusterAssetGroupField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.ClusterAssetGroup, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve `clusterAssetGroup` field"), pretty.ServiceClass)
		return nil, gqlerror.NewInternal()
	}

	item, err := r.rafterRetriever.ClusterAssetGroup().Find(obj.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", rafterPretty.ClusterAssetGroup, pretty.ServiceClass, obj.Name))
		return nil, gqlerror.New(err, rafterPretty.ClusterAssetGroup)
	}

	clusterAssetGroup, err := r.rafterRetriever.ClusterAssetGroupConverter().ToGQL(item)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", rafterPretty.ClusterAssetGroup))
		return nil, gqlerror.New(err, rafterPretty.ClusterAssetGroup)
	}

	return clusterAssetGroup, nil
}

func (r *serviceClassResolver) ServiceClassAssetGroupField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.AssetGroup, error) {
	if obj == nil {
		glog.Error(errors.New("%s cannot be empty in order to resolve `assetGroup` field"), pretty.ServiceClass)
		return nil, gqlerror.NewInternal()
	}

	item, err := r.rafterRetriever.AssetGroup().Find(obj.Namespace, obj.Name)
	if err != nil {
		if module.IsDisabledModuleError(err) {
			return nil, err
		}
		glog.Error(errors.Wrapf(err, "while gathering %s for %s %s", rafterPretty.AssetGroup, pretty.ServiceClass, obj.Name))
		return nil, gqlerror.New(err, rafterPretty.AssetGroup)
	}

	assetGroup, err := r.rafterRetriever.AssetGroupConverter().ToGQL(item)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s", rafterPretty.AssetGroup))
		return nil, gqlerror.New(err, rafterPretty.AssetGroup)
	}

	return assetGroup, nil
}
