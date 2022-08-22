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
    add_prefix   __k8s__
    match        test-logpipeline.*
    name         nest
    nested_under kubernetes
    operation    lift

[FILTER]
    match      test-logpipeline.*
    name       record_modifier
    remove_key __k8s__annotations

[FILTER]
    match      test-logpipeline.*
    name       record_modifier
    remove_key __k8s__labels

[FILTER]
    match         test-logpipeline.*
    name          nest
    nest_under    kubernetes
    operation     nest
    remove_prefix __k8s__
    wildcard      __k8s__*

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
