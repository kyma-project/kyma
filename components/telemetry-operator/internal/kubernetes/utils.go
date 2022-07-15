package kubernetes

import (
	"context"
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fs"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Utils struct {
	client client.Client
}

func NewUtils(client client.Client) *Utils {
	return &Utils{
		client: client,
	}
}

func (u *Utils) GetOrCreateConfigMap(ctx context.Context, name types.NamespacedName) (corev1.ConfigMap, error) {
	cm := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: name.Name, Namespace: name.Namespace}}
	err := u.GetOrCreate(ctx, &cm)
	if err != nil {
		return corev1.ConfigMap{}, err
	}
	return cm, nil
}

func (u *Utils) GetOrCreateSecret(ctx context.Context, name types.NamespacedName) (corev1.Secret, error) {
	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name.Name, Namespace: name.Namespace}}
	err := u.GetOrCreate(ctx, &secret)
	if err != nil {
		return corev1.Secret{}, err
	}
	return secret, nil
}

// Gets or creates the given obj in the Kubernetes cluster. obj must be a struct pointer so that obj can be updated with the content returned by the Server.
func (u *Utils) GetOrCreate(ctx context.Context, obj client.Object) error {
	err := u.client.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	if err != nil && errors.IsNotFound(err) {
		return u.client.Create(ctx, obj)
	}
	return err
}

func (u *Utils) GetFluentBitConfig(ctx context.Context, currentBaseDirectory, fluentBitParsersConfigMapKey string, fluentBitConfigMap types.NamespacedName, pipelineConfig fluentbit.PipelineConfig) ([]fs.File, error) {
	var configFiles []fs.File
	fluentBitSectionsConfigDirectory := currentBaseDirectory + "/dynamic"
	fluentBitParsersConfigDirectory := currentBaseDirectory + "/dynamic-parsers"
	fluentBitFilesDirectory := currentBaseDirectory + "/files"

	fmt.Printf("I am here")
	var generalCm corev1.ConfigMap
	if err := u.client.Get(ctx, fluentBitConfigMap, &generalCm); err != nil {
		return nil, err
	}
	for key, data := range generalCm.Data {
		configFiles = append(configFiles, fs.File{
			Path: currentBaseDirectory,
			Name: key,
			Data: data,
		})
	}

	var logPipelines *telemetryv1alpha1.LogPipelineList
	err := u.client.List(ctx, logPipelines)
	if err != nil {
		return []fs.File{}, err
	}

	for _, logPipeline := range logPipelines.Items {
		for _, file := range logPipeline.Spec.Files {
			configFiles = append(configFiles, fs.File{
				Path: fluentBitFilesDirectory,
				Name: file.Name,
				Data: file.Content,
			})
		}

		sectionsConfig, err := fluentbit.MergeSectionsConfig(&logPipeline, pipelineConfig)
		if err != nil {
			return []fs.File{}, err
		}
		configFiles = append(configFiles, fs.File{
			Path: fluentBitSectionsConfigDirectory,
			Name: logPipeline.Name + ".conf",
			Data: sectionsConfig,
		})
	}

	var logparsers telemetryv1alpha1.LogParserList
	if err := u.client.List(ctx, &logparsers); err != nil {
		return nil, err
	}
	parsersConfig := fluentbit.MergeParsersConfig(&logparsers)
	configFiles = append(configFiles, fs.File{
		Path: fluentBitParsersConfigDirectory,
		Name: fluentBitParsersConfigMapKey,
		Data: parsersConfig,
	})

	return configFiles, nil
}
