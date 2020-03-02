package resource_watcher

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type NamespaceService struct {
	coreClient *v1.CoreV1Client
	baseNamespace string
	excludedNamespaces []string
}

func NewNamespaceService(coreClient *v1.CoreV1Client, config ResourceWatcherConfig) *NamespaceService {
	return &NamespaceService{
		coreClient: coreClient,
		baseNamespace: config.BaseNamespace,
		excludedNamespaces: config.ExcludedNamespaces,
	}
}

func (s *NamespaceService) GetNamespace(namespace string) (*corev1.Namespace, error) {
	ns, err := s.coreClient.Namespaces().Get(namespace, metav1.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while get Namespaces %s", namespace)
	}

	isExcluded := s.IsExcludedNamespace(ns.Name)
	if isExcluded {
		return nil, nil
	}

	return ns, nil
}

func (s *NamespaceService) GetNamespaces() ([]*corev1.Namespace, error) {
	list, err := s.coreClient.Namespaces().List(metav1.ListOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "while list Namespaces")
	}
	if list == nil || len(list.Items) == 0 {
		return nil, nil
	}

	namespaces := make([]*corev1.Namespace, 0)
	for _, namespace := range list.Items {
		if !s.IsExcludedNamespace(namespace.Name) {
			namespaces = append(namespaces, &namespace)
		}
	}
	return namespaces, nil
}

func (s *NamespaceService) IsExcludedNamespace(namespace string) bool {
	for _, name := range s.excludedNamespaces {
		if name == namespace {
			return true
		}
	}
	return false
}

func (s *NamespaceService) HasBaseNamespace(namespace string) bool {
	return s.baseNamespace == namespace
}