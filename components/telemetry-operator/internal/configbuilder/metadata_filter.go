package configbuilder

import (
	"fmt"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

func createKubernetesMetadataFilter(pipeline *telemetryv1alpha1.LogPipeline) string {
	input := pipeline.Spec.Input.Application
	if input.KeepAnnotations && !input.DropLabels {
		return ""
	}

	liftFilter := NewFilterSectionBuilder().
		AddConfigParam("name", "nest").
		AddConfigParam("match", fmt.Sprintf("%s.*", pipeline.Name)).
		AddConfigParam("operation", "lift").
		AddConfigParam("nested_under", "kubernetes").
		AddConfigParam("add_prefix", "__k8s__").
		Build()

	var dropAnnotationsFilter string
	if !input.KeepAnnotations {
		dropAnnotationsFilter = createDropAnnotations(pipeline.Name)
	}

	var dropLabelsFilter string
	if input.DropLabels {
		dropLabelsFilter = createDropLabels(pipeline.Name)
	}

	nestFilter := NewFilterSectionBuilder().
		AddConfigParam("name", "nest").
		AddConfigParam("match", fmt.Sprintf("%s.*", pipeline.Name)).
		AddConfigParam("operation", "nest").
		AddConfigParam("wildcard", "__k8s__*").
		AddConfigParam("nest_under", "kubernetes").
		AddConfigParam("remove_prefix", "__k8s__").
		Build()

	return fmt.Sprintf("%s%s%s%s", liftFilter, dropAnnotationsFilter, dropLabelsFilter, nestFilter)
}

func createDropAnnotations(pipelineName string) string {
	return NewFilterSectionBuilder().
		AddConfigParam("name", "record_modifier").
		AddConfigParam("match", fmt.Sprintf("%s.*", pipelineName)).
		AddConfigParam("remove_key", "__k8s__annotations").
		Build()
}

func createDropLabels(pipelineName string) string {
	return NewFilterSectionBuilder().
		AddConfigParam("name", "record_modifier").
		AddConfigParam("match", fmt.Sprintf("%s.*", pipelineName)).
		AddConfigParam("remove_key", "__k8s__labels").
		Build()
}
