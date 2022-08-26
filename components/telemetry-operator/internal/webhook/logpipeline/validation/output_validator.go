package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

//go:generate mockery --name OutputValidator --filename output_validator.go
type OutputValidator interface {
	Validate(logPipeline *telemetryv1alpha1.LogPipeline) error
}

type outputValidator struct {
	deniedOutputPlugins []string
}

func NewOutputValidator(deniedOutputPlugins ...string) OutputValidator {
	return &outputValidator{
		deniedOutputPlugins: deniedOutputPlugins,
	}
}

func (v *outputValidator) Validate(pipeline *telemetryv1alpha1.LogPipeline) error {
	output := pipeline.Spec.Output

	if err := checkSingleOutputPlugin(output); err != nil {
		return err
	}

	if output.IsHTTPDefined() {
		if err := validateHTTPOutput(output.HTTP); err != nil {
			return err
		}
	}

	if output.IsLokiDefined() {
		if err := validateLokiOutput(output.Loki); err != nil {
			return err
		}
	}

	if output.IsCustomDefined() {
		if err := v.validateCustomOutput(pipeline.Spec.Output.Custom); err != nil {
			return err
		}
	}

	return nil
}

func checkSingleOutputPlugin(output telemetryv1alpha1.Output) error {
	if !output.IsAnyDefined() {
		return fmt.Errorf("no output is defined, you must define one output")
	}
	if !output.IsSingleDefined() {
		return fmt.Errorf("multiple output plugins are defined, you must define only one output")
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

func (v *outputValidator) validateCustomOutput(content string) error {
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

	for _, deniedPlugin := range v.deniedOutputPlugins {
		if strings.EqualFold(pluginName, deniedPlugin) {
			return fmt.Errorf("plugin '%s' is forbidden. ", pluginName)
		}
	}

	if section.ContainsKey("match") {
		return fmt.Errorf("plugin '%s' contains match condition. Match conditions are forbidden", pluginName)
	}

	if section.ContainsKey("storage.total_limit_size") {
		return fmt.Errorf("plugin '%s' contains forbidden configuration key 'storage.total_limit_size'", pluginName)
	}

	return nil
}
