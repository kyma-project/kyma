package cms

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

const (
	ViewContextLabel = "cms.kyma-project.io/view-context"
	GroupNameLabel   = "cms.kyma-project.io/group-name"
	DocsTopicLabel   = "cms.kyma-project.io/docs-topic"
	OrderLabel       = "cms.kyma-project.io/order"
)

type cmsRetriever struct {
	ClusterDocsTopicGetter       shared.ClusterDocsTopicGetter
	DocsTopicGetter              shared.DocsTopicGetter
	GqlClusterDocsTopicConverter shared.GqlClusterDocsTopicConverter
	GqlDocsTopicConverter        shared.GqlDocsTopicConverter
}

func (r *cmsRetriever) ClusterDocsTopic() shared.ClusterDocsTopicGetter {
	return r.ClusterDocsTopicGetter
}

func (r *cmsRetriever) DocsTopic() shared.DocsTopicGetter {
	return r.DocsTopicGetter
}

func (r *cmsRetriever) ClusterDocsTopicConverter() shared.GqlClusterDocsTopicConverter {
	return r.GqlClusterDocsTopicConverter
}

func (r *cmsRetriever) DocsTopicConverter() shared.GqlDocsTopicConverter {
	return r.GqlDocsTopicConverter
}

type PluggableContainer struct {
	*module.Pluggable
	cfg *resolverConfig

	Resolver        Resolver
	CmsRetriever    *cmsRetriever
	informerFactory dynamicinformer.DynamicSharedInformerFactory
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration, assetStoreRetriever shared.AssetStoreRetriever) (*PluggableContainer, error) {
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Dynamic Clientset")
	}

	container := &PluggableContainer{
		cfg: &resolverConfig{
			dynamicClient:        dynamicClient,
			informerResyncPeriod: informerResyncPeriod,
			assetStoreRetriever:  assetStoreRetriever,
		},
		Pluggable:    module.NewPluggable("cms"),
		CmsRetriever: &cmsRetriever{},
	}

	err = container.Disable()
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (r *PluggableContainer) Enable() error {
	informerResyncPeriod := r.cfg.informerResyncPeriod
	dynamicClient := r.cfg.dynamicClient

	assetStoreRetriever := r.cfg.assetStoreRetriever

	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, informerResyncPeriod)
	r.informerFactory = informerFactory

	clusterDocsTopicService, err := newClusterDocsTopicService(informerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "clusterdocstopics",
	}).Informer())
	if err != nil {
		return errors.Wrapf(err, "while creating clusterDocsTopic service")
	}

	docsTopicService, err := newDocsTopicService(informerFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "docstopics",
	}).Informer())
	if err != nil {
		return errors.Wrapf(err, "while creating docsTopic service")
	}

	r.Pluggable.EnableAndSyncDynamicInformerFactory(r.informerFactory, func() {
		r.Resolver = &domainResolver{
			clusterDocsTopicResolver: newClusterDocsTopicResolver(clusterDocsTopicService, assetStoreRetriever),
			docsTopicResolver:        newDocsTopicResolver(docsTopicService, assetStoreRetriever),
		}
		r.CmsRetriever.ClusterDocsTopicGetter = clusterDocsTopicService
		r.CmsRetriever.DocsTopicGetter = docsTopicService
		r.CmsRetriever.GqlClusterDocsTopicConverter = &clusterDocsTopicConverter{}
		r.CmsRetriever.GqlDocsTopicConverter = &docsTopicConverter{}
	})

	return nil
}

func (r *PluggableContainer) Disable() error {
	r.Pluggable.Disable(func(disabledErr error) {
		r.Resolver = disabled.NewResolver(disabledErr)
		r.CmsRetriever.ClusterDocsTopicGetter = disabled.NewClusterDocsTopicSvc(disabledErr)
		r.CmsRetriever.DocsTopicGetter = disabled.NewDocsTopicSvc(disabledErr)
		r.CmsRetriever.GqlClusterDocsTopicConverter = disabled.NewGqlClusterDocsTopicConverter(disabledErr)
		r.CmsRetriever.GqlDocsTopicConverter = disabled.NewGqlDocsTopicConverter(disabledErr)
		r.informerFactory = nil
	})

	return nil
}

type resolverConfig struct {
	dynamicClient        dynamic.Interface
	informerResyncPeriod time.Duration
	assetStoreRetriever  shared.AssetStoreRetriever
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	ClusterDocsTopicsQuery(ctx context.Context, viewContext *string, groupName *string) ([]gqlschema.ClusterDocsTopic, error)
	ClusterDocsTopicEventSubscription(ctx context.Context) (<-chan gqlschema.ClusterDocsTopicEvent, error)
	DocsTopicEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.DocsTopicEvent, error)
	ClusterDocsTopicAssetsField(ctx context.Context, obj *gqlschema.ClusterDocsTopic, types []string) ([]gqlschema.ClusterAsset, error)
	DocsTopicAssetsField(ctx context.Context, obj *gqlschema.DocsTopic, types []string) ([]gqlschema.Asset, error)
}

type domainResolver struct {
	*clusterDocsTopicResolver
	*docsTopicResolver
}
