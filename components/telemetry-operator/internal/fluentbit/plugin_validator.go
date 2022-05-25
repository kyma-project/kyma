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
	allowedFilterPlugins []string
	allowedOutputPlugins []string
}

func NewPluginValidator(allowedFilterPlugins []string, allowedOutputPlugins []string) PluginValidator {
	return &pluginValidator{
		allowedFilterPlugins: allowedFilterPlugins,
		allowedOutputPlugins: allowedOutputPlugins,
	}
}

func (v *pluginValidator) Validate(logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error {
	err := v.validateFilterPlugins(logPipeline)
	if err != nil {
		return err
	}

	err = v.validateOutputPlugins(logPipeline, logPipelines)
	if err != nil {
		return err
	}

	return nil
}

func (v *pluginValidator) validateFilterPlugins(logPipeline *telemetryv1alpha1.LogPipeline) error {
	for _, filterPlugin := range logPipeline.Spec.Filters {
		section, err := parseSection(filterPlugin.Content)
		if err != nil {
			return err
		}
		name, err := getSectionName(section)
		if err != nil {
			return err
		}
		if !isPluginAllowed(name, v.allowedFilterPlugins) {
			return fmt.Errorf("filter plugin '%s' is not allowed", name)
		}

		matchCond, err := getMatchCondition(section)
		if matchCond == "*" {
			return fmt.Errorf("filter plugin '%s' with match condition '*' (match all) is not allowed", name)
		}
	}
	return nil
}

func (v *pluginValidator) validateOutputPlugins(logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error {
	for _, outputPlugin := range logPipeline.Spec.Outputs {
		section, err := parseSection(outputPlugin.Content)
		if err != nil {
			return err
		}
		name, err := getSectionName(section)
		if err != nil {
			return err
		}
		if !isPluginAllowed(name, v.allowedOutputPlugins) {
			return fmt.Errorf("output plugin '%s' is not allowed", name)
		}

		matchCond, err := getMatchCondition(section)
		if err != nil {
			return err
		}
		if matchCond == "*" {
			return fmt.Errorf("output plugin '%s' with match condition '*' (match all) is not allowed", name)
		}

		if err = isMatchCondAllowed(name, matchCond, logPipeline.Name, logPipelines); err != nil {
			return err
		}

	}
	return nil
}

func isPluginAllowed(pluginName string, allowedPlugins []string) bool {
	if len(allowedPlugins) == 0 {
		return true
	}
	for _, allowedPlugin := range allowedPlugins {
		if strings.EqualFold(pluginName, allowedPlugin) {
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

func getMatchCondition(section map[string]string) (string, error) {
	if matchCond, hasKey := section["match"]; hasKey {
		return matchCond, nil
	}
	return "", fmt.Errorf("configuration section does not have match attribute")
}

func isMatchCondAllowed(pluginName, matchCond, logPipelineName string, logPipelines *telemetryv1alpha1.LogPipelineList) error {
	matchCondValid := false
	var pipelineNames []string

	if pluginName == logPipelineName {
		return nil
	}

	for _, l := range logPipelines.Items {
		if matchCond == fmt.Sprintf("%s.*", l.Name) {
			matchCondValid = true
		}
		pipelineNames = append(pipelineNames, l.Name)
	}
	if !matchCondValid {
		return fmt.Errorf("output plugin '%s' should have match condition matching any of the existing logpipeline names '%s'", pluginName, pipelineNames)
	}
	return nil
}
