package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config"

	"github.com/pkg/errors"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

//go:generate mockery --name PluginValidator --filename plugin_validator.go
type PluginValidator interface {
	Validate(logPipeline *telemetryv1alpha1.LogPipeline) error
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
	return logPipeline.Spec.Output.Custom != ""
}

// Validate returns an error if validation fails
// because of using denied plugins or errors in match conditions
// for filters or outputs.
func (pv *pluginValidator) Validate(logPipeline *telemetryv1alpha1.LogPipeline) error {
	err := pv.validateFilters(logPipeline)
	if err != nil {
		return errors.Wrap(err, "error validating filter plugins")
	}
	err = pv.validateOutput(logPipeline)
	if err != nil {
		return errors.Wrap(err, "error validating output plugin")
	}
	return nil
}

func (pv *pluginValidator) validateFilters(pipeline *telemetryv1alpha1.LogPipeline) error {
	for _, filterPlugin := range pipeline.Spec.Filters {
		if err := validateCustom(filterPlugin.Custom, pv.deniedFilterPlugins); err != nil {
			return err
		}
	}
	return nil
}

func (pv *pluginValidator) validateOutput(pipeline *telemetryv1alpha1.LogPipeline) error {
	if err := checkSingleOutputPlugin(pipeline.Spec.Output); err != nil {
		return err
	}
	if err := validateCustom(pipeline.Spec.Output.Custom, pv.deniedOutputPlugins); err != nil {
		return err
	}
	if err := validateHTTPOutput(pipeline.Spec.Output.HTTP); err != nil {
		return err
	}
	if err := validateLokiOutput(pipeline.Spec.Output.Loki); err != nil {
		return err
	}
	return nil
}

func checkSingleOutputPlugin(output telemetryv1alpha1.Output) error {
	outputPluginCount := 0
	if len(output.Custom) != 0 {
		outputPluginCount++
	}
	if output.HTTP.Host.IsDefined() {
		outputPluginCount++
	}
	if output.Loki.URL.IsDefined() {
		outputPluginCount++
	}

	if outputPluginCount == 0 {
		return fmt.Errorf("no output is defined, you must define one output")
	}
	if outputPluginCount > 1 {
		return fmt.Errorf("multiple output plugins are defined, you must define only one output")
	}
	return nil
}

func validateCustom(content string, denied []string) error {
	if content == "" {
		return nil
	}

	customSection, err := config.ParseCustomSection(content)
	if err != nil {
		return err
	}

	if !customSection.ContainsKey("name") {
		return fmt.Errorf("configuration section does not have name attribute")
	}

	pluginName := customSection.GetByKey("name").Value

	for _, deniedPlugin := range denied {
		if strings.EqualFold(pluginName, deniedPlugin) {
			return fmt.Errorf("plugin '%s' is not supported. ", pluginName)
		}
	}

	if customSection.ContainsKey("match") {
		return fmt.Errorf("plugin '%s' contains match condition. Match conditions are forbidden", pluginName)
	}

	return nil
}

func validateLokiOutput(lokiOutput *telemetryv1alpha1.LokiOutput) error {
	if lokiOutput.URL.Value != "" && !validURL(lokiOutput.URL.Value) {
		return fmt.Errorf("invalid hostname '%s'", lokiOutput.URL.Value)
	}
	if !lokiOutput.URL.IsDefined() && (len(lokiOutput.Labels) != 0 || len(lokiOutput.RemoveKeys) != 0) {
		return fmt.Errorf("loki output needs to have a URL configured")
	}
	return nil

}

func validateHTTPOutput(httpOutput *telemetryv1alpha1.HTTPOutput) error {
	if httpOutput.Host.Value != "" && !validHostname(httpOutput.Host.Value) {
		return fmt.Errorf("invalid hostname '%s'", httpOutput.Host.Value)
	}
	if httpOutput.URI != "" && !strings.HasPrefix(httpOutput.URI, "/") {
		return fmt.Errorf("uri has to start with /")
	}
	if !httpOutput.Host.IsDefined() && (httpOutput.User.IsDefined() || httpOutput.Password.IsDefined() || httpOutput.URI != "" || httpOutput.Port != "" || httpOutput.Compress != "" || httpOutput.TLSConfig.Disabled || httpOutput.TLSConfig.SkipCertificateValidation) {
		return fmt.Errorf("http output needs to have a host configured")
	}
	return nil
}

func validURL(host string) bool {
	host = strings.Trim(host, " ")

	_, err := url.ParseRequestURI(host)
	if err != nil {
		return false
	}

	u, err := url.Parse(host)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return true
}

func validHostname(host string) bool {
	host = strings.Trim(host, " ")
	re, _ := regexp.Compile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)
	return re.MatchString(host)
}
