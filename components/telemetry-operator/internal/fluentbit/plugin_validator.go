package fluentbit

import (
	"fmt"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
)

//go:generate mockery --name PluginValidator --filename plugin_validator.go
type PluginValidator interface {
	Validate(logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error
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

func (v *pluginValidator) Validate(logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error {

	if err := validateFilters(filterPlugin.Custom, logPipeline, v.deniedFilterPlugins, logPipelines); err != nil {
		return err
	}

	if err := validateOutputs(filterPlugin.Custom, logPipeline, v.deniedFilterPlugins, logPipelines); err != nil {
		return err
	}

	return nil

	for _, filterPlugin := range logPipeline.Spec.Filters {
		if err := checkIfPluginIsValid("filter", filterPlugin.Custom, logPipeline, v.deniedFilterPlugins, logPipelines); err != nil {
			return err
		}
	}

	for _, outputPlugin := range logPipeline.Spec.Outputs {
		if err := checkIfPluginIsValid("output", outputPlugin.Custom, logPipeline, v.deniedOutputPlugins, logPipelines); err != nil {
			return err
		}
	}

	return nil
}

func validateFilters(filters []telemetryv1alpha1.Filter, logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error {
	for _, filterPlugin := range filters {
		if err := checkIfPluginIsValid("filter", filterPlugin.Custom, logPipeline, v.deniedFilterPlugins, logPipelines); err != nil {
			return err
		}
	}
}

func checkForDeniedFilters(pluginContent string, deniedPlugins []string) error {
	section, err := parseSection(pluginContent)
	if err != nil {
		return err
	}

	pluginName, err := getSectionName(section)
	if err != nil {
		return err
	}

	for _, deniedPlugin := range deniedPlugins {
		if strings.EqualFold(pluginName, deniedPlugin) {
			return fmt.Errorf("filter plugin '%s' is not supported. ", pluginName)
		}
	}
	return nil
}

func checkForDeniedOutputs(pluginContent string, deniedPlugins []string) error {
	section, err := parseSection(pluginContent)
	if err != nil {
		return err
	}

	pluginName, err := getSectionName(section)
	if err != nil {
		return err
	}

	for _, deniedPlugin := range deniedPlugins {
		if strings.EqualFold(pluginName, deniedPlugin) {
			return fmt.Errorf("output plugin '%s' is not supported. ", pluginName)
		}
	}
	return nil
}

func validateFiltersMatchCondition(pluginContent string, logPipeline *telemetryv1alpha1.LogPipeline, deniedPlugins []string, logPipelines *telemetryv1alpha1.LogPipelineList) error {

}

func validateOutputsMatchCondition(pluginContent string, logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error {
	section, err := parseSection(pluginContent)
	if err != nil {
		return err
	}

	pluginName, err := getSectionName(section)
	if err != nil {
		return err
	}

	matchCond := getMatchCondition(section)
	if matchCond == "" {
		return nil
	}
	if matchCond == "*" {
		return fmt.Errorf("output plugin '%s' with match condition '*' (match all) is not allowed", pluginName)
	}

	if isValid, logPipelinesNames := isMatchCondValid(matchCond, logPipeline.Name, logPipelines); !isValid {
		validLogPipelinesNames := fmt.Sprintf("'%s' (current logpipeline name)", logPipeline.Name)
		if len(logPipelinesNames) > 0 {
			validLogPipelinesNames += fmt.Sprintf(" or '%s' (other existing logpipelines names)", logPipelinesNames)
		}
		return fmt.Errorf("output plugin '%s' with match condition '%s' is not allowed. Valid match conditions are: %s",
			pluginName, matchCond, validLogPipelinesNames)
	}
	return nil
}

func getSectionName(section map[string]string) (string, error) {
	if name, hasKey := section["name"]; hasKey {
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
