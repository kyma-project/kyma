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

package logpipeline

import (
	"context"
	"fmt"
	"net/http"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/webhook/logpipeline/validation"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//go:generate mockery --name DryRunner --filename dryrun.go
type DryRunner interface {
	RunPipeline(ctx context.Context, pipeline *telemetryv1alpha1.LogPipeline) error
}

const (
	StatusReasonConfigurationError = "InvalidConfiguration"
)

// +kubebuilder:webhook:path=/validate-logpipeline,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.kyma-project.io,resources=logpipelines,verbs=create;update,versions=v1alpha1,name=vlogpipeline.kb.io,admissionReviewVersions=v1
type ValidatingWebhookHandler struct {
	client.Client
	inputValidator        validation.InputValidator
	variablesValidator    validation.VariablesValidator
	filterValidator       validation.FilterValidator
	maxPipelinesValidator validation.MaxPipelinesValidator
	outputValidator       validation.OutputValidator
	fileValidator         validation.FilesValidator
	decoder               *admission.Decoder
	dryRunner             DryRunner
}

func NewValidatingWebhookHandler(client client.Client, inputValidator validation.InputValidator, variablesValidator validation.VariablesValidator, filterValidator validation.FilterValidator, maxPipelinesValidator validation.MaxPipelinesValidator, outputValidator validation.OutputValidator, fileValidator validation.FilesValidator, dryRunner DryRunner) *ValidatingWebhookHandler {
	return &ValidatingWebhookHandler{
		Client:                client,
		inputValidator:        inputValidator,
		variablesValidator:    variablesValidator,
		filterValidator:       filterValidator,
		maxPipelinesValidator: maxPipelinesValidator,
		outputValidator:       outputValidator,
		fileValidator:         fileValidator,
		dryRunner:             dryRunner,
	}
}

func (v *ValidatingWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := logf.FromContext(ctx)

	logPipeline := &telemetryv1alpha1.LogPipeline{}
	if err := v.decoder.Decode(req, logPipeline); err != nil {
		log.Error(err, "Failed to decode LogPipeline")
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := v.validateLogPipeline(ctx, logPipeline); err != nil {
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

	if logPipeline.ContainsCustomPlugin() {
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

func (v *ValidatingWebhookHandler) validateLogPipeline(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) error {
	log := logf.FromContext(ctx)

	var logPipelines telemetryv1alpha1.LogPipelineList
	if err := v.List(ctx, &logPipelines); err != nil {
		return err
	}

	if err := v.maxPipelinesValidator.Validate(logPipeline, &logPipelines); err != nil {
		log.Error(err, "Maximum number of log pipelines reached")
		return err
	}

	if err := v.inputValidator.Validate(&logPipeline.Spec.Input); err != nil {
		log.Error(err, "Failed to validate Fluent Bit input")
		return err
	}

	if err := v.variablesValidator.Validate(ctx, logPipeline, &logPipelines); err != nil {
		log.Error(err, "Failed to validate variables")
		return err
	}

	if err := v.filterValidator.Validate(logPipeline); err != nil {
		log.Error(err, "Failed to validate plugins")
		return err
	}

	if err := v.outputValidator.Validate(logPipeline); err != nil {
		log.Error(err, "Failed to validate Fluent Bit output")
		return err
	}

	if err := v.fileValidator.Validate(logPipeline, &logPipelines); err != nil {
		log.Error(err, "Failed to validate Fluent Bit config")
		return err
	}

	if err := v.dryRunner.RunPipeline(ctx, logPipeline); err != nil {
		log.Error(err, "Failed to validate Fluent Bit config")
		return err
	}

	return nil
}

func (v *ValidatingWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
