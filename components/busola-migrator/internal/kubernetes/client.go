package kubernetes

import (
	"context"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/model"

	"github.com/pkg/errors"
	apirbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	subjectKind  = "User"
	roleRefKind  = "ClusterRole"
	rbacAPIGroup = "rbac.authorization.k8s.io"
)

var rbacConfig = map[model.UserPermission]struct {
	clusterRoleBindingName string
	clusterRoleName        string
}{
	model.UserPermissionAdmin: {
		clusterRoleBindingName: "cluster-admin-users",
		clusterRoleName:        "kyma-admin",
	},
	model.UserPermissionDeveloper: {
		clusterRoleBindingName: "cluster-developer-users",
		clusterRoleName:        "kyma-essentials",
	},
}

//go:generate mockery --name=CRBInterface --output=automock --outpkg=automock --case=underscore
type CRBInterface interface {
	Create(ctx context.Context, clusterRoleBinding *apirbacv1.ClusterRoleBinding, opts metav1.CreateOptions) (*apirbacv1.ClusterRoleBinding, error)
	Update(ctx context.Context, clusterRoleBinding *apirbacv1.ClusterRoleBinding, opts metav1.UpdateOptions) (*apirbacv1.ClusterRoleBinding, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apirbacv1.ClusterRoleBinding, error)
}

type Client struct {
	clusterRoleBindingsClient CRBInterface
}

func New(kubeConfig *rest.Config) (Client, error) {
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return Client{}, errors.Wrap(err, "while creating Kubernetes clientset for kubeconfig")
	}

	return Client{
		clusterRoleBindingsClient: clientset.RbacV1().ClusterRoleBindings(),
	}, nil
}

func (c Client) EnsureUserPermissions(user model.User) error {
	if user.IsAdmin {
		if err := c.ensureUserPermission(user.Email, model.UserPermissionAdmin); err != nil {
			return errors.Wrapf(err, "while ensuring user %s has %s permission", user.Email, model.UserPermissionAdmin)
		}
	}
	if user.IsDeveloper {
		if err := c.ensureUserPermission(user.Email, model.UserPermissionDeveloper); err != nil {
			return errors.Wrapf(err, "while ensuring user %s has %s permission", user.Email, model.UserPermissionDeveloper)
		}
	}
	return nil
}

func (c Client) ensureUserPermission(userEmail string, permission model.UserPermission) error {
	crb, err := c.ensureClusterRoleBindingExists(rbacConfig[permission].clusterRoleBindingName, rbacConfig[permission].clusterRoleName)
	if err != nil {
		return errors.Wrapf(err, "while ensuring Cluster Role Binding with name %s exists", rbacConfig[permission].clusterRoleBindingName)
	}
	err = c.ensureUserIsSubjectOfClusterRoleBinding(userEmail, crb)
	if err != nil {
		return errors.Wrapf(err, "while ensuring user %s is a subject of Cluster Role Binding %s", userEmail, rbacConfig[permission].clusterRoleBindingName)
	}
	return nil
}

func (c Client) ensureClusterRoleBindingExists(crbName, crName string) (*apirbacv1.ClusterRoleBinding, error) {
	crb, err := c.clusterRoleBindingsClient.Get(context.Background(), crbName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return c.createClusterRoleBinding(crbName, crName)
		}
		return nil, errors.Wrapf(err, "while getting Cluster Role Binding %s", crbName)
	}

	return crb, nil
}

func (c Client) createClusterRoleBinding(crbName, crName string) (*apirbacv1.ClusterRoleBinding, error) {
	crb, err := c.clusterRoleBindingsClient.Create(context.Background(), buildClusterRoleBinding(crbName, crName), metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "while creating Cluster Role Binding %s", crbName)
	}
	return crb, nil
}

func (c Client) ensureUserIsSubjectOfClusterRoleBinding(userEmail string, crb *apirbacv1.ClusterRoleBinding) error {
	if crbContainsSubject(crb, userEmail) {
		return nil
	}
	crb.Subjects = append(crb.Subjects, buildSubject(userEmail))
	_, err := c.clusterRoleBindingsClient.Update(context.Background(), crb, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrapf(err, "while updating Cluster Role Binding %s", crb.Name)
	}
	return nil
}

func crbContainsSubject(crb *apirbacv1.ClusterRoleBinding, subjectName string) bool {
	for _, subject := range crb.Subjects {
		if subject.APIGroup == rbacAPIGroup && subject.Kind == subjectKind && subject.Name == subjectName {
			return true
		}
	}
	return false
}

func buildSubject(subjectName string) apirbacv1.Subject {
	return apirbacv1.Subject{
		Kind:     subjectKind,
		APIGroup: rbacAPIGroup,
		Name:     subjectName,
	}
}

func buildClusterRoleBinding(clusterRoleBindingName, clusterRoleName string) *apirbacv1.ClusterRoleBinding {
	return &apirbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
			Labels: map[string]string{
				"kyma-project.io/uaa": "migrated",
			},
		},
		RoleRef: apirbacv1.RoleRef{
			APIGroup: rbacAPIGroup,
			Kind:     roleRefKind,
			Name:     clusterRoleName,
		},
	}
}
