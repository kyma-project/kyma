package sync

import (
	"bytes"
	"context"
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type Helper struct {
	client client.Client
}

func NewSecretHelper(client client.Client) *Helper {
	return &Helper{
		client: client,
	}
}

func (s *Helper) ValidateSecretsExist(ctx context.Context, logpipeline *telemetryv1alpha1.LogPipeline) bool {
	for _, v := range logpipeline.Spec.Variables {
		_, err := s.fetchSecret(ctx, v.ValueFrom)
		if err != nil {
			return false
		}
	}
	if logpipeline.Spec.Output.HTTP.Host.ValueFrom.IsSecretRef() {
		_, err := s.fetchSecret(ctx, logpipeline.Spec.Output.HTTP.Host.ValueFrom)
		if err != nil {
			return false
		}
	}
	if logpipeline.Spec.Output.HTTP.User.ValueFrom.IsSecretRef() {
		_, err := s.fetchSecret(ctx, logpipeline.Spec.Output.HTTP.User.ValueFrom)
		if err != nil {
			return false
		}
	}
	if logpipeline.Spec.Output.HTTP.Password.ValueFrom.IsSecretRef() {
		_, err := s.fetchSecret(ctx, logpipeline.Spec.Output.HTTP.Password.ValueFrom)
		if err != nil {
			return false
		}
	}
	return true
}

func (s *Helper) fetchSecret(ctx context.Context, fromType telemetryv1alpha1.ValueFromType) (*corev1.Secret, error) {
	log := logf.FromContext(ctx)

	secretKey := fromType.SecretKey
	var referencedSecret corev1.Secret
	if err := s.client.Get(ctx, types.NamespacedName{Name: secretKey.Name, Namespace: secretKey.Namespace}, &referencedSecret); err != nil {
		log.Error(err, "Failed reading secret '%s' from namespace '%s'", secretKey.Name, secretKey.Namespace)
		return nil, err
	}
	if _, ok := referencedSecret.Data[secretKey.Key]; !ok {
		return nil, fmt.Errorf("unable to find key '%s' in secret '%s'", secretKey.Key, secretKey.Name)
	}

	return &referencedSecret, nil
}

func (s *Helper) copySecretData(ctx context.Context, valueFrom telemetryv1alpha1.ValueFromType, targetKey string, secretData map[string][]byte) error {
	log := logf.FromContext(ctx)
	var referencedSecret *corev1.Secret
	referencedSecret, err := s.fetchSecret(ctx, valueFrom)
	if err != nil {
		log.Error(err, "unable to find secret")
		return err
	}
	// Check if any secret has been changed
	fetchedSecretData, err := fetchSecretData(*referencedSecret, valueFrom, targetKey)
	if err != nil {
		log.Error(err, "unable to fetch secret data")
		return err
	}
	for k, v := range fetchedSecretData {
		secretData[k] = v
	}
	return nil
}

func checkIfSecretHasChanged(newSecret, oldSecret map[string][]byte) bool {
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

func fetchSecretData(referencedSecret corev1.Secret, valFrom telemetryv1alpha1.ValueFromType, targetKey string) (map[string][]byte, error) {
	data := make(map[string][]byte)
	if v, found := referencedSecret.Data[valFrom.SecretKey.Key]; found {
		data[targetKey] = v
		return data, nil
	}
	return data, fmt.Errorf("the key '%s' cannot be found in the given secret '%s'", valFrom.SecretKey.Key, referencedSecret.Name)
}
