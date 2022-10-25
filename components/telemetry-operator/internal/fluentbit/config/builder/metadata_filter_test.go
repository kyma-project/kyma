package builder

import (
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/require"
)

func TestCreateKubernetesMetadataFilterKeepAll(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "test-logpipeline"},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Input: telemetryv1alpha1.Input{
				Application: telemetryv1alpha1.ApplicationInput{
					KeepAnnotations: true,
					DropLabels:      false}}}}

	actual := createKubernetesMetadataFilter(logPipeline)
	require.Equal(t, "", actual)
}

func TestCreateKubernetesMetadataFilterDropAll(t *testing.T) {
	expected := `[FILTER]
    name         nest
    match        test-logpipeline.*
    add_prefix   __kyma__
    nested_under kubernetes
    operation    lift

[FILTER]
    name       record_modifier
    match      test-logpipeline.*
    remove_key __kyma__annotations

[FILTER]
    name       record_modifier
    match      test-logpipeline.*
    remove_key __kyma__labels

[FILTER]
    name          nest
    match         test-logpipeline.*
    nest_under    kubernetes
    operation     nest
    remove_prefix __kyma__
    wildcard      __kyma__*

`
	logPipeline := &telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "test-logpipeline"},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Input: telemetryv1alpha1.Input{
				Application: telemetryv1alpha1.ApplicationInput{
					KeepAnnotations: false,
					DropLabels:      true}}}}

	actual := createKubernetesMetadataFilter(logPipeline)
	require.Equal(t, expected, actual)
}
