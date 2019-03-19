package cms

import (
	"fmt"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/cms/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/cache"
)

//go:generate mockery -name=docsTopicSvc -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=docsTopicSvc -case=underscore -output disabled -outpkg disabled
type docsTopicSvc interface {
	Find(namespace, name string) (*v1alpha1.DocsTopic, error)
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

type docsTopicService struct {
	informer  cache.SharedIndexInformer
	notifier  notifier
	extractor extractor.DocsTopicUnstructuredExtractor
}

func newDocsTopicService(informer cache.SharedIndexInformer) (*docsTopicService, error) {
	svc := &docsTopicService{
		informer:  informer,
		extractor: extractor.DocsTopicUnstructuredExtractor{},
	}

	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)
	svc.notifier = notifier

	return svc, nil
}

func (svc *docsTopicService) Find(namespace, name string) (*v1alpha1.DocsTopic, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	docsTopic, err := svc.extractor.Do(item)
	if err != nil {
		return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *DocsTopic", item)
	}

	return docsTopic, nil
}

func (svc *docsTopicService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *docsTopicService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}
