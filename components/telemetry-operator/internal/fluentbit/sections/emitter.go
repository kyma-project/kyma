package sections

import (
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"strings"
)

const emitterTemplate string = `
Name                  rewrite_tag
Match                 %s.*
Emitter_Name          %s
Emitter_Storage.type  %s
Emitter_Mem_Buf_Limit %s`

// TODO delete and keep default case for all empty
var emitterRules = map[string]string{
	"default":          `$log "^.*$" %s.$TAG true`,
	"namespaceInclude": `$kubernetes['namespace_name'] "^%s$" %s.$TAG true`,
	"namespaceExclude": `$kubernetes['namespace_name'] "^(?!.*%s).*" %s.$TAG true`,
}

// CreateEmitter creates the Fluent Bit Rewrite Tag Filter section
func CreateEmitter(config fluentbit.PipelineConfig, logPipeline *telemetryv1alpha1.LogPipeline) string {
	var emitter = fmt.Sprintf(emitterTemplate, config.InputTag, logPipeline.Name, config.StorageType, config.MemoryBufferLimit)
	var sb strings.Builder
	sb.WriteString(fluentbit.BuildConfigSection(fluentbit.FilterConfigHeader, emitter))
	sb.WriteString(appendEmitterRules(logPipeline))
	return sb.String()
}

func appendEmitterRules(logPipeline *telemetryv1alpha1.LogPipeline) string {
	var sb strings.Builder
	if len(logPipeline.Spec.Input.Application.Namespaces) > 0 {
		sb.WriteString("Rule                  $kubernetes['namespace_name']\"^")
		sb.WriteString(strings.Join(logPipeline.Spec.Input.Application.Namespaces, "|"))
		sb.WriteString(fmt.Sprintf("$\" %s.$TAG true", logPipeline.Name))
		sb.WriteByte('\n')
	}

	if len(logPipeline.Spec.Input.Application.ExcludeNamespaces) > 0 {
		sb.WriteString("Rule                  $kubernetes['namespace_name']\"^(?!")
		sb.WriteString(strings.Join(logPipeline.Spec.Input.Application.ExcludeNamespaces, "$|"))
		sb.WriteString(fmt.Sprintf("$).*\" %s.$TAG true", logPipeline.Name))
		sb.WriteByte('\n')
	}

	if len(logPipeline.Spec.Input.Application.Containers) > 0 {
		sb.WriteString("Rule                  $kubernetes['container_name']\"^")
		sb.WriteString(strings.Join(logPipeline.Spec.Input.Application.Containers, "|"))
		sb.WriteString(fmt.Sprintf("$\" %s.$TAG true", logPipeline.Name))
		sb.WriteByte('\n')
	}

	if len(logPipeline.Spec.Input.Application.ExcludeContainers) > 0 {
		sb.WriteString("Rule                  $kubernetes['container_name']\"^(?!")
		sb.WriteString(strings.Join(logPipeline.Spec.Input.Application.ExcludeContainers, "$|"))
		sb.WriteString(fmt.Sprintf("$).*\" %s.$TAG true", logPipeline.Name))
		sb.WriteByte('\n')
	}

	return sb.String()
}
