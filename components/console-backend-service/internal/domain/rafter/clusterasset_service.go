package rafter

import (
	"fmt"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/pretty"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	notifierResource "github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

const (
	AssetGroupName_Index = "assetGroupName"
	AssetGroupName_Type_Index = "assetGroupName/type"
)

//go:generate mockery -name=clusterAssetSvc -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=clusterAssetSvc -case=underscore -output disabled -outpkg disabled
type clusterAssetSvc interface {
	Find(name string) (*v1beta1.ClusterAsset, error)
	ListForClusterAssetGroupByType(assetGroupName string, types []string) ([]*v1beta1.ClusterAsset, error)
	Subscribe(listener notifierResource.Listener)
	Unsubscribe(listener notifierResource.Listener)
}

type clusterAssetService struct {
	*resource.Service
	notifier notifierResource.Notifier
	extractor extractor.ClusterAssetUnstructuredExtractor
}

func newClusterAssetService(serviceFactory *resource.ServiceFactory) (*clusterAssetService, error) {
	svc := &clusterAssetService{
		Service: serviceFactory.ForResource(schema.GroupVersionResource{
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "clusterassets",
		}),
		extractor: extractor.ClusterAssetUnstructuredExtractor{},
	}

	svc.Informer.HasSynced()

	err := svc.AddIndexers(cache.Indexers{
		AssetGroupName_Index: func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil {
				return nil, errors.New("Cannot convert item")
			}

			return []string{entity.Labels[AssetGroupLabel]}, nil
		},
		AssetGroupName_Type_Index: func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil {
				return nil, errors.New("Cannot convert item")
			}

			return []string{fmt.Sprintf("%s/%s", entity.Labels[AssetGroupLabel], entity.Labels[TypeLabel])}, nil
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "while adding indexers")
	}

	notifier := notifierResource.NewNotifier()
	svc.Informer.AddEventHandler(notifier)
	svc.notifier = notifier

	return svc, nil
}

func (svc *clusterAssetService) Find(name string) (*v1beta1.ClusterAsset, error) {
	item, exists, err := svc.Informer.GetStore().GetByKey(name)
	if err != nil || !exists {
		return nil, err
	}

	clusterAsset, err := svc.extractor.Do(item)
	if err != nil {
		return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *%s", item, pretty.ClusterAssetType)
	}

	return clusterAsset, nil
}

func (svc *clusterAssetService) ListForClusterAssetGroupByType(assetGroupName string, types []string) ([]*v1beta1.ClusterAsset, error) {
	var items []interface{}
	var err error
	if len(types) == 0 {
		items, err = svc.Informer.GetIndexer().ByIndex(AssetGroupName_Index, assetGroupName)
	} else {
		for _, typeArg := range types {
			its, err := svc.Informer.GetIndexer().ByIndex(AssetGroupName_Type_Index, fmt.Sprintf("%s/%s", assetGroupName, typeArg))
			if err != nil {
				return nil, err
			}
			items = append(items, its...)
		}
	}

	if err != nil {
		return nil, err
	}

	var clusterAssets []*v1beta1.ClusterAsset
	for _, item := range items {
		clusterAsset, err := svc.extractor.Do(item)
		if err != nil {
			return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *%s", item, pretty.ClusterAssetType)
		}

		clusterAssets = append(clusterAssets, clusterAsset)
	}

	return clusterAssets, nil
}

func (svc *clusterAssetService) Subscribe(listener notifierResource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *clusterAssetService) Unsubscribe(listener notifierResource.Listener) {
	svc.notifier.DeleteListener(listener)
}
