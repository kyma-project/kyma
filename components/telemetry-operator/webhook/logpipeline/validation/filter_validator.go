package validation

import (
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config"

	"github.com/pkg/errors"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

//go:generate mockery --name PluginValidator --filename plugin_validator.go
type FilterValidator interface {
	Validate(logPipeline *telemetryv1alpha1.LogPipeline) error
}

type filterValidator struct {
	deniedFilterPlugins []string
}

func NewFilterValidator(deniedFilterPlugins ...string) FilterValidator {
	return &filterValidator{
		deniedFilterPlugins: deniedFilterPlugins,
	}
}

func (v *filterValidator) Validate(logPipeline *telemetryv1alpha1.LogPipeline) error {
	err := v.validateFilters(logPipeline)
	if err != nil {
		return errors.Wrap(err, "error validating filter plugins")
	}
	return nil
}

func (v *filterValidator) validateFilters(pipeline *telemetryv1alpha1.LogPipeline) error {
	for _, filterPlugin := range pipeline.Spec.Filters {
		if err := v.validateCustomFilter(filterPlugin.Custom); err != nil {
			return err
		}
	}
	return nil
}

func (v *filterValidator) validateCustomFilter(content string) error {
	if content == "" {
		return nil
	}

	section, err := config.ParseCustomSection(content)
	if err != nil {
		return err
	}

	if !section.ContainsKey("name") {
		return fmt.Errorf("configuration section does not have name attribute")
	}

	pluginName := section.GetByKey("name").Value

	for _, deniedPlugin := range v.deniedFilterPlugins {
		if strings.EqualFold(pluginName, deniedPlugin) {
			return fmt.Errorf("plugin '%s' is forbidden. ", pluginName)
		}
	}

	if section.ContainsKey("match") {
		return fmt.Errorf("plugin '%s' contains match condition. Match conditions are forbidden", pluginName)
	}

	return nil
}
