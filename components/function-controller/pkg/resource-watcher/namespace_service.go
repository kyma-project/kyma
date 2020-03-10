package resource_watcher

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type NamespaceService struct {
	coreClient v1.CoreV1Interface
	config     Config
}

func NewNamespaceService(coreClient v1.CoreV1Interface, config Config) *NamespaceService {
	return &NamespaceService{
		coreClient: coreClient,
		config:     config,
	}
}

func (s *NamespaceService) GetNamespaces() ([]string, error) {
	list, err := s.coreClient.Namespaces().List(metav1.ListOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "while list Namespaces")
	}

	namespaces := make([]string, 0)

	if list == nil || len(list.Items) == 0 {
		return namespaces, nil
	}

	for _, namespace := range list.Items {
		if !s.IsExcludedNamespace(namespace.Name) && namespace.Status.Phase != corev1.NamespaceTerminating {
			namespaces = append(namespaces, namespace.Name)
		}
	}
	return namespaces, nil
}

func (s *NamespaceService) IsExcludedNamespace(namespace string) bool {
	for _, name := range s.config.ExcludedNamespaces {
		if name == namespace {
			return true
		}
	}
	return false
}
