package serverless

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/convert"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type functionService struct {
	*resource.Service
}

func newFunctionService(serviceFactory *resource.ServiceFactory) *functionService {
	return &functionService{
		Service: serviceFactory.ForResource(schema.GroupVersionResource{
			Version:  v1alpha1.SchemeGroupVersion.Version,
			Group:    v1alpha1.SchemeGroupVersion.Group,
			Resource: "functions",
		}),
	}
}

func (svc *functionService) List(namespace string) ([]*v1alpha1.Function, error) {
	results := make([]*v1alpha1.Function, 0)
	err := svc.ListInIndex("namespace", namespace, &results)
	return results, err
}

func (svc *functionService) Delete(name string, namespace string) error {
	return svc.Client.Namespace(namespace).Delete(name, &metav1.DeleteOptions{})
}

func (svc *functionService) Create(name string, namespace string, labels gqlschema.Labels, size string, runtime string) (*v1alpha1.Function, error) {
	function, err := convert.FunctionToUnstructured(&v1alpha1.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "serverless.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: v1alpha1.FunctionSpec{
			Size:    size,
			Runtime: runtime,
		},
	})
	if err != nil {
		return &v1alpha1.Function{}, err
	}

	created, err := svc.Client.Namespace(namespace).Create(function, metav1.CreateOptions{})
	if err != nil {
		return &v1alpha1.Function{}, err
	}

	return convert.UnstructuredToFunction(created)
}
