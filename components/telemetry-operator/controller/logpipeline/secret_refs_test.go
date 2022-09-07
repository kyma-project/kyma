package logpipeline

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLookupSecretRefFields(t *testing.T) {
	tests := []struct {
		name     string
		given    telemetryv1alpha1.LogPipeline
		expected []fieldDescriptor
	}{
		{
			name: "only variables",
			given: telemetryv1alpha1.LogPipeline{
				Spec: telemetryv1alpha1.LogPipelineSpec{
					Variables: []telemetryv1alpha1.VariableRef{
						{
							Name: "password-1",
							ValueFrom: telemetryv1alpha1.ValueFromSource{
								SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{Name: "secret-1", Key: "password"},
							},
						},
						{
							Name: "password-2",
							ValueFrom: telemetryv1alpha1.ValueFromSource{
								SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{Name: "secret-2", Key: "password"},
							},
						},
					},
				},
			},

			expected: []fieldDescriptor{
				{
					secretKeyRef:    telemetryv1alpha1.SecretKeyRef{Name: "secret-1", Key: "password"},
					targetSecretKey: "password-1",
				},
				{
					secretKeyRef:    telemetryv1alpha1.SecretKeyRef{Name: "secret-2", Key: "password"},
					targetSecretKey: "password-2",
				},
			},
		},
		{
			name: "http output secret refs",
			given: telemetryv1alpha1.LogPipeline{
				ObjectMeta: v1.ObjectMeta{
					Name: "cls",
				},
				Spec: telemetryv1alpha1.LogPipelineSpec{
					Output: telemetryv1alpha1.Output{
						HTTP: &telemetryv1alpha1.HTTPOutput{
							Host: telemetryv1alpha1.ValueType{
								ValueFrom: &telemetryv1alpha1.ValueFromSource{
									SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
										Name: "creds", Namespace: "default", Key: "host",
									},
								},
							},
							User: telemetryv1alpha1.ValueType{
								ValueFrom: &telemetryv1alpha1.ValueFromSource{
									SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
										Name: "creds", Namespace: "default", Key: "user",
									},
								},
							},
							Password: telemetryv1alpha1.ValueType{
								ValueFrom: &telemetryv1alpha1.ValueFromSource{
									SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
										Name: "creds", Namespace: "default", Key: "password",
									},
								},
							},
						},
					},
				},
			},
			expected: []fieldDescriptor{
				{
					secretKeyRef:    telemetryv1alpha1.SecretKeyRef{Name: "creds", Namespace: "default", Key: "host"},
					targetSecretKey: "CLS_DEFAULT_CREDS_HOST",
				},
				{
					secretKeyRef:    telemetryv1alpha1.SecretKeyRef{Name: "creds", Namespace: "default", Key: "user"},
					targetSecretKey: "CLS_DEFAULT_CREDS_USER",
				},
				{
					secretKeyRef:    telemetryv1alpha1.SecretKeyRef{Name: "creds", Namespace: "default", Key: "password"},
					targetSecretKey: "CLS_DEFAULT_CREDS_PASSWORD",
				},
			},
		},
		{
			name: "loki output secret refs",
			given: telemetryv1alpha1.LogPipeline{
				ObjectMeta: v1.ObjectMeta{
					Name: "loki",
				},
				Spec: telemetryv1alpha1.LogPipelineSpec{
					Output: telemetryv1alpha1.Output{
						Loki: &telemetryv1alpha1.LokiOutput{
							URL: telemetryv1alpha1.ValueType{
								ValueFrom: &telemetryv1alpha1.ValueFromSource{
									SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
										Name: "creds", Namespace: "default", Key: "url",
									},
								},
							},
						},
					},
				},
			},
			expected: []fieldDescriptor{
				{
					secretKeyRef:    telemetryv1alpha1.SecretKeyRef{Name: "creds", Namespace: "default", Key: "url"},
					targetSecretKey: "LOKI_DEFAULT_CREDS_URL",
				},
			},
		},
		{
			name: "output secret refs and variables",
			given: telemetryv1alpha1.LogPipeline{
				ObjectMeta: v1.ObjectMeta{
					Name: "loki",
				},
				Spec: telemetryv1alpha1.LogPipelineSpec{
					Output: telemetryv1alpha1.Output{
						Loki: &telemetryv1alpha1.LokiOutput{
							URL: telemetryv1alpha1.ValueType{
								ValueFrom: &telemetryv1alpha1.ValueFromSource{
									SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
										Name: "creds", Namespace: "default", Key: "url",
									},
								},
							},
						},
					},
					Variables: []telemetryv1alpha1.VariableRef{
						{
							Name: "password-1",
							ValueFrom: telemetryv1alpha1.ValueFromSource{
								SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{Name: "secret-1", Key: "password"},
							},
						},
						{
							Name: "password-2",
							ValueFrom: telemetryv1alpha1.ValueFromSource{
								SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{Name: "secret-2", Key: "password"},
							},
						},
					},
				},
			},
			expected: []fieldDescriptor{
				{
					secretKeyRef:    telemetryv1alpha1.SecretKeyRef{Name: "creds", Namespace: "default", Key: "url"},
					targetSecretKey: "LOKI_DEFAULT_CREDS_URL",
				},
				{
					secretKeyRef:    telemetryv1alpha1.SecretKeyRef{Name: "secret-1", Key: "password"},
					targetSecretKey: "password-1",
				},
				{
					secretKeyRef:    telemetryv1alpha1.SecretKeyRef{Name: "secret-2", Key: "password"},
					targetSecretKey: "password-2",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := lookupSecretRefFields(&test.given)
			require.ElementsMatch(t, test.expected, actual)
		})
	}
}

func TestHasSecretRef(t *testing.T) {
	pipeline := telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				Loki: &telemetryv1alpha1.LokiOutput{
					URL: telemetryv1alpha1.ValueType{
						ValueFrom: &telemetryv1alpha1.ValueFromSource{
							SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{
								Name: "creds", Namespace: "default", Key: "url",
							},
						},
					},
				},
			},
			Variables: []telemetryv1alpha1.VariableRef{
				{
					Name: "password-1",
					ValueFrom: telemetryv1alpha1.ValueFromSource{
						SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{Name: "secret-1", Namespace: "default", Key: "password"},
					},
				},
				{
					Name: "password-2",
					ValueFrom: telemetryv1alpha1.ValueFromSource{
						SecretKeyRef: &telemetryv1alpha1.SecretKeyRef{Name: "secret-2", Namespace: "default", Key: "password"},
					},
				},
			},
		},
	}

	require.True(t, hasSecretRef(&pipeline, "secret-1", "default"))
	require.True(t, hasSecretRef(&pipeline, "secret-2", "default"))
	require.True(t, hasSecretRef(&pipeline, "creds", "default"))

	require.False(t, hasSecretRef(&pipeline, "secret-1", "kube-system"))
	require.False(t, hasSecretRef(&pipeline, "unknown", "default"))
}
