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
	err := v.validateFilterPlugins(logPipeline, logPipelines)
	if err != nil {
		return err
	}

	err = v.validateOutputPlugins(logPipeline, logPipelines)
	if err != nil {
		return err
	}

	return nil
}

func (v *pluginValidator) validateFilterPlugins(logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error {
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

		matchCond := getMatchCondition(section)

		if matchCond == "" {
			return nil
		}

		if matchCond == "*" {
			return fmt.Errorf("filter plugin '%s' with match condition '*' (match all) is not allowed", name)
		}
		if err = isMatchCondAllowed(name, matchCond, logPipeline.Name, logPipelines); err != nil {
			return err
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

		matchCond := getMatchCondition(section)

		if matchCond == "" {
			return nil
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

func getMatchCondition(section map[string]string) string {
	if matchCond, hasKey := section["match"]; hasKey {
		return matchCond
	}
	return ""
}

func isMatchCondAllowed(pluginName, matchCond, logPipelineName string, logPipelines *telemetryv1alpha1.LogPipelineList) error {
	matchCondValid := false
	var pipelineNames []string

	if matchCond == fmt.Sprintf("%s.", logPipelineName) {
		return nil
	}

	for _, l := range logPipelines.Items {
		if matchCond == fmt.Sprintf("%s.*", l.Name) {
			matchCondValid = true
		}
		pipelineNames = append(pipelineNames, l.Name)
	}

	if !matchCondValid {
		if len(logPipelines.Items) == 0 {
			return fmt.Errorf("output plugin '%s' should have match condition (pipelineName.*) matching the logpipeline name '%s'", pluginName, logPipelineName)
		}
		return fmt.Errorf("output plugin '%s' should have match condition (pipelineName.*) matching any of the current '%s' or existing logpipeline names '%s'", pluginName, logPipelineName, pipelineNames)
	}
	return nil
}
