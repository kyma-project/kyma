package rafter

import (
	"fmt"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/pretty"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/extractor"
	notifierResource "github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"math"
	"sort"
	"strconv"
)

const (
	ViewContext_GroupName_Index = "viewContext/groupName"
	ViewContext_Index = "viewContext"
	GroupName_Index = "groupName"
)

//go:generate mockery -name=clusterAssetGroupSvc -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=clusterAssetGroupSvc -case=underscore -output disabled -outpkg disabled
type clusterAssetGroupSvc interface {
	Find(name string) (*v1beta1.ClusterAssetGroup, error)
	List(viewContext *string, groupName *string) ([]*v1beta1.ClusterAssetGroup, error)
	Subscribe(listener notifierResource.Listener)
	Unsubscribe(listener notifierResource.Listener)
}

type clusterAssetGroupService struct {
	*resource.Service
	notifier notifierResource.Notifier
	extractor extractor.ClusterAssetGroupUnstructuredExtractor
}

func newClusterAssetGroupService(serviceFactory *resource.ServiceFactory) (*clusterAssetGroupService, error) {
	svc := &clusterAssetGroupService{
		Service: serviceFactory.ForResource(schema.GroupVersionResource{
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "clusterassetgroups",
		}),
		extractor: extractor.ClusterAssetGroupUnstructuredExtractor{},
	}

	err := svc.AddIndexers(cache.Indexers{
		ViewContext_GroupName_Index: func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil {
				return nil, errors.New("Cannot convert item")
			}

			return []string{fmt.Sprintf("%s/%s", entity.Labels[ViewContextLabel], entity.Labels[GroupNameLabel])}, nil
		},
		ViewContext_Index: func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil {
				return nil, errors.New("Cannot convert item")
			}

			return []string{entity.Labels[ViewContextLabel]}, nil
		},
		GroupName_Index: func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil {
				return nil, errors.New("Cannot convert item")
			}

			return []string{entity.Labels[GroupNameLabel]}, nil
		},
	})
	if err != nil {
		return nil, err
	}

	notifier := notifierResource.NewNotifier()
	svc.Informer.AddEventHandler(notifier)
	svc.notifier = notifier

	return svc, nil
}

func (svc *clusterAssetGroupService) Find(name string) (*v1beta1.ClusterAssetGroup, error) {
	item, exists, err := svc.Informer.GetStore().GetByKey(name)
	if err != nil || !exists {
		return nil, err
	}

	clusterAssetGroup, err := svc.extractor.Do(item)
	if err != nil {
		return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *%s", item, pretty.ClusterAssetGroupType)
	}

	return clusterAssetGroup, nil
}

func (svc *clusterAssetGroupService) List(viewContext *string, groupName *string) ([]*v1beta1.ClusterAssetGroup, error) {
	var items []interface{}
	var err error
	if viewContext != nil && groupName != nil {
		items, err = svc.Informer.GetIndexer().ByIndex(ViewContext_GroupName_Index, fmt.Sprintf("%s/%s", *viewContext, *groupName))
	} else if viewContext != nil {
		items, err = svc.Informer.GetIndexer().ByIndex(ViewContext_Index, *viewContext)
	} else if groupName != nil {
		items, err = svc.Informer.GetIndexer().ByIndex(GroupName_Index, *groupName)
	} else {
		items = svc.Informer.GetStore().List()
	}

	if err != nil {
		return nil, err
	}

	var clusterAssetGroups []*v1beta1.ClusterAssetGroup
	for _, item := range items {
		clusterAssetGroup, err := svc.extractor.Do(item)
		if err != nil {
			return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *%s", item, pretty.ClusterAssetGroupType)
		}

		clusterAssetGroups = append(clusterAssetGroups, clusterAssetGroup)
	}

	err = svc.sortByOrder(clusterAssetGroups)
	if err != nil {
		return nil, errors.Wrapf(err, "while sorting []*%s by order", pretty.ClusterAssetGroupType)
	}

	return clusterAssetGroups, nil
}

func (svc *clusterAssetGroupService) Subscribe(listener notifierResource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *clusterAssetGroupService) Unsubscribe(listener notifierResource.Listener) {
	svc.notifier.DeleteListener(listener)
}

func (svc *clusterAssetGroupService) sortByOrder(assetGroups []*v1beta1.ClusterAssetGroup) error {
	parseString := func(str string) (uint8, error) {
		if str == "" {
			return math.MaxUint8, nil
		}

		i, err := strconv.ParseInt(str, 10, 16)
		if err != nil {
			return uint8(0), err
		}
		return uint8(i), nil
	}

	var err error
	sort.Slice(assetGroups, func(i, j int) bool {
		if err != nil {
			return false
		}

		firstItem, err := parseString(assetGroups[i].Labels[OrderLabel])
		if err != nil {
			return false
		}

		secondItem, err := parseString(assetGroups[j].Labels[OrderLabel])
		if err != nil {
			return false
		}

		return firstItem < secondItem
	})

	return err
}
