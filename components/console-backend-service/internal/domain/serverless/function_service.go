package serverless

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	notifierResource "github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//go:generate mockery -name=functionSvc -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=functionSvc -case=underscore -output disabled -outpkg disabled
type functionSvc interface {
	Find(namespace, name string) (*v1alpha1.Function, error)
	List(namespace string) ([]*v1alpha1.Function, error)
	Create(function *v1alpha1.Function) (*v1alpha1.Function, error)
	Update(function *v1alpha1.Function) (*v1alpha1.Function, error)
	Delete(function gqlschema.FunctionMetadataInput) error
	DeleteMany(functions []gqlschema.FunctionMetadataInput) error
	Subscribe(listener notifierResource.Listener)
	Unsubscribe(listener notifierResource.Listener)
}

type functionService struct {
	*resource.Service
	notifier  notifierResource.Notifier
	extractor *functionUnstructuredExtractor
}

var functionTypeMeta = metav1.TypeMeta{
	Kind:       "Function",
	APIVersion: "serverless.kyma-project.io/v1alpha1",
}

func newFunctionService(serviceFactory *resource.ServiceFactory) *functionService {
	svc := &functionService{
		Service: serviceFactory.ForResource(schema.GroupVersionResource{
			Version:  v1alpha1.GroupVersion.Version,
			Group:    v1alpha1.GroupVersion.Group,
			Resource: "functions",
		}),
		extractor: &functionUnstructuredExtractor{},
	}

	notifier := notifierResource.NewNotifier()
	svc.Informer.AddEventHandler(notifier)
	svc.notifier = notifier

	return svc
}

func (svc *functionService) Find(namespace, name string) (*v1alpha1.Function, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := svc.Informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	function, err := svc.extractor.do(item)
	if err != nil {
		return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *%s", item, pretty.FunctionType)
	}

	return function, nil
}

func (svc *functionService) List(namespace string) ([]*v1alpha1.Function, error) {
	items, err := svc.Informer.GetIndexer().ByIndex("namespace", namespace)
	if err != nil {
		return nil, err
	}

	var functions []*v1alpha1.Function
	for _, item := range items {
		function, err := svc.extractor.do(item)
		if err != nil {
			return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *%s", item, pretty.FunctionType)
		}

		functions = append(functions, function)
	}

	return functions, nil
}

func (svc *functionService) Create(function *v1alpha1.Function) (*v1alpha1.Function, error) {
	if function == nil {
		return nil, errors.New(fmt.Sprintf("%s can't be nil", pretty.FunctionType))
	}
	function.TypeMeta = functionTypeMeta

	u, err := svc.extractor.toUnstructured(function)
	if err != nil {
		return nil, err
	}

	created, err := svc.Client.Namespace(function.ObjectMeta.Namespace).Create(u, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return svc.extractor.fromUnstructured(created)
}

func (svc *functionService) Update(function *v1alpha1.Function) (*v1alpha1.Function, error) {
	oldFunction, err := svc.Find(function.ObjectMeta.Namespace, function.ObjectMeta.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "while finding %s [name: %s]", pretty.FunctionType, function.ObjectMeta.Name)
	}

	if oldFunction == nil {
		return nil, apiErrors.NewNotFound(schema.GroupResource{
			Group:    v1alpha1.GroupVersion.Group,
			Resource: "functions",
		}, function.ObjectMeta.Name)
	}
	function.ObjectMeta.ResourceVersion = oldFunction.ObjectMeta.ResourceVersion
	function.TypeMeta = functionTypeMeta

	u, err := svc.extractor.toUnstructured(function)
	if err != nil {
		return nil, err
	}

	updated, err := svc.Client.Namespace(function.ObjectMeta.Namespace).Update(u, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return svc.extractor.fromUnstructured(updated)
}

func (svc *functionService) Delete(function gqlschema.FunctionMetadataInput) error {
	return svc.Client.Namespace(function.Namespace).Delete(function.Name, &metav1.DeleteOptions{})
}

func (svc *functionService) DeleteMany(functions []gqlschema.FunctionMetadataInput) error {
	for _, function := range functions {
		err := svc.Delete(function)
		if err != nil {
			return err
		}
	}
	return nil
}

func (svc *functionService) Subscribe(listener notifierResource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *functionService) Unsubscribe(listener notifierResource.Listener) {
	svc.notifier.DeleteListener(listener)
}
