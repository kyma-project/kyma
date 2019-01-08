package k8s

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application/pretty"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// TODO: Write unit tests

const envLabelSelector = "env=true"

type environmentService struct {
	nsInterface corev1.NamespaceInterface
}

func newEnvironmentService(nsInterface corev1.NamespaceInterface) *environmentService {
	return &environmentService{
		nsInterface: nsInterface,
	}
}

func (svc *environmentService) List() ([]v1.Namespace, error) {
	list, err := svc.nsInterface.List(metav1.ListOptions{
		LabelSelector: envLabelSelector, // namespaces with label env=true are treated as environments
	})
	if err != nil {
		return nil, errors.Wrapf(err, "while listing %s", pretty.Environment)
	}

	return list.Items, nil
}
