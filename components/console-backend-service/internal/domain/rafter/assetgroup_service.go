package rafter

import (
	"fmt"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/pretty"
	notifierResource "github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate mockery -name=assetGroupSvc -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=assetGroupSvc -case=underscore -output disabled -outpkg disabled
type assetGroupSvc interface {
	Find(namespace, name string) (*v1beta1.AssetGroup, error)
	Subscribe(listener notifierResource.Listener)
	Unsubscribe(listener notifierResource.Listener)
}

type assetGroupService struct {
	*resource.Service
	notifier notifierResource.Notifier
	extractor extractor.AssetGroupUnstructuredExtractor
}

func newAssetGroupService(serviceFactory *resource.ServiceFactory) (*assetGroupService, error) {
	svc := &assetGroupService{
		Service: serviceFactory.ForResource(schema.GroupVersionResource{
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "assetgroups",
		}),
		extractor: extractor.AssetGroupUnstructuredExtractor{},
	}

	notifier := notifierResource.NewNotifier()
	svc.Informer.AddEventHandler(notifier)
	svc.notifier = notifier

	return svc, nil
}

func (svc *assetGroupService) Find(namespace, name string) (*v1beta1.AssetGroup, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := svc.Informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	assetGroup, err := svc.extractor.Do(item)
	if err != nil {
		return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *%s", item, pretty.AssetGroupType)
	}

	return assetGroup, nil
}

func (svc *assetGroupService) Subscribe(listener notifierResource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *assetGroupService) Unsubscribe(listener notifierResource.Listener) {
	svc.notifier.DeleteListener(listener)
}
