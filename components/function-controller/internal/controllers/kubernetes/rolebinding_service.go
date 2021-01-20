package kubernetes

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
)

type RoleBindingService interface {
	IsBase(role *rbacv1.RoleBinding) bool
	ListBase(ctx context.Context) ([]rbacv1.RoleBinding, error)
	UpdateNamespace(ctx context.Context, logger logr.Logger, namespace string, baseInstance *rbacv1.RoleBinding) error
}

var _ RoleBindingService = &roleBindingService{}

type roleBindingService struct {
	client resource.Client
	config Config
}

func NewRoleBindingService(client resource.Client, config Config) RoleBindingService {
	return &roleBindingService{
		client: client,
		config: config,
	}
}

func (r *roleBindingService) ListBase(ctx context.Context) ([]rbacv1.RoleBinding, error) {
	rolebindings := &rbacv1.RoleBindingList{}
	if err := r.client.ListByLabel(ctx, r.config.BaseNamespace, map[string]string{RbacLabel: RoleBindingLabelValue}, rolebindings); err != nil {
		return nil, err
	}

	return rolebindings.Items, nil
}

func (r *roleBindingService) IsBase(roleBinding *rbacv1.RoleBinding) bool {
	label, ok := roleBinding.Labels[RbacLabel]
	return roleBinding.Namespace == r.config.BaseNamespace && ok && label == RoleBindingLabelValue
}

func (r *roleBindingService) UpdateNamespace(ctx context.Context, logger logr.Logger, namespace string, baseInstance *rbacv1.RoleBinding) error {
	logger.Info(fmt.Sprintf("Updating RoleBinding '%s/%s'", namespace, baseInstance.GetName()))
	instance := &rbacv1.RoleBinding{}
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: baseInstance.GetName()}, instance); err != nil {
		if apierrors.IsNotFound(err) {
			return r.createRoleBinding(ctx, logger, namespace, baseInstance)
		}
		logger.Error(err, fmt.Sprintf("Gathering existing RoleBinding '%s/%s' failed", namespace, baseInstance.GetName()))
		return err
	}

	return r.updateRoleBinding(ctx, logger, instance, baseInstance)
}

func (r *roleBindingService) createRoleBinding(ctx context.Context, logger logr.Logger, namespace string, baseInstance *rbacv1.RoleBinding) error {
	if len(baseInstance.Subjects) != 1 {
		return fmt.Errorf("base RoleBinding %s in namespace %s has %d subjects, expected 1, cannot create", baseInstance.GetName(), baseInstance.GetNamespace(), len(baseInstance.Subjects))
	}

	copiedSubject := baseInstance.Subjects[0].DeepCopy()
	copiedSubject.Namespace = namespace

	roleBinding := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:        baseInstance.GetName(),
			Namespace:   namespace,
			Labels:      baseInstance.Labels,
			Annotations: baseInstance.Annotations,
		},
		Subjects: []rbacv1.Subject{
			*copiedSubject,
		},
		RoleRef: baseInstance.RoleRef,
	}

	logger.Info(fmt.Sprintf("Creating RoleBinding '%s/%s'", roleBinding.GetNamespace(), roleBinding.GetName()))
	if err := r.client.Create(ctx, &roleBinding); err != nil {
		logger.Error(err, fmt.Sprintf("Creating RoleBinding '%s/%s' failed", roleBinding.GetNamespace(), roleBinding.GetName()))
		return err
	}

	return nil
}

func (r *roleBindingService) updateRoleBinding(ctx context.Context, logger logr.Logger, instance, baseInstance *rbacv1.RoleBinding) error {
	if len(baseInstance.Subjects) != 1 {
		return fmt.Errorf("base RoleBinding %s in namespace %s has %d subjects, expected 1, cannot update", baseInstance.GetName(), baseInstance.GetNamespace(), len(baseInstance.Subjects))
	}

	copiedRoleBinding := instance.DeepCopy()
	copiedRoleBinding.Annotations = baseInstance.GetAnnotations()
	copiedRoleBinding.Labels = baseInstance.GetLabels()
	copiedRoleBinding.RoleRef = baseInstance.RoleRef

	copiedSubject := baseInstance.Subjects[0].DeepCopy()
	copiedSubject.Namespace = instance.GetNamespace()
	copiedRoleBinding.Subjects = []rbacv1.Subject{*copiedSubject}

	if err := r.client.Update(ctx, copiedRoleBinding); err != nil {
		logger.Error(err, fmt.Sprintf("Updating RoleBinding '%s/%s' failed", copiedRoleBinding.GetNamespace(), copiedRoleBinding.GetName()))
		return err
	}

	return nil
}
