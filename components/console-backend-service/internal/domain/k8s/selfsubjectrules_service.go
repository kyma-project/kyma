package k8s

import (
	"context"
	"fmt"

	authv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

type selfSubjectRulesService struct {
	client v1.AuthorizationV1Interface
}

func newSelfSubjectRulesService(client v1.AuthorizationV1Interface) *selfSubjectRulesService {
	return &selfSubjectRulesService{
		client: client,
	}
}

func (svc *selfSubjectRulesService) Create(ctx context.Context, ssrr *authv1.SelfSubjectRulesReview) (*authv1.SelfSubjectRulesReview, error) {

	token := ctx.Value("token").(string)
	result := svc.client.RESTClient().Post().
		AbsPath("/apis/authorization.k8s.io/v1").
		Resource("selfsubjectrulesreviews").
		Body(ssrr).
		SetHeader("Authorization", token).
		Do()

	response, err := result.Get()
	ssrrout := response.(*authv1.SelfSubjectRulesReview)
	fmt.Printf("REsponse : %-v", ssrrout)

	return ssrrout, err
}

// func (svc *serviceService) List(namespace string, pagingParams pager.PagingParams) ([]*v1.Service, error) {
// 	items, err := pager.FromIndexer(svc.informer.GetIndexer(), "namespace", namespace).Limit(pagingParams)
// 	if err != nil {
// 		return nil, err
// 	}
// 	services := make([]*v1.Service, len(items))
// 	for i, item := range items {
// 		service, ok := item.(*v1.Service)
// 		if !ok {
// 			return nil, fmt.Errorf("incorrect item type: %T, should be: *Service", item)
// 		}
// 		service.TypeMeta = metav1.TypeMeta{
// 			Kind:       "Service",
// 			APIVersion: "v1",
// 		}
// 		services[i] = service
// 	}

// 	return services, nil
// }

// func (svc *serviceService) Find(name, namespace string) (*v1.Service, error) {
// 	key := fmt.Sprintf("%s/%s", namespace, name)

// 	item, exists, err := svc.informer.GetStore().GetByKey(key)
// 	if err != nil || !exists {
// 		return nil, err
// 	}

// 	service, ok := item.(*v1.Service)
// 	if !ok {
// 		return nil, fmt.Errorf("incorrect item type: %T, should be: *v1.Service", item)
// 	}
// 	svc.ensureTypeMeta(service)
// 	return service, nil
// }

// func (svc *serviceService) ensureTypeMeta(service *v1.Service) {
// 	service.TypeMeta = svc.serviceTypeMetadata()
// }

// func (svc *serviceService) serviceTypeMetadata() metav1.TypeMeta {
// 	return metav1.TypeMeta{
// 		Kind:       "Service",
// 		APIVersion: "v1",
// 	}
// }

// func (svc *serviceService) Subscribe(listener resource.Listener) {
// 	svc.notifier.AddListener(listener)
// }

// func (svc *serviceService) Unsubscribe(listener resource.Listener) {
// 	svc.notifier.DeleteListener(listener)
// }

// func (svc *serviceService) Update(name, namespace string, update v1.Service) (*v1.Service, error) {
// 	err := svc.checkUpdatePreconditions(name, namespace, update)
// 	if err != nil {
// 		return nil, err
// 	}

// 	updated, err := svc.client.Services(namespace).Update(&update)
// 	if err != nil {
// 		return nil, err
// 	}

// 	svc.ensureTypeMeta(updated)

// 	return updated, nil
// }

// func (svc *serviceService) Delete(name, namespace string) error {
// 	return svc.client.Services(namespace).Delete(name, nil)
// }

// func (svc *serviceService) checkUpdatePreconditions(name string, namespace string, update v1.Service) error {
// 	var errs apierror.ErrorFieldAggregate
// 	if name != update.Name {
// 		errs = append(errs, apierror.NewInvalidField("metadata.name", update.Name, fmt.Sprintf("name of updated object does not match the original (%s)", name)))
// 	}
// 	if namespace != update.Namespace {
// 		errs = append(errs, apierror.NewInvalidField("metadata.namespace", update.Namespace, fmt.Sprintf("namespace of updated object does not match the original (%s)", namespace)))
// 	}
// 	typeMeta := svc.serviceTypeMetadata()
// 	if update.Kind != typeMeta.Kind {
// 		errs = append(errs, apierror.NewInvalidField("kind", update.Kind, "service's kind should not be changed"))
// 	}
// 	if update.APIVersion != typeMeta.APIVersion {
// 		errs = append(errs, apierror.NewInvalidField("apiVersion", update.APIVersion, "service's apiVersion should not be changed"))
// 	}

// 	if len(errs) > 0 {
// 		return apierror.NewInvalid(pretty.Service, errs)
// 	}

// 	return nil
// }
