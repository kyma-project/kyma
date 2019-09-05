package roles

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type RolesClientInterface interface {
	GetRole(name string, namespace string) (*rbacv1.Role, error)
	GetList(namespace string, opts metav1.ListOptions) (*rbacv1.RoleList, error)
	CreateRole(role *rbacv1.Role, namespace string) (*rbacv1.Role, error)
	DeleteRole(name string, namespace string) error
}

type RolesClient struct {
	Clientset *kubernetes.Clientset
}

func (rc *RolesClient) GetRole(name string, namespace string) (*rbacv1.Role, error) {
	return rc.Clientset.Rbac().Roles(namespace).Get(name, metav1.GetOptions{})
}

func (rc *RolesClient) GetList(namespace string, opts metav1.ListOptions) (*rbacv1.RoleList, error) {
	return rc.Clientset.Rbac().Roles(namespace).List(opts)
}

func (rc *RolesClient) CreateRole(role *rbacv1.Role, namespace string) (*rbacv1.Role, error) {
	return rc.Clientset.Rbac().Roles(namespace).Create(role)
}

func (rc *RolesClient) DeleteRole(name string, namespace string) error {
	return rc.Clientset.Rbac().Roles(namespace).Delete(name, nil)
}
