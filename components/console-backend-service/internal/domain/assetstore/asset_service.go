package assetstore

import (
	"fmt"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/cache"
)

//go:generate mockery -name=assetSvc -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=assetSvc -case=underscore -output disabled -outpkg disabled
type assetSvc interface {
	Find(namespace, name string) (*v1alpha2.Asset, error)
	ListForDocsTopicByType(namespace, docsTopicName string, types []string) ([]*v1alpha2.Asset, error)
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

type assetService struct {
	informer  cache.SharedIndexInformer
	notifier  notifier
	extractor extractor.AssetUnstructuredExtractor
}

func newAssetService(informer cache.SharedIndexInformer) (*assetService, error) {
	svc := &assetService{
		informer:  informer,
		extractor: extractor.AssetUnstructuredExtractor{},
	}

	err := svc.informer.AddIndexers(cache.Indexers{
		"ns/docsTopicName": func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil {
				return nil, errors.New("Cannot convert item")
			}

			return []string{fmt.Sprintf("%s/%s", entity.Namespace, entity.Labels[CmsDocsTopicLabel])}, nil
		},
		"ns/docsTopicName/type": func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil {
				return nil, errors.New("Cannot convert item")
			}

			return []string{fmt.Sprintf("%s/%s/%s", entity.Namespace, entity.Labels[CmsDocsTopicLabel], entity.Labels[CmsTypeLabel])}, nil
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

func (svc *assetService) Find(namespace, name string) (*v1alpha2.Asset, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)

	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	asset, err := svc.extractor.Do(item)
	if err != nil {
		return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *Asset", item)
	}

	return asset, nil
}

func (svc *assetService) ListForDocsTopicByType(namespace, docsTopicName string, types []string) ([]*v1alpha2.Asset, error) {
	var items []interface{}
	var err error
	if len(types) == 0 {
		items, err = svc.informer.GetIndexer().ByIndex("ns/docsTopicName", fmt.Sprintf("%s/%s", namespace, docsTopicName))
	} else {
		for _, typeArg := range types {
			its, err := svc.informer.GetIndexer().ByIndex("ns/docsTopicName/type", fmt.Sprintf("%s/%s/%s", namespace, docsTopicName, typeArg))
			if err != nil {
				return nil, err
			}
			items = append(items, its...)
		}
	}

	if err != nil {
		return nil, err
	}

	var assets []*v1alpha2.Asset
	for _, item := range items {
		asset, err := svc.extractor.Do(item)
		if err != nil {
			return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *Asset", item)
		}

		assets = append(assets, asset)
	}

	return assets, nil
}

func (svc *assetService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *assetService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}
