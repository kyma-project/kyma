package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
)

type ServiceAccountService interface {
	IsBase(serviceAccount *corev1.ServiceAccount) bool
	ListBase(ctx context.Context) ([]corev1.ServiceAccount, error)
	UpdateNamespace(ctx context.Context, logger *zap.SugaredLogger, namespace string, baseInstance *corev1.ServiceAccount) error
}

type serviceAccountService struct {
	client resource.Client
	config Config
}

func NewServiceAccountService(client resource.Client, config Config) ServiceAccountService {
	return &serviceAccountService{
		client: client,
		config: config,
	}
}

func (r *serviceAccountService) ListBase(ctx context.Context) ([]corev1.ServiceAccount, error) {
	serviceAccounts := &corev1.ServiceAccountList{}
	if err := r.client.ListByLabel(ctx, r.config.BaseNamespace, map[string]string{ConfigLabel: ServiceAccountLabelValue}, serviceAccounts); err != nil {
		return nil, err
	}

	return serviceAccounts.Items, nil
}

func (r *serviceAccountService) IsBase(serviceAccount *corev1.ServiceAccount) bool {
	return serviceAccount.Namespace == r.config.BaseNamespace && serviceAccount.Labels[ConfigLabel] == ServiceAccountLabelValue
}

func (r *serviceAccountService) UpdateNamespace(ctx context.Context, logger *zap.SugaredLogger, namespace string, baseInstance *corev1.ServiceAccount) error {
	logger.Info(fmt.Sprintf("Updating ServiceAccount '%s/%s'", namespace, baseInstance.GetName()))
	serviceAccount := &corev1.ServiceAccount{}
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: baseInstance.GetName()}, serviceAccount); err != nil {
		if errors.IsNotFound(err) {
			return r.createServiceAccount(ctx, logger, namespace, baseInstance)
		}
		logger.Error(err, fmt.Sprintf("Gathering existing ServiceAccount '%s/%s' failed", namespace, baseInstance.GetName()))
		return err
	}

	return r.updateServiceAccount(ctx, logger, serviceAccount, baseInstance)
}

func (r *serviceAccountService) createServiceAccount(ctx context.Context, logger *zap.SugaredLogger, namespace string, baseInstance *corev1.ServiceAccount) error {
	secrets := r.shiftSecretTokens(baseInstance)
	serviceAccount := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        baseInstance.GetName(),
			Namespace:   namespace,
			Labels:      baseInstance.Labels,
			Annotations: baseInstance.Annotations,
		},
		Secrets:                      secrets,
		ImagePullSecrets:             baseInstance.ImagePullSecrets,
		AutomountServiceAccountToken: baseInstance.AutomountServiceAccountToken,
	}

	logger.Info(fmt.Sprintf("Creating ServiceAccount '%s/%s'", serviceAccount.GetNamespace(), serviceAccount.GetName()))
	if err := r.client.Create(ctx, &serviceAccount); err != nil {
		logger.Error(err, fmt.Sprintf("Creating ServiceAccount '%s/%s'", serviceAccount.GetNamespace(), serviceAccount.GetName()))
		return err
	}

	return nil
}

func (r *serviceAccountService) updateServiceAccount(ctx context.Context, logger *zap.SugaredLogger, instance, baseInstance *corev1.ServiceAccount) error {
	tokens := r.extractSecretTokens(instance)
	secrets := r.shiftSecretTokens(baseInstance)
	secrets = append(secrets, tokens...)

	copy := instance.DeepCopy()
	copy.Annotations = baseInstance.GetAnnotations()
	copy.Labels = baseInstance.GetLabels()
	copy.ImagePullSecrets = baseInstance.ImagePullSecrets
	copy.AutomountServiceAccountToken = baseInstance.AutomountServiceAccountToken
	copy.Secrets = secrets

	if err := r.client.Update(ctx, copy); err != nil {
		logger.Error(err, fmt.Sprintf("Updating ServiceAccount '%s/%s' failed", copy.GetNamespace(), copy.GetName()))
		return err
	}

	return nil
}

func (*serviceAccountService) shiftSecretTokens(baseInstance *corev1.ServiceAccount) []corev1.ObjectReference {
	prefix := fmt.Sprintf("%s-token", baseInstance.Name)

	secrets := make([]corev1.ObjectReference, 0)
	for _, secret := range baseInstance.Secrets {
		if !strings.HasPrefix(secret.Name, prefix) {
			secrets = append(secrets, secret)
		}
	}

	return secrets
}

func (*serviceAccountService) extractSecretTokens(serviceAccount *corev1.ServiceAccount) []corev1.ObjectReference {
	prefix := fmt.Sprintf("%s-token", serviceAccount.Name)

	secrets := make([]corev1.ObjectReference, 0)
	for _, secret := range serviceAccount.Secrets {
		if strings.HasPrefix(secret.Name, prefix) {
			secrets = append(secrets, secret)
		}
	}

	return secrets
}
