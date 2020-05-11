package kubernetes

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
)

type SecretService interface {
	IsBase(secret *corev1.Secret) bool
	ListBase(ctx context.Context) ([]corev1.Secret, error)
	UpdateNamespace(ctx context.Context, logger logr.Logger, namespace string, baseInstance *corev1.Secret) error
}

var _ SecretService = &secretService{}

type secretService struct {
	client resource.Client
	config Config
}

func NewSecretService(client resource.Client, config Config) SecretService {
	return &secretService{
		client: client,
		config: config,
	}
}

func (r *secretService) ListBase(ctx context.Context) ([]corev1.Secret, error) {
	secrets := &corev1.SecretList{}
	if err := r.client.ListByLabel(ctx, r.config.BaseNamespace, map[string]string{ConfigLabel: CredentialsLabelValue}, secrets); err != nil {
		return nil, err
	}

	return secrets.Items, nil
}

func (r *secretService) IsBase(secret *corev1.Secret) bool {
	return secret.Namespace == r.config.BaseNamespace && secret.Labels[ConfigLabel] == CredentialsLabelValue
}

func (r *secretService) UpdateNamespace(ctx context.Context, logger logr.Logger, namespace string, baseInstance *corev1.Secret) error {
	logger.Info(fmt.Sprintf("Updating Secret '%s/%s'", namespace, baseInstance.GetName()))
	instance := &corev1.Secret{}
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: baseInstance.GetName()}, instance); err != nil {
		if errors.IsNotFound(err) {
			return r.createSecret(ctx, logger, namespace, baseInstance)
		}
		logger.Error(err, fmt.Sprintf("Gathering existing Secret '%s/%s' failed", namespace, baseInstance.GetName()))
		return err
	}

	return r.updateSecret(ctx, logger, instance, baseInstance)
}

func (r *secretService) createSecret(ctx context.Context, logger logr.Logger, namespace string, baseInstance *corev1.Secret) error {
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        baseInstance.GetName(),
			Namespace:   namespace,
			Labels:      baseInstance.Labels,
			Annotations: baseInstance.Annotations,
		},
		Data:       baseInstance.Data,
		StringData: baseInstance.StringData,
		Type:       baseInstance.Type,
	}

	logger.Info(fmt.Sprintf("Creating Secret '%s/%s'", secret.GetNamespace(), secret.GetName()))
	if err := r.client.Create(ctx, &secret); err != nil {
		logger.Error(err, fmt.Sprintf("Creating Secret '%s/%s' failed", secret.GetNamespace(), secret.GetName()))
		return err
	}

	return nil
}

func (r *secretService) updateSecret(ctx context.Context, logger logr.Logger, instance, baseInstance *corev1.Secret) error {
	copy := instance.DeepCopy()
	copy.Annotations = baseInstance.GetAnnotations()
	copy.Labels = baseInstance.GetLabels()
	copy.Data = baseInstance.Data
	copy.StringData = baseInstance.StringData
	copy.Type = baseInstance.Type

	if err := r.client.Update(ctx, copy); err != nil {
		logger.Error(err, fmt.Sprintf("Updating Secret '%s/%s' failed", copy.GetNamespace(), copy.GetName()))
		return err
	}

	return nil
}
