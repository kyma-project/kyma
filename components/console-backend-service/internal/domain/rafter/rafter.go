package rafter

import (
	"context"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/pkg/errors"
)

const (
	ViewContextLabel = "rafter.kyma-project.io/view-context"
	GroupNameLabel   = "rafter.kyma-project.io/group-name"
	AssetGroupLabel  = "rafter.kyma-project.io/asset-group"
	OrderLabel       = "rafter.kyma-project.io/order"
	TypeLabel      	 = "rafter.kyma-project.io/type"
)

type Config struct {
	Address   string `envconfig:"default=minio.kyma.local"`
	Secure    bool   `envconfig:"default=true"`
	VerifySSL bool   `envconfig:"default=true"`
}

type PluggableContainer struct {
	*module.Pluggable
	cfg Config
	Resolver
	Retriever *retriever
	serviceFactory *resource.ServiceFactory
}

func New(serviceFactory *resource.ServiceFactory, reCfg Config) (*PluggableContainer, error) {
	resolver := &PluggableContainer{
		Pluggable:      module.NewPluggable("rafter"),
		cfg:            reCfg,
		serviceFactory: serviceFactory,
		Retriever: 		&retriever{},
	}

	err := resolver.Disable()
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

func (r *PluggableContainer) Enable() error {
	clusterAssetService, err := newClusterAssetService(r.serviceFactory)
	if err != nil {
		return errors.Wrapf(err, "while creating %s service", pretty.ClusterAssetType)
	}
	clusterAssetConverter := newClusterAssetConverter()

	assetService, err := newAssetService(r.serviceFactory)
	if err != nil {
		return errors.Wrapf(err, "while creating %s service", pretty.AssetType)
	}
	assetConverter := newAssetConverter()

	clusterAssetGroupService, err := newClusterAssetGroupService(r.serviceFactory)
	if err != nil {
		return errors.Wrapf(err, "while creating %s service", pretty.ClusterAssetGroupType)
	}
	clusterAssetGroupConverter := newClusterAssetGroupConverter()

	assetGroupService, err := newAssetGroupService(r.serviceFactory)
	if err != nil {
		return errors.Wrapf(err, "while creating %s service", pretty.AssetGroupType)
	}
	assetGroupConverter := newAssetGroupConverter()

	specificationService, err := newSpecificationService(r.cfg)
	if err != nil {
		return errors.Wrap(err, "while creating Specification Service")
	}

	fileService := newFileService()
	fileConverter := newFileConverter()

	r.Pluggable.EnableAndSyncDynamicInformerFactory(r.serviceFactory.InformerFactory, func() {
		r.Resolver = &domainResolver{
			clusterAssetResolver: newClusterAssetResolver(clusterAssetService, clusterAssetConverter, fileService, fileConverter),
			assetResolver: newAssetResolver(assetService, assetConverter, fileService, fileConverter),
			clusterAssetGroupResolver: newClusterAssetGroupResolver(clusterAssetGroupService, clusterAssetGroupConverter, clusterAssetService, clusterAssetConverter),
			assetGroupResolver: newAssetGroupResolver(assetGroupService, assetGroupConverter, assetService, assetConverter),
		}
		r.Retriever.ClusterAssetGroupGetter = clusterAssetGroupService
		r.Retriever.AssetGroupGetter = assetGroupService
		r.Retriever.GqlClusterAssetGroupConverter = clusterAssetGroupConverter
		r.Retriever.GqlAssetGroupConverter = assetGroupConverter
		r.Retriever.ClusterAssetGetter = clusterAssetService
		r.Retriever.SpecificationSvc = specificationService
	})

	return nil
}

func (r *PluggableContainer) Disable() error {
	r.Pluggable.Disable(func(disabledErr error) {
		r.Resolver = disabled.NewResolver(disabledErr)
		r.Retriever.ClusterAssetGroupGetter = disabled.NewClusterAssetGroupSvc(disabledErr)
		r.Retriever.AssetGroupGetter = disabled.NewAssetGroupSvc(disabledErr)
		r.Retriever.GqlClusterAssetGroupConverter = disabled.NewGqlClusterAssetGroupConverter(disabledErr)
		r.Retriever.GqlAssetGroupConverter = disabled.NewGqlAssetGroupConverter(disabledErr)
		r.Retriever.ClusterAssetGetter = disabled.NewClusterAssetSvc(disabledErr)
		r.Retriever.SpecificationSvc = disabled.NewSpecificationSvc(disabledErr)
	})

	return nil
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	ClusterAssetGroupsQuery(ctx context.Context, viewContext *string, groupName *string) ([]gqlschema.ClusterAssetGroup, error)

	ClusterAssetGroupEventSubscription(ctx context.Context) (<-chan gqlschema.ClusterAssetGroupEvent, error)
	AssetGroupEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.AssetGroupEvent, error)

	ClusterAssetGroupAssetsField(ctx context.Context, obj *gqlschema.ClusterAssetGroup, types []string) ([]gqlschema.RafterClusterAsset, error)
	AssetGroupAssetsField(ctx context.Context, obj *gqlschema.AssetGroup, types []string) ([]gqlschema.RafterAsset, error)

	ClusterAssetEventSubscription(ctx context.Context) (<-chan gqlschema.RafterClusterAssetEvent, error)
	AssetEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.RafterAssetEvent, error)

	ClusterAssetFilesField(ctx context.Context, obj *gqlschema.RafterClusterAsset, filterExtensions []string) ([]gqlschema.File, error)
	AssetFilesField(ctx context.Context, obj *gqlschema.RafterAsset, filterExtensions []string) ([]gqlschema.File, error)
}

type domainResolver struct {
	*clusterAssetResolver
	*assetResolver
	*clusterAssetGroupResolver
	*assetGroupResolver
}
