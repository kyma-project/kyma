package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/pretty"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// TODO: Write unit tests

const envLabelSelector = "env=true"

type namespaceService struct {
	nsInterface corev1.NamespaceInterface
}

func newNamespaceService(nsInterface corev1.NamespaceInterface) *namespaceService {
	return &namespaceService{
		nsInterface: nsInterface,
	}
}

func (svc *namespaceService) List() ([]v1.Namespace, error) {
	list, err := svc.nsInterface.List(metav1.ListOptions{
		LabelSelector: envLabelSelector, // namespaces with label env=true are treated as customer namespaces
	})
	if err != nil {
		return nil, errors.Wrapf(err, "while listing %s", pretty.Namespace)
	}

	return list.Items, nil
}
