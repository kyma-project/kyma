package rafter

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	notifierResource "github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

const (
	Namespace_AssetGroupName_Index      = "namespace/assetGroupName"
	Namespace_AssetGroupName_Type_Index = "namespace/assetGroupName/type"
)

//go:generate mockery -name=assetSvc -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=assetSvc -case=underscore -output disabled -outpkg disabled
type assetSvc interface {
	Find(namespace, name string) (*v1beta1.Asset, error)
	ListForAssetGroupByType(namespace, assetGroupName string, types []string) ([]*v1beta1.Asset, error)
	Subscribe(listener notifierResource.Listener)
	Unsubscribe(listener notifierResource.Listener)
}

type assetService struct {
	*resource.Service
	notifier  notifierResource.Notifier
	extractor extractor.AssetUnstructuredExtractor
}

func newAssetService(serviceFactory *resource.ServiceFactory) (*assetService, error) {
	svc := &assetService{
		Service: serviceFactory.ForResource(schema.GroupVersionResource{
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "assets",
		}),
		extractor: extractor.AssetUnstructuredExtractor{},
	}

	err := svc.AddIndexers(cache.Indexers{
		Namespace_AssetGroupName_Index: func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil {
				return nil, errors.New("Cannot convert item")
			}

			return []string{fmt.Sprintf("%s/%s", entity.Namespace, entity.Labels[AssetGroupLabel])}, nil
		},
		Namespace_AssetGroupName_Type_Index: func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil {
				return nil, errors.New("Cannot convert item")
			}

			return []string{fmt.Sprintf("%s/%s/%s", entity.Namespace, entity.Labels[AssetGroupLabel], entity.Labels[TypeLabel])}, nil
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

func (svc *assetService) Find(namespace, name string) (*v1beta1.Asset, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := svc.Informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	asset, err := svc.extractor.Do(item)
	if err != nil {
		return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *%s", item, pretty.AssetType)
	}

	return asset, nil
}

func (svc *assetService) ListForAssetGroupByType(namespace, assetGroupName string, types []string) ([]*v1beta1.Asset, error) {
	var items []interface{}
	var err error
	if len(types) == 0 {
		items, err = svc.Informer.GetIndexer().ByIndex(Namespace_AssetGroupName_Index, fmt.Sprintf("%s/%s", namespace, assetGroupName))
	} else {
		for _, typeArg := range types {
			its, err := svc.Informer.GetIndexer().ByIndex(Namespace_AssetGroupName_Type_Index, fmt.Sprintf("%s/%s/%s", namespace, assetGroupName, typeArg))
			if err != nil {
				return nil, err
			}
			items = append(items, its...)
		}
	}

	if err != nil {
		return nil, err
	}

	var assets []*v1beta1.Asset
	for _, item := range items {
		asset, err := svc.extractor.Do(item)
		if err != nil {
			return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *%s", item, pretty.AssetType)
		}

		assets = append(assets, asset)
	}

	return assets, nil
}

func (svc *assetService) Subscribe(listener notifierResource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *assetService) Unsubscribe(listener notifierResource.Listener) {
	svc.notifier.DeleteListener(listener)
}
