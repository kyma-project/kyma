package namespaces

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type NamespacesClientInterface interface {
	GetNamespace(name string) (*v1.Namespace, error)
	UpdateNamespace(namespace *v1.Namespace) (*v1.Namespace, error)
}

type NamespacesClient struct {
	Clientset *kubernetes.Clientset
}

func (nc *NamespacesClient) GetNamespace(name string) (*v1.Namespace, error) {
	return nc.Clientset.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
}

func (nc *NamespacesClient) UpdateNamespace(namespace *v1.Namespace) (*v1.Namespace, error) {
	return nc.Clientset.CoreV1().Namespaces().Update(namespace)
}
