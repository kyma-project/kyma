/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhook

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/validation"

	"github.com/google/uuid"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fs"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	fluentBitConfigDirectory       = "/tmp/dry-run"
	StatusReasonConfigurationError = "InvalidConfiguration"
	fluentBitParsersConfigMapKey   = "parsers.conf"
)

//+kubebuilder:webhook:path=/validate-logpipeline,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.kyma-project.io,resources=logpipelines,verbs=create;update,versions=v1alpha1,name=vlogpipeline.kb.io,admissionReviewVersions=v1
type LogPipelineValidator struct {
	client.Client
	fluentBitConfigMap    types.NamespacedName
	inputValidator        validation.InputValidator
	variablesValidator    validation.VariablesValidator
	configValidator       validation.ConfigValidator
	pluginValidator       validation.PluginValidator
	maxPipelinesValidator validation.MaxPipelinesValidator
	outputValidator       validation.OutputValidator
	fileValidator         validation.FilesValidator
	pipelineConfig        fluentbit.PipelineConfig
	fsWrapper             fs.Wrapper
	decoder               *admission.Decoder
	daemonSetUtils        *fluentbit.DaemonSetUtils
}

func NewLogPipeLineValidator(
	client client.Client,
	fluentBitConfigMap string,
	namespace string,
	inputValidator validation.InputValidator,
	variablesValidator validation.VariablesValidator,
	configValidator validation.ConfigValidator,
	pluginValidator validation.PluginValidator,
	maxPipelinesValidator validation.MaxPipelinesValidator,
	outputValidator validation.OutputValidator,
	fileValidator validation.FilesValidator,
	pipelineConfig fluentbit.PipelineConfig,
	fsWrapper fs.Wrapper,
	restartsTotal prometheus.Counter) *LogPipelineValidator {
	fluentBitConfigMapNamespacedName := types.NamespacedName{
		Name:      fluentBitConfigMap,
		Namespace: namespace,
	}
	daemonSetUtils := fluentbit.NewDaemonSetUtils(client, fluentBitConfigMapNamespacedName, restartsTotal)

	return &LogPipelineValidator{
		Client:                client,
		fluentBitConfigMap:    fluentBitConfigMapNamespacedName,
		inputValidator:        inputValidator,
		variablesValidator:    variablesValidator,
		configValidator:       configValidator,
		pluginValidator:       pluginValidator,
		maxPipelinesValidator: maxPipelinesValidator,
		outputValidator:       outputValidator,
		fileValidator:         fileValidator,
		fsWrapper:             fsWrapper,
		pipelineConfig:        pipelineConfig,
		daemonSetUtils:        daemonSetUtils,
	}
}

func (v *LogPipelineValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := logf.FromContext(ctx)

	logPipeline := &telemetryv1alpha1.LogPipeline{}
	if err := v.decoder.Decode(req, logPipeline); err != nil {
		log.Error(err, "Failed to decode LogPipeline")
		return admission.Errored(http.StatusBadRequest, err)
	}

	currentBaseDirectory := fluentBitConfigDirectory + uuid.New().String()

	if err := v.validateLogPipeline(ctx, currentBaseDirectory, logPipeline); err != nil {
		log.Error(err, "LogPipeline rejected")
		return admission.Response{
			AdmissionResponse: admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Code:    int32(http.StatusForbidden),
					Reason:  StatusReasonConfigurationError,
					Message: err.Error(),
				},
			},
		}
	}
	var warnMsg []string

	if v.pluginValidator.ContainsCustomPlugin(logPipeline) {
		helpText := "https://kyma-project.io/docs/kyma/latest/01-overview/main-areas/observability/obsv-04-telemetry-in-kyma/"
		msg := fmt.Sprintf("Logpipeline '%s' uses unsupported custom filters or outputs. We recommend changing the pipeline to use supported filters or output. See the documentation: %s", logPipeline.Name, helpText)
		warnMsg = append(warnMsg, msg)
	}

	if len(warnMsg) != 0 {
		return admission.Response{
			AdmissionResponse: admissionv1.AdmissionResponse{
				Allowed:  true,
				Warnings: warnMsg,
			},
		}
	}

	return admission.Allowed("LogPipeline validation successful")
}

func (v *LogPipelineValidator) validateLogPipeline(ctx context.Context, currentBaseDirectory string, logPipeline *telemetryv1alpha1.LogPipeline) error {
	log := logf.FromContext(ctx)

	defer func() {
		if err := v.fsWrapper.RemoveDirectory(currentBaseDirectory); err != nil {
			log.Error(err, "Failed to remove Fluent Bit config directory")
		}
	}()

	var parser *telemetryv1alpha1.LogParser

	configFiles, err := v.daemonSetUtils.GetFluentBitConfig(ctx, currentBaseDirectory, fluentBitParsersConfigMapKey, v.fluentBitConfigMap, v.pipelineConfig, logPipeline, parser)
	if err != nil {
		return err
	}

	log.Info("Fluent Bit config files count", "count", len(configFiles))
	for _, configFile := range configFiles {
		if err = v.fsWrapper.CreateAndWrite(configFile); err != nil {
			log.Error(err, "Failed to write Fluent Bit config file", "filename", configFile.Name, "path", configFile.Path)
			return err
		}
	}

	var logPipelines telemetryv1alpha1.LogPipelineList
	if err := v.List(ctx, &logPipelines); err != nil {
		return err
	}

	if err = v.maxPipelinesValidator.Validate(logPipeline, &logPipelines); err != nil {
		log.Error(err, "Maximum number of log pipelines reached")
		return err
	}

	if err = v.inputValidator.Validate(&logPipeline.Spec.Input); err != nil {
		log.Error(err, "Failed to validate Fluent Bit input")
		return err
	}

	if err = v.variablesValidator.Validate(ctx, logPipeline, &logPipelines); err != nil {
		log.Error(err, "Failed to validate variables")
		return err
	}

	if err = v.pluginValidator.Validate(logPipeline, &logPipelines); err != nil {
		log.Error(err, "Failed to validate plugins")
		return err
	}

	if err = v.outputValidator.Validate(logPipeline); err != nil {
		log.Error(err, "Failed to validate Fluent Bit output")
		return err
	}

	if err = v.configValidator.Validate(ctx, fmt.Sprintf("%s/fluent-bit.conf", currentBaseDirectory)); err != nil {
		log.Error(err, "Failed to validate Fluent Bit config")
		return err
	}

	if err = v.fileValidator.Validate(logPipeline, &logPipelines); err != nil {
		log.Error(err, "Failed to validate Fluent Bit config")
		return err
	}

	return nil
}

func (v *LogPipelineValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
