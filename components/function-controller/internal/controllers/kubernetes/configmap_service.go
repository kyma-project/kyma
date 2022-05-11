package kubernetes

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
)

type ConfigMapService interface {
	IsBase(configMap *corev1.ConfigMap) bool
	ListBase(ctx context.Context) ([]corev1.ConfigMap, error)
	UpdateNamespace(ctx context.Context, logger *zap.SugaredLogger, namespace string, baseInstance *corev1.ConfigMap) error
}

var _ ConfigMapService = &configMapService{}

type configMapService struct {
	client resource.Client
	config Config
}

func NewConfigMapService(client resource.Client, config Config) ConfigMapService {
	return &configMapService{
		client: client,
		config: config,
	}
}

func (r *configMapService) ListBase(ctx context.Context) ([]corev1.ConfigMap, error) {
	configMaps := corev1.ConfigMapList{}
	if err := r.client.ListByLabel(ctx, r.config.BaseNamespace, map[string]string{ConfigLabel: RuntimeLabelValue}, &configMaps); err != nil {
		return nil, err
	}

	return configMaps.Items, nil
}

func (r *configMapService) IsBase(configMap *corev1.ConfigMap) bool {
	return configMap.Namespace == r.config.BaseNamespace && configMap.Labels[ConfigLabel] == RuntimeLabelValue
}

func (r *configMapService) UpdateNamespace(ctx context.Context, logger *zap.SugaredLogger, namespace string, baseInstance *corev1.ConfigMap) error {
	logger.Info(fmt.Sprintf("Updating ConfigMap '%s/%s'", namespace, baseInstance.GetName()))
	instance := &corev1.ConfigMap{}
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: baseInstance.GetName()}, instance); err != nil {
		if errors.IsNotFound(err) {
			return r.createConfigMap(ctx, logger, namespace, baseInstance)
		}
		logger.Error(err, fmt.Sprintf("Gathering existing ConfigMap '%s/%s' failed", namespace, baseInstance.GetName()))
		return err
	}

	return r.updateConfigMap(ctx, logger, instance, baseInstance)
}

func (r *configMapService) createConfigMap(ctx context.Context, logger *zap.SugaredLogger, namespace string, baseInstance *corev1.ConfigMap) error {
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        baseInstance.GetName(),
			Namespace:   namespace,
			Labels:      baseInstance.Labels,
			Annotations: baseInstance.Annotations,
		},
		Data:       baseInstance.Data,
		BinaryData: baseInstance.BinaryData,
	}

	logger.Info(fmt.Sprintf("Creating ConfigMap '%s/%s'", configMap.GetNamespace(), configMap.GetName()))
	if err := r.client.Create(ctx, &configMap); err != nil {
		logger.Error(err, fmt.Sprintf("Creating ConfigMap '%s/%s' failed", configMap.GetNamespace(), configMap.GetName()))
		return err
	}

	return nil
}

func (r *configMapService) updateConfigMap(ctx context.Context, logger *zap.SugaredLogger, instance, baseInstance *corev1.ConfigMap) error {
	copy := instance.DeepCopy()
	copy.Annotations = baseInstance.GetAnnotations()
	copy.Labels = baseInstance.GetLabels()
	copy.Data = baseInstance.Data
	copy.BinaryData = baseInstance.BinaryData

	if err := r.client.Update(ctx, copy); err != nil {
		logger.Error(err, fmt.Sprintf("Updating ConfigMap '%s/%s' failed", copy.GetNamespace(), copy.GetName()))
		return err
	}

	return nil
}
