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
	supportedFilterPlugins []string
	supportedOutputPlugins []string
	deniedFilterPlugins    []string
	deniedOutputPlugins    []string
}

func NewPluginValidator(supportedFilerPlugins, supportedOutputPlugins, deniedFilterPlugins, deniedOutputPlugins []string) PluginValidator {
	return &pluginValidator{
		supportedFilterPlugins: supportedFilerPlugins,
		supportedOutputPlugins: supportedOutputPlugins,
		deniedFilterPlugins:    deniedFilterPlugins,
		deniedOutputPlugins:    deniedOutputPlugins,
	}
}

func (v *pluginValidator) Validate(logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error {

	for _, filterPlugin := range logPipeline.Spec.Filters {
		if err := checkIfPluginIsValid("filter", filterPlugin.Content, logPipeline, v.supportedFilterPlugins, v.deniedFilterPlugins, logPipelines); err != nil {
			return err
		}
	}

	for _, outputPlugin := range logPipeline.Spec.Outputs {
		if err := checkIfPluginIsValid("output", outputPlugin.Content, logPipeline, v.supportedOutputPlugins, v.deniedOutputPlugins, logPipelines); err != nil {
			return err
		}
	}

	return nil
}

func checkIfPluginIsValid(pluginType, pluginContent string, logPipeline *telemetryv1alpha1.LogPipeline, supportedPlugins, deniedPlugins []string, logPipelines *telemetryv1alpha1.LogPipelineList) error {
	var err error
	var section map[string]string
	var pluginName string

	if section, err = parseSection(pluginContent); err != nil {
		return err
	}
	if pluginName, err = getSectionName(section); err != nil {
		return err
	}

	if !isPluginSupported(pluginName, supportedPlugins, deniedPlugins, logPipeline.Spec.EnableUnsupportedPlugins) {
		return fmt.Errorf("%s plugin '%s' is not supported. Following plugins are supported '%s' ", pluginType, pluginName, supportedPlugins)
	}

	matchCond := getMatchCondition(section)
	if matchCond == "" {
		return nil
	}
	if matchCond == "*" {
		return fmt.Errorf("%s plugin '%s' with match condition '*' (match all) is not allowed", pluginType, pluginName)
	}

	if isValid, logPipelinesNames := isMatchCondValid(matchCond, logPipeline.Name, logPipelines); !isValid {
		validLogPipelinesNames := fmt.Sprintf("'%s' (current logpipeline name)", logPipeline.Name)
		if len(logPipelinesNames) > 0 {
			validLogPipelinesNames += fmt.Sprintf(" or '%s' (other existing logpipelines names)", logPipelinesNames)
		}
		return fmt.Errorf("%s plugin '%s' with match condition '%s' is not allowed. Valid match conditions are: %s",
			pluginType, pluginName, matchCond, validLogPipelinesNames)
	}
	return nil
}

func isPluginSupported(pluginName string, supportedPlugins, deniedPlugins []string, enableAllPlugins bool) bool {
	for _, deniedPlugin := range deniedPlugins {
		if strings.EqualFold(pluginName, deniedPlugin) {
			return false
		}
	}

	if len(supportedPlugins) == 0 || enableAllPlugins {
		return true
	}
	for _, supportedPlugin := range supportedPlugins {
		if strings.EqualFold(pluginName, supportedPlugin) {
			return true
		}
	}
	return false
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
