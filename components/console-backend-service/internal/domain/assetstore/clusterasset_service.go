package assetstore

import (
	"fmt"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/cache"
)

//go:generate mockery -name=clusterAssetSvc -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=clusterAssetSvc -case=underscore -output disabled -outpkg disabled
type clusterAssetSvc interface {
	Find(name string) (*v1alpha2.ClusterAsset, error)
	ListForDocsTopicByType(docsTopicName string, types []string) ([]*v1alpha2.ClusterAsset, error)
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

type clusterAssetService struct {
	informer  cache.SharedIndexInformer
	notifier  notifier
	extractor extractor.ClusterAssetUnstructuredExtractor
}

func newClusterAssetService(informer cache.SharedIndexInformer) (*clusterAssetService, error) {
	svc := &clusterAssetService{
		informer:  informer,
		extractor: extractor.ClusterAssetUnstructuredExtractor{},
	}

	err := svc.informer.AddIndexers(cache.Indexers{
		"docsTopicName": func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil {
				return nil, errors.New("Cannot convert item")
			}

			return []string{entity.Labels[CmsDocsTopicLabel]}, nil
		},
		"docsTopicName/type": func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil {
				return nil, errors.New("Cannot convert item")
			}

			return []string{fmt.Sprintf("%s/%s", entity.Labels[CmsDocsTopicLabel], entity.Labels[CmsTypeLabel])}, nil
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "while adding indexers")
	}

	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)
	svc.notifier = notifier

	return svc, nil
}

func (svc *clusterAssetService) Find(name string) (*v1alpha2.ClusterAsset, error) {
	item, exists, err := svc.informer.GetStore().GetByKey(name)
	if err != nil || !exists {
		return nil, err
	}

	clusterAsset, err := svc.extractor.Do(item)
	if err != nil {
		return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *ClusterAsset", item)
	}

	return clusterAsset, nil
}

func (svc *clusterAssetService) ListForDocsTopicByType(docsTopicName string, types []string) ([]*v1alpha2.ClusterAsset, error) {
	var items []interface{}
	var err error
	if len(types) == 0 {
		items, err = svc.informer.GetIndexer().ByIndex("docsTopicName", docsTopicName)
	} else {
		for _, typeArg := range types {
			its, err := svc.informer.GetIndexer().ByIndex("docsTopicName/type", fmt.Sprintf("%s/%s", docsTopicName, typeArg))
			if err != nil {
				return nil, err
			}
			items = append(items, its...)
		}
	}

	if err != nil {
		return nil, err
	}

	var clusterAssets []*v1alpha2.ClusterAsset
	for _, item := range items {
		clusterAsset, err := svc.extractor.Do(item)
		if err != nil {
			return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *ClusterAsset", item)
		}

		clusterAssets = append(clusterAssets, clusterAsset)
	}

	return clusterAssets, nil
}

func (svc *clusterAssetService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *clusterAssetService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}
