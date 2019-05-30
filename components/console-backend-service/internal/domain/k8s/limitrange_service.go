package k8s

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/apierror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

func newLimitRangeService(informer cache.SharedIndexInformer, client corev1.CoreV1Interface) *limitRangeService {
	return &limitRangeService{
		informer: informer,
		client:   client,
	}
}

type limitRangeService struct {
	informer cache.SharedIndexInformer
	client   corev1.CoreV1Interface
}

func (svc *limitRangeService) List(ns string) ([]*v1.LimitRange, error) {
	items, err := svc.informer.GetIndexer().ByIndex(cache.NamespaceIndex, ns)
	if err != nil {
		return []*v1.LimitRange{}, errors.Wrapf(err, "cannot list limit ranges from ns: %s", ns)
	}

	var result []*v1.LimitRange
	for _, item := range items {
		lr, ok := item.(*v1.LimitRange)
		if !ok {
			return nil, errors.Errorf("unexpected item type: %T, should be *LimitRange", item)
		}
		result = append(result, lr)
	}

	return result, nil
}

func (svc *limitRangeService) Create(namespace string, name string, limit gqlschema.LimitRangeInput) (*v1.LimitRange, error) {
	var errs apierror.ErrorFieldAggregate

	parsedMaxMemory, errMaxMemory := resource.ParseQuantity(*limit.Max.Memory)
	if errMaxMemory != nil {
		errs = append(errs, apierror.NewInvalidField("max.memory", *limit.Max.Memory, fmt.Sprintf("while parsing %s limit range", pretty.LimitRange)))
	}

	parsedDefaultMemory, errDefaultMemory := resource.ParseQuantity(*limit.Default.Memory)
	if errDefaultMemory != nil {
		errs = append(errs, apierror.NewInvalidField("default.memory", *limit.Default.Memory, fmt.Sprintf("while parsing %s limit range", pretty.LimitRange)))
	}

	parsedDefaultRequestMemory, errDefaultRequestMemory := resource.ParseQuantity(*limit.DefaultRequest.Memory)
	if errDefaultRequestMemory != nil {
		errs = append(errs, apierror.NewInvalidField("defaultRequest.memory", *limit.DefaultRequest.Memory, fmt.Sprintf("while parsing %s limit range", pretty.LimitRange)))
	}

	if limit.Type != "Container" && limit.Type != "Pod" {
		errs = append(errs, apierror.NewInvalidField("type", limit.Type, fmt.Sprintf("while parsing %s limit range, type has to be one of: 'Container', 'Pod'", pretty.LimitRange)))
	}

	if len(errs) > 0 {
		return nil, apierror.NewInvalid(pretty.LimitRange, errs)
	}

	limitRange := v1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.LimitRangeSpec{
			Limits: []v1.LimitRangeItem{
				{
					Max: v1.ResourceList{
						v1.ResourceMemory: parsedMaxMemory,
					},
					Default: v1.ResourceList{
						v1.ResourceMemory: parsedDefaultMemory,
					},
					DefaultRequest: v1.ResourceList{
						v1.ResourceMemory: parsedDefaultRequestMemory,
					},
					Type: v1.LimitType(limit.Type),
				},
			},
		},
	}

	return svc.client.LimitRanges(namespace).Create(&limitRange)
}
