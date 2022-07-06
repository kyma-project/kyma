package secret

import (
	"bytes"
	"context"
	"fmt"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
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
		_, err := s.FetchSecret(ctx, v.ValueFrom)
		if err != nil {
			return false
		}
	}
	return true
}

func (s *Helper) FetchSecret(ctx context.Context, fromType telemetryv1alpha1.ValueFromType) (*corev1.Secret, error) {
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
func IsSecretRef(fromType telemetryv1alpha1.ValueFromType) bool {
	return fromType.SecretKey.Name != "" && fromType.SecretKey.Key != ""
}

func CheckIfSecretHasChanged(newSecret, oldSecret map[string][]byte) bool {
	if len(newSecret) != len(oldSecret) {
		return true
	}
	for k, newSecretVal := range newSecret {
		if oldSecretVal, ok := oldSecret[k]; !ok || bytes.Compare(newSecretVal, oldSecretVal) != 0 {
			return true
		}
	}
	return false
}

func FetchSecretData(referencedSecret corev1.Secret, secretRef telemetryv1alpha1.VariableReference) (map[string][]byte, error) {
	valFrom := secretRef.ValueFrom
	data := make(map[string][]byte)
	for k, v := range referencedSecret.Data {
		if k == valFrom.SecretKey.Key {
			data[secretRef.Name] = v
			return data, nil
		}
	}
	return data, fmt.Errorf("the key '%s' cannot be found in the given secret '%s'", valFrom.SecretKey.Key, referencedSecret.Name)
}
