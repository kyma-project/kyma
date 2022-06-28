package validation

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
)

//go:generate mockery --name PluginValidator --filename plugin_validator.go
type PluginValidator interface {
	Validate(logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error
	ContainsCustomPlugin(logPipeline *telemetryv1alpha1.LogPipeline) bool
}

type pluginValidator struct {
	deniedFilterPlugins []string
	deniedOutputPlugins []string
}

func NewPluginValidator(deniedFilterPlugins, deniedOutputPlugins []string) PluginValidator {
	return &pluginValidator{
		deniedFilterPlugins: deniedFilterPlugins,
		deniedOutputPlugins: deniedOutputPlugins,
	}
}

// ContainsCustomPlugin returns true if the pipeline
// contains any custom filters or outputs
func (pv *pluginValidator) ContainsCustomPlugin(logPipeline *telemetryv1alpha1.LogPipeline) bool {
	for _, filterPlugin := range logPipeline.Spec.Filters {
		if filterPlugin.Custom != "" {
			return true
		}
	}
	if logPipeline.Spec.Output.Custom != "" {
		return true
	}
	for _, f := range logPipeline.Spec.Filters {
		if f.Custom != "" {
			return true
		}
	}
	return false
}

// Validate returns an error if validation fails
// because of using denied plugins or errors in match conditions
// for filters or outputs.
func (pv *pluginValidator) Validate(logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error {
	err := pv.validateFilters(logPipeline, logPipelines)
	if err != nil {
		return errors.Wrap(err, " error validating Filter plugins")
	}
	err = pv.validateOutputs(logPipeline, logPipelines)
	if err != nil {
		return errors.Wrap(err, " error validating Output plugins")
	}
	return nil
}

func (pv *pluginValidator) validateFilters(pipeline *telemetryv1alpha1.LogPipeline, pipelines *telemetryv1alpha1.LogPipelineList) error {
	for _, filterPlugin := range pipeline.Spec.Filters {
		if err := checkIfPluginIsValid(filterPlugin.Custom, pipeline, pv.deniedFilterPlugins, pipelines); err != nil {
			return err
		}
	}
	return nil
}

func (pv *pluginValidator) validateOutputs(pipeline *telemetryv1alpha1.LogPipeline, pipelines *telemetryv1alpha1.LogPipelineList) error {

	if err := checkIfPluginIsValid(pipeline.Spec.Output.Custom, pipeline, pv.deniedOutputPlugins, pipelines); err != nil {
		return err
	}

	return nil
}

func checkIfPluginIsValid(content string, pipeline *telemetryv1alpha1.LogPipeline, denied []string, pipelines *telemetryv1alpha1.LogPipelineList) error {
	customSection, err := parseSection(content)
	if err != nil {
		return err
	}
	name, err := getCustomName(customSection)
	if err != nil {
		return err
	}

	for _, deniedPlugin := range denied {
		if strings.EqualFold(name, deniedPlugin) {
			return fmt.Errorf("plugin '%s' is not supported. ", name)
		}
	}

	matchCondition := getMatchCondition(customSection)
	if matchCondition == "" {
		return nil
	}
	if matchCondition == "*" {
		return fmt.Errorf("plugin '%s' with match condition '*' (match all) is not allowed", name)
	}

	if isValid, logPipelinesNames := isMatchCondValid(matchCondition, pipeline.Name, pipelines); !isValid {
		validLogPipelinesNames := fmt.Sprintf("'%s' (current logpipeline name)", pipeline.Name)
		if len(logPipelinesNames) > 0 {
			validLogPipelinesNames += fmt.Sprintf(" or '%s' (other existing logpipelines names)", logPipelinesNames)
		}
		return fmt.Errorf("plugin '%s' with match condition '%s' is not allowed. Valid match conditions are: %s",
			name, matchCondition, validLogPipelinesNames)
	}
	return nil
}

func getCustomName(custom map[string]string) (string, error) {
	fmt.Printf("custom: %v", custom)
	if name, hasKey := custom["name"]; hasKey {
		return name, nil
	}
	return "", fmt.Errorf("configuration section does not have name attribute")
}

func getMatchCondition(section map[string]string) string {
	if matchCond, hasKey := section["match"]; hasKey {
		return matchCond
	}
	return ""
}

func isMatchCondValid(matchCond, logPipelineName string, logPipelines *telemetryv1alpha1.LogPipelineList) (bool, []string) {
	if strings.HasPrefix(matchCond, fmt.Sprintf("%s.", logPipelineName)) {
		return true, nil
	}

	var pipelineNames []string
	for _, l := range logPipelines.Items {
		if strings.HasPrefix(matchCond, fmt.Sprintf("%s.", l.Name)) {
			return true, nil
		}
		pipelineNames = append(pipelineNames, l.Name)
	}

	return false, pipelineNames
}

func parseSection(section string) (map[string]string, error) {
	result := make(map[string]string)

	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, " ")
		if !found {
			return nil, fmt.Errorf("invalid line: %s", line)
		}
		result[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}
	return result, nil
}
