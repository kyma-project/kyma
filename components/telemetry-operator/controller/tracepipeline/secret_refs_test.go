package tracepipeline

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestFetchSecretValue(t *testing.T) {
	data := map[string][]byte{
		"myKey": []byte("myValue"),
	}
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-secret",
			Namespace: "default",
		},
		Data: data,
	}
	client := fake.NewClientBuilder().WithObjects(&secret).Build()

	value := telemetryv1alpha1.ValueType{
		ValueFrom: &telemetryv1alpha1.ValueFromSource{
			SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
				Name:      "my-secret",
				Namespace: "default",
				Key:       "myKey",
			},
		},
	}

	fetchedData, err := fetchSecretValue(ctx, client, value)

	require.Nil(t, err)
	require.Equal(t, string(fetchedData), "myValue")
}

func TestFetchValueFromNonExistingSecret(t *testing.T) {
	client := fake.NewClientBuilder().Build()

	value := telemetryv1alpha1.ValueType{
		ValueFrom: &telemetryv1alpha1.ValueFromSource{
			SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
				Name:      "my-secret",
				Namespace: "default",
				Key:       "myKey",
			},
		},
	}

	_, err := fetchSecretValue(ctx, client, value)
	require.Error(t, err)
}

func TestFetchValueFromNonExistingKey(t *testing.T) {
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-secret",
			Namespace: "default",
		},
	}
	client := fake.NewClientBuilder().WithObjects(&secret).Build()

	value := telemetryv1alpha1.ValueType{
		ValueFrom: &telemetryv1alpha1.ValueFromSource{
			SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
				Name:      "my-secret",
				Namespace: "default",
				Key:       "myKey",
			},
		},
	}

	_, err := fetchSecretValue(ctx, client, value)
	require.Error(t, err)
}

func TestFetchFromCr(t *testing.T) {
	client := fake.NewClientBuilder().Build()
	pipeline := telemetryv1alpha1.TracePipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pipeline",
		},
		Spec: telemetryv1alpha1.TracePipelineSpec{
			Output: telemetryv1alpha1.TracePipelineOutput{
				Otlp: &telemetryv1alpha1.OtlpOutput{
					Endpoint: telemetryv1alpha1.ValueType{
						Value: "endpoint",
					},
					Headers: []telemetryv1alpha1.Header{
						{
							Name: "Authorization",
							ValueType: telemetryv1alpha1.ValueType{
								Value: "Bearer xyz",
							},
						},
						{
							Name: "Test",
							ValueType: telemetryv1alpha1.ValueType{
								Value: "123",
							},
						},
					},
				},
			},
		},
	}

	data, err := fetchSecretData(ctx, client, pipeline.Spec.Output.Otlp)
	require.NoError(t, err)
	require.Contains(t, data, otlpEndpointVariable)
	require.Contains(t, data, "HEADER_AUTHORIZATION")
	require.Contains(t, data, "HEADER_TEST")
	require.NotContains(t, data, basicAuthHeaderVariable)
}

func TestFetchFromSecret(t *testing.T) {
	data := map[string][]byte{
		"user":     []byte("secret-username"),
		"password": []byte("secret-password"),
		"endpoint": []byte("secret-endpoint"),
		"token":    []byte("Bearer 123"),
		"test":     []byte("123"),
	}
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-secret",
			Namespace: "default",
		},
		Data: data,
	}
	client := fake.NewClientBuilder().WithObjects(&secret).Build()

	pipeline := telemetryv1alpha1.TracePipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pipeline",
		},
		Spec: telemetryv1alpha1.TracePipelineSpec{
			Output: telemetryv1alpha1.TracePipelineOutput{
				Otlp: &telemetryv1alpha1.OtlpOutput{
					Endpoint: telemetryv1alpha1.ValueType{
						ValueFrom: &telemetryv1alpha1.ValueFromSource{
							SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
								Name:      "my-secret",
								Namespace: "default",
								Key:       "endpoint",
							},
						},
					},
					Authentication: &telemetryv1alpha1.AuthenticationOptions{
						Basic: &telemetryv1alpha1.BasicAuthOptions{
							User: telemetryv1alpha1.ValueType{
								ValueFrom: &telemetryv1alpha1.ValueFromSource{
									SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
										Name:      "my-secret",
										Namespace: "default",
										Key:       "user",
									},
								},
							},
							Password: telemetryv1alpha1.ValueType{
								ValueFrom: &telemetryv1alpha1.ValueFromSource{
									SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
										Name:      "my-secret",
										Namespace: "default",
										Key:       "password",
									},
								},
							},
						},
					},
					Headers: []telemetryv1alpha1.Header{
						{
							Name: "Authorization",
							ValueType: telemetryv1alpha1.ValueType{
								ValueFrom: &telemetryv1alpha1.ValueFromSource{
									SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
										Name:      "my-secret",
										Namespace: "default",
										Key:       "token",
									},
								},
							},
						},
						{
							Name: "Test",
							ValueType: telemetryv1alpha1.ValueType{
								ValueFrom: &telemetryv1alpha1.ValueFromSource{
									SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
										Name:      "my-secret",
										Namespace: "default",
										Key:       "test",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	data, err := fetchSecretData(ctx, client, pipeline.Spec.Output.Otlp)
	require.NoError(t, err)
	require.Contains(t, data, otlpEndpointVariable)
	require.Equal(t, string(data[otlpEndpointVariable]), "secret-endpoint")
	require.Contains(t, data, basicAuthHeaderVariable)
	require.Contains(t, data, "HEADER_AUTHORIZATION")
	require.Contains(t, data, "HEADER_TEST")
	require.Equal(t, string(data[basicAuthHeaderVariable]), getBasicAuthHeader("secret-username", "secret-password"))
}

func TestFetchFromSecretWithMissingKey(t *testing.T) {
	data := map[string][]byte{
		"user":     []byte("secret-username"),
		"password": []byte("secret-password"),
	}
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-secret",
			Namespace: "default",
		},
		Data: data,
	}
	client := fake.NewClientBuilder().WithObjects(&secret).Build()

	pipeline := telemetryv1alpha1.TracePipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pipeline",
		},
		Spec: telemetryv1alpha1.TracePipelineSpec{
			Output: telemetryv1alpha1.TracePipelineOutput{
				Otlp: &telemetryv1alpha1.OtlpOutput{
					Endpoint: telemetryv1alpha1.ValueType{
						ValueFrom: &telemetryv1alpha1.ValueFromSource{
							SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
								Name:      "my-secret",
								Namespace: "default",
								Key:       "endpoint",
							},
						},
					},
					Authentication: &telemetryv1alpha1.AuthenticationOptions{
						Basic: &telemetryv1alpha1.BasicAuthOptions{
							User: telemetryv1alpha1.ValueType{
								ValueFrom: &telemetryv1alpha1.ValueFromSource{
									SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
										Name:      "my-secret",
										Namespace: "default",
										Key:       "user",
									},
								},
							},
							Password: telemetryv1alpha1.ValueType{
								ValueFrom: &telemetryv1alpha1.ValueFromSource{
									SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
										Name:      "my-secret",
										Namespace: "default",
										Key:       "password",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := fetchSecretData(ctx, client, pipeline.Spec.Output.Otlp)
	require.Error(t, err)
}

func TestFetchSecretDataFromNonExistingSecret(t *testing.T) {
	client := fake.NewClientBuilder().Build()
	pipeline := telemetryv1alpha1.TracePipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pipeline",
		},
		Spec: telemetryv1alpha1.TracePipelineSpec{
			Output: telemetryv1alpha1.TracePipelineOutput{
				Otlp: &telemetryv1alpha1.OtlpOutput{
					Endpoint: telemetryv1alpha1.ValueType{
						ValueFrom: &telemetryv1alpha1.ValueFromSource{
							SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
								Name:      "my-secret",
								Namespace: "default",
								Key:       "myKey",
							},
						},
					},
				},
			},
		},
	}

	_, err := fetchSecretData(ctx, client, pipeline.Spec.Output.Otlp)
	require.Error(t, err)
}
