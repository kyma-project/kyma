package logpipeline

import (
	"bytes"
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

type secretHelper struct {
	client client.Client
}

func newSecretHelper(client client.Client) *secretHelper {
	return &secretHelper{
		client: client,
	}
}

func (s *secretHelper) ValidatePipelineSecretsExist(ctx context.Context, logpipeline *telemetryv1alpha1.LogPipeline) bool {
	for _, v := range logpipeline.Spec.Variables {
		if !v.ValueFrom.IsSecretRef() {
			return false
		}

		_, err := s.get(ctx, *v.ValueFrom.SecretRef)
		if err != nil {
			return false
		}
	}

	output := logpipeline.Spec.Output
	if !output.IsHTTPDefined() {
		return true
	}
	if output.HTTP.Host.ValueFrom != nil && output.HTTP.Host.ValueFrom.IsSecretRef() {
		_, err := s.get(ctx, *output.HTTP.Host.ValueFrom.SecretRef)
		if err != nil {
			return false
		}
	}
	if output.HTTP.User.ValueFrom != nil && output.HTTP.User.ValueFrom.IsSecretRef() {
		_, err := s.get(ctx, *output.HTTP.User.ValueFrom.SecretRef)
		if err != nil {
			return false
		}
	}
	if output.HTTP.Password.ValueFrom != nil && output.HTTP.Password.ValueFrom.IsSecretRef() {
		_, err := s.get(ctx, *output.HTTP.Password.ValueFrom.SecretRef)
		if err != nil {
			return false
		}
	}

	return true
}

func (s *secretHelper) CopySecretData(ctx context.Context, from telemetryv1alpha1.SecretRef, targetKey string, secretData map[string][]byte) error {
	log := logf.FromContext(ctx)
	var referencedSecret *corev1.Secret
	referencedSecret, err := s.get(ctx, from)
	if err != nil {
		log.Error(err, "unable to find secret")
		return err
	}
	// Check if any secret has been changed
	fetchedSecretData, err := GetSecretData(*referencedSecret, from, targetKey)
	if err != nil {
		log.Error(err, "unable to get secret data")
		return err
	}
	for k, v := range fetchedSecretData {
		secretData[k] = v
	}
	return nil
}

func (s *secretHelper) get(ctx context.Context, from telemetryv1alpha1.SecretRef) (*corev1.Secret, error) {
	log := logf.FromContext(ctx)

	var secret corev1.Secret
	if err := s.client.Get(ctx, types.NamespacedName{Name: from.Name, Namespace: from.Namespace}, &secret); err != nil {
		log.Error(err, fmt.Sprintf("Failed reading secret '%s' from namespace '%s'", from.Name, from.Namespace))
		return nil, err
	}
	if _, ok := secret.Data[from.Key]; !ok {
		return nil, fmt.Errorf("unable to find key '%s' in secret '%s'", from.Key, from.Name)
	}

	return &secret, nil
}

func SecretHasChanged(oldSecret, newSecret map[string][]byte) bool {
	if len(newSecret) != len(oldSecret) {
		return true
	}
	for k, newSecretVal := range newSecret {
		if oldSecretVal, ok := oldSecret[k]; !ok || !bytes.Equal(newSecretVal, oldSecretVal) {
			return true
		}
	}
	return false
}

func GetSecretData(secret corev1.Secret, from telemetryv1alpha1.SecretRef, targetKey string) (map[string][]byte, error) {
	data := make(map[string][]byte)
	if v, found := secret.Data[from.Key]; found {
		data[targetKey] = v
		return data, nil
	}
	return data, fmt.Errorf("the key '%s' cannot be found in the given secret '%s'", from.Key, secret.Name)
}
