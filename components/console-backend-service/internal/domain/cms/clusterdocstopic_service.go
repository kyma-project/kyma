package cms

import (
	"fmt"

	"math"
	"sort"
	"strconv"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/cache"
)

//go:generate mockery -name=clusterDocsTopicSvc -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=clusterDocsTopicSvc -case=underscore -output disabled -outpkg disabled
type clusterDocsTopicSvc interface {
	Find(name string) (*v1alpha1.ClusterDocsTopic, error)
	List(viewContext *string, groupName *string) ([]*v1alpha1.ClusterDocsTopic, error)
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

type clusterDocsTopicService struct {
	informer  cache.SharedIndexInformer
	notifier  notifier
	extractor extractor.ClusterDocsTopicUnstructuredExtractor
}

func newClusterDocsTopicService(informer cache.SharedIndexInformer) (*clusterDocsTopicService, error) {
	svc := &clusterDocsTopicService{
		informer:  informer,
		extractor: extractor.ClusterDocsTopicUnstructuredExtractor{},
	}

	err := svc.informer.AddIndexers(cache.Indexers{
		"viewContext/groupName": func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil {
				return nil, errors.New("Cannot convert item")
			}

			return []string{fmt.Sprintf("%s/%s", entity.Labels[ViewContextLabel], entity.Labels[GroupNameLabel])}, nil
		},
		"viewContext": func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil {
				return nil, errors.New("Cannot convert item")
			}

			return []string{entity.Labels[ViewContextLabel]}, nil
		},
		"groupName": func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil {
				return nil, errors.New("Cannot convert item")
			}

			return []string{entity.Labels[GroupNameLabel]}, nil
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

func (svc *clusterDocsTopicService) Find(name string) (*v1alpha1.ClusterDocsTopic, error) {
	item, exists, err := svc.informer.GetStore().GetByKey(name)
	if err != nil || !exists {
		return nil, err
	}

	clusterDocsTopic, err := svc.extractor.Do(item)
	if err != nil {
		return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *ClusterDocsTopic", item)
	}

	return clusterDocsTopic, nil
}

func (svc *clusterDocsTopicService) List(viewContext *string, groupName *string) ([]*v1alpha1.ClusterDocsTopic, error) {
	var items []interface{}
	var err error
	if viewContext != nil && groupName != nil {
		items, err = svc.informer.GetIndexer().ByIndex("viewContext/groupName", fmt.Sprintf("%s/%s", *viewContext, *groupName))
	} else if viewContext != nil {
		items, err = svc.informer.GetIndexer().ByIndex("viewContext", *viewContext)
	} else if groupName != nil {
		items, err = svc.informer.GetIndexer().ByIndex("groupName", *groupName)
	} else {
		items = svc.informer.GetStore().List()
	}

	if err != nil {
		return nil, err
	}

	var clusterDocsTopics []*v1alpha1.ClusterDocsTopic
	for _, item := range items {
		clusterDocsTopic, err := svc.extractor.Do(item)
		if err != nil {
			return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *ClusterDocsTopic", item)
		}

		clusterDocsTopics = append(clusterDocsTopics, clusterDocsTopic)
	}

	err = svc.sortByOrder(clusterDocsTopics)
	if err != nil {
		return nil, errors.Wrap(err, "while sorting []*ClusterDocsTopic by order")
	}

	return clusterDocsTopics, nil
}

func (svc *clusterDocsTopicService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *clusterDocsTopicService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}

func (svc *clusterDocsTopicService) sortByOrder(docsTopics []*v1alpha1.ClusterDocsTopic) error {
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
	sort.Slice(docsTopics, func(i, j int) bool {
		if err != nil {
			return false
		}

		firstItem, err := parseString(docsTopics[i].Labels[OrderLabel])
		if err != nil {
			return false
		}

		secondItem, err := parseString(docsTopics[j].Labels[OrderLabel])
		if err != nil {
			return false
		}

		return firstItem < secondItem
	})

	return err
}
