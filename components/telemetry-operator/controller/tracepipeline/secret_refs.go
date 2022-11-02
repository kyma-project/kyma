package tracepipeline

import (
	"context"
	"encoding/base64"
	"fmt"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func fetchSecretData(ctx context.Context, c client.Reader, output *telemetryv1alpha1.OtlpOutput) (map[string][]byte, error) {
	secretData := map[string][]byte{}

	if output.Authentication != nil && output.Authentication.Basic.IsDefined() {
		username, err := fetchSecretValue(ctx, c, output.Authentication.Basic.User)
		if err != nil {
			return nil, err
		}
		password, err := fetchSecretValue(ctx, c, output.Authentication.Basic.Password)
		if err != nil {
			return nil, err
		}
		basicAuthHeader := getBasicAuthHeader(string(username), string(password))
		secretData[basicAuthHeaderVariable] = []byte(basicAuthHeader)
	}

	endpoint, err := fetchSecretValue(ctx, c, output.Endpoint)
	if err != nil {
		return nil, err
	}
	secretData[otlpEndpointVariable] = endpoint

	return secretData, nil
}

func fetchSecretValue(ctx context.Context, c client.Reader, value telemetryv1alpha1.ValueType) ([]byte, error) {
	if value.Value != "" {
		return []byte(value.Value), nil
	}
	if value.ValueFrom.IsSecretKeyRef() {
		lookupKey := types.NamespacedName{
			Name:      value.ValueFrom.SecretKeyRef.Name,
			Namespace: value.ValueFrom.SecretKeyRef.Namespace,
		}

		var secret corev1.Secret
		if err := c.Get(ctx, lookupKey, &secret); err != nil {
			return nil, err
		}

		if secretValue, found := secret.Data[value.ValueFrom.SecretKeyRef.Key]; found {
			return secretValue, nil
		}
		return nil, fmt.Errorf("referenced key not found in Secret")
	}

	return nil, fmt.Errorf("either value or secretReference have to be defined")
}

func getBasicAuthHeader(username string, password string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
}
