package builder

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestCreateRecordModifierFilter(t *testing.T) {
	expected := `[FILTER]
    match  foo.*
    name   record_modifier
    record cluster_identifier ${KUBERNETES_SERVICE_HOST}

`
	logPipeline := &telemetryv1alpha1.LogPipeline{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}

	actual := createRecordModifierFilter(logPipeline)
	require.Equal(t, expected, actual, "Fluent Bit Permanent parser config is invalid")
}

func TestCreateLuaDedotFilterWithDefinedHostAndDedotSet(t *testing.T) {
	expected := `[FILTER]
    call   kubernetes_map_keys
    match  foo.*
    name   lua
    script /fluent-bit/scripts/filter-script.lua

`
	logPipeline := &telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{Name: "foo"},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				HTTP: telemetryv1alpha1.HTTPOutput{
					Dedot: true,
					Host: telemetryv1alpha1.ValueType{
						Value: "localhost"}}}}}

	actual := createLuaDedotFilter(logPipeline)
	require.Equal(t, expected, actual)
}

func TestCreateLuaDedotFilterWithUndefinedHost(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				HTTP: telemetryv1alpha1.HTTPOutput{
					Dedot: true}}}}

	actual := createLuaDedotFilter(logPipeline)
	require.Equal(t, "", actual)
}

func TestCreateLuaDedotFilterWithDedotFalse(t *testing.T) {
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				HTTP: telemetryv1alpha1.HTTPOutput{
					Dedot: false,
					Host: telemetryv1alpha1.ValueType{
						Value: "localhost"}}}}}

	actual := createLuaDedotFilter(logPipeline)
	require.Equal(t, "", actual)
}

func TestMergeSectionsConfig(t *testing.T) {
	expected := `[FILTER]
    Emitter_Mem_Buf_Limit 10M
    Emitter_Name          foo
    Emitter_Storage.type  filesystem
    Match                 kube.*
    Name                  rewrite_tag
    Rule                  $kubernetes['namespace_name'] "^(?!kyma-system$|kyma-integration$|kube-system$|istio-system$).*" foo.$TAG true

[FILTER]
    match  foo.*
    name   record_modifier
    record cluster_identifier ${KUBERNETES_SERVICE_HOST}

[FILTER]
    match foo.*
    name  grep
    regex log aa

[FILTER]
    call   kubernetes_map_keys
    match  foo.*
    name   lua
    script /fluent-bit/scripts/filter-script.lua

[OUTPUT]
    allow_duplicated_headers true
    format                   json
    host                     localhost
    match                    foo.*
    name                     http
    port                     443
    storage.total_limit_size 1G
    tls                      on
    tls.verify               on

`
	logPipeline := &telemetryv1alpha1.LogPipeline{
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Input: telemetryv1alpha1.Input{
				Application: telemetryv1alpha1.ApplicationInput{
					KeepAnnotations: true,
					DropLabels:      false,
				},
			},
			Filters: []telemetryv1alpha1.Filter{
				{
					Custom: `
						name grep
						regex log aa
					`,
				},
			},
			Output: telemetryv1alpha1.Output{
				HTTP: telemetryv1alpha1.HTTPOutput{
					Dedot: true,
					Host: telemetryv1alpha1.ValueType{
						Value: "localhost",
					},
				},
			},
		},
	}
	logPipeline.Name = "foo"
	pipelineConfig := PipelineConfig{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
		FsBufferLimit:     "1G",
	}

	actual, err := MergeSectionsConfig(logPipeline, pipelineConfig)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}
