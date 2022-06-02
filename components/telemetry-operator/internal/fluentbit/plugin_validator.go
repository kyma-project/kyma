package fluentbit

import (
	"fmt"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
)

//go:generate mockery --name PluginValidator --filename plugin_validator.go
type PluginValidator interface {
	Validate(logPipeline *telemetryv1alpha1.LogPipeline) error
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

func (v *pluginValidator) Validate(logPipeline *telemetryv1alpha1.LogPipeline) error {
	err := v.validateFilterPlugins(logPipeline)
	if err != nil {
		return err
	}

	err = v.validateOutputPlugins(logPipeline)
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
		if !isPluginAllowed(name, v.allowedFilterPlugins, logPipeline.Spec.EnableUnsupportedPlugins) {
			return fmt.Errorf("filter plugin %s is not allowed", name)
		}
	}
	return nil
}

func (v *pluginValidator) validateOutputPlugins(logPipeline *telemetryv1alpha1.LogPipeline) error {
	for _, outputPlugin := range logPipeline.Spec.Outputs {
		section, err := parseSection(outputPlugin.Content)
		if err != nil {
			return err
		}
		name, err := getSectionName(section)
		if err != nil {
			return err
		}
		if !isPluginAllowed(name, v.allowedOutputPlugins, logPipeline.Spec.EnableUnsupportedPlugins) {
			return fmt.Errorf("output plugin %s is not allowed", name)
		}
	}
	return nil
}

func isPluginAllowed(pluginName string, allowedPlugins []string, enableAllPlugins bool) bool {
	if len(allowedPlugins) == 0 || enableAllPlugins {
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
