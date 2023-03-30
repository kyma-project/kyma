package webhook

import (
	"encoding/json"
	"testing"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var one int32 = 1
var two int32 = 2

func ValidV1Alpha2Function() serverlessv1alpha2.Function {
	f := serverlessv1alpha2.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       serverlessv1alpha2.FunctionKind,
			APIVersion: serverlessv1alpha2.FunctionApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testfunc",
			Namespace: "default",
		},
		Spec: serverlessv1alpha2.FunctionSpec{
			ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
				Build: &serverlessv1alpha2.ResourceRequirements{
					Resources: &corev1.ResourceRequirements{
						Limits: map[corev1.ResourceName]resource.Quantity{
							"cpu":    resource.MustParse("700m"),
							"memory": resource.MustParse("700Mi"),
						},
						Requests: map[corev1.ResourceName]resource.Quantity{
							"cpu":    resource.MustParse("200m"),
							"memory": resource.MustParse("200Mi"),
						},
					},
				},
				Function: &serverlessv1alpha2.ResourceRequirements{
					Resources: &corev1.ResourceRequirements{
						Limits: map[corev1.ResourceName]resource.Quantity{
							"cpu":    resource.MustParse("400m"),
							"memory": resource.MustParse("512Mi"),
						},
						Requests: map[corev1.ResourceName]resource.Quantity{
							"cpu":    resource.MustParse("200m"),
							"memory": resource.MustParse("256Mi"),
						},
					},
				},
			},
			Source: serverlessv1alpha2.Source{
				Inline: &serverlessv1alpha2.InlineSource{
					Source: `def main(event, context):\n  return \"hello world\"\n`,
				}},
			ScaleConfig: &serverlessv1alpha2.ScaleConfig{
				MinReplicas: &one,
				MaxReplicas: &two,
			},
			Runtime: serverlessv1alpha2.Python39,
		}}
	return f
}

func Marshall(t *testing.T, obj interface{}) string {
	out, err := json.Marshal(obj)
	require.NoError(t, err)
	return string(out)
}
