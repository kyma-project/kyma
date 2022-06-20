package fluentbit

import (
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestValidateVariables(t *testing.T) {

	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Variables: []telemetryv1alpha1.VariableReference{
				{
					Name: "foo1",
					ValueFrom: telemetryv1alpha1.ValueFromType{SecretKey: telemetryv1alpha1.SecretKeyRef{
						Name:      "fooN",
						Namespace: "fooNs",
						Key:       "foo",
					},
					},
				},
				{
					Name: "foo2",
					ValueFrom: telemetryv1alpha1.ValueFromType{SecretKey: telemetryv1alpha1.SecretKeyRef{
						Name:      "fooN",
						Namespace: "fooNs",
						Key:       "foo",
					}},
				},
			},
		},
	}
	logPipeline.Name = "pipe1"
	logPipelines := &telemetryv1alpha1.LogPipelineList{
		Items: []telemetryv1alpha1.LogPipeline{*logPipeline},
	}

	newLogPipeline := &telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pipe2",
		},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Variables: []telemetryv1alpha1.VariableReference{{
				Name: "foo2",
				ValueFrom: telemetryv1alpha1.ValueFromType{SecretKey: telemetryv1alpha1.SecretKeyRef{
					Name:      "fooN",
					Namespace: "fooNs",
					Key:       "foo",
				}},
			}},
		},
	}

	err := ValidateVariables(newLogPipeline, logPipelines)
	require.Error(t, err)
}
