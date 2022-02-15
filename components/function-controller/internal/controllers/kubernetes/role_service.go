package kubernetes

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
)

type RoleService interface {
	IsBase(role *rbacv1.Role) bool
	ListBase(ctx context.Context) ([]rbacv1.Role, error)
	UpdateNamespace(ctx context.Context, logger logr.Logger, namespace string, baseInstance *rbacv1.Role) error
}

var _ RoleService = &roleService{}

type roleService struct {
	client resource.Client
	config Config
}

func NewRoleService(client resource.Client, config Config) RoleService {
	return &roleService{
		client: client,
		config: config,
	}
}

func (r *roleService) ListBase(ctx context.Context) ([]rbacv1.Role, error) {
	roles := &rbacv1.RoleList{}
	if err := r.client.ListByLabel(ctx, r.config.BaseNamespace, map[string]string{RbacLabel: RoleLabelValue}, roles); err != nil {
		return nil, err
	}

	return roles.Items, nil
}

func (r *roleService) IsBase(role *rbacv1.Role) bool {
	label, ok := role.Labels[RbacLabel]
	return role.Namespace == r.config.BaseNamespace && ok && label == RoleLabelValue
}

func (r *roleService) UpdateNamespace(ctx context.Context, logger logr.Logger, namespace string, baseInstance *rbacv1.Role) error {
	logger.Info(fmt.Sprintf("Updating Role '%s/%s'", namespace, baseInstance.GetName()))
	instance := &rbacv1.Role{}
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: baseInstance.GetName()}, instance); err != nil {
		if errors.IsNotFound(err) {
			return r.createRole(ctx, logger, namespace, baseInstance)
		}
		logger.Error(err, fmt.Sprintf("Gathering existing Role '%s/%s' failed", namespace, baseInstance.GetName()))
		return err
	}

	return r.updateRole(ctx, logger, instance, baseInstance)
}

func (r *roleService) createRole(ctx context.Context, logger logr.Logger, namespace string, baseInstance *rbacv1.Role) error {
	role := rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:        baseInstance.GetName(),
			Namespace:   namespace,
			Labels:      baseInstance.Labels,
			Annotations: baseInstance.Annotations,
		},
		Rules: baseInstance.Rules,
	}

	logger.Info(fmt.Sprintf("Creating Role '%s/%s'", role.GetNamespace(), role.GetName()))
	if err := r.client.Create(ctx, &role); err != nil {
		logger.Error(err, fmt.Sprintf("Creating Role '%s/%s' failed", role.GetNamespace(), role.GetName()))
		return err
	}

	return nil
}

func (r *roleService) updateRole(ctx context.Context, logger logr.Logger, instance, baseInstance *rbacv1.Role) error {
	copiedRole := instance.DeepCopy()
	copiedRole.Annotations = baseInstance.GetAnnotations()
	copiedRole.Labels = baseInstance.GetLabels()
	copiedRole.Rules = baseInstance.Rules

	if err := r.client.Update(ctx, copiedRole); err != nil {
		logger.Error(err, fmt.Sprintf("Updating Role '%s/%s' failed", copiedRole.GetNamespace(), copiedRole.GetName()))
		return err
	}

	return nil
}
