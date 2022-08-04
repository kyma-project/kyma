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

package logparser

import (
	"context"
	"net/http"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/webhook/logpipeline"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/google/uuid"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fs"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/validation"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//go:generate mockery --name DryRunner --filename mocks.go
type DryRunner interface {
	RunParser(ctx context.Context, configFilePath string) error
}

const (
	fluentBitConfigDirectory     = "/tmp/dry-run"
	fluentBitParsersConfigMapKey = "parsers.conf"
)

//+kubebuilder:webhook:path=/validate-logparser,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.kyma-project.io,resources=logparsers,verbs=create;update,versions=v1alpha1,name=vlogparser.kb.io,admissionReviewVersions=v1
type ValidatingWebhookHandler struct {
	client.Client
	fluentBitConfigMap types.NamespacedName
	parserValidator    validation.ParserValidator
	pipelineConfig     fluentbit.PipelineConfig
	dryRunner          DryRunner
	fsWrapper          fs.Wrapper
	decoder            *admission.Decoder
	daemonSetUtils     *fluentbit.DaemonSetUtils
}

func NewValidatingWebhookHandler(
	client client.Client,
	fluentBitConfigMap string,
	namespace string,
	parserValidator validation.ParserValidator,
	pipelineConfig fluentbit.PipelineConfig,
	dryRunner DryRunner,
	fsWrapper fs.Wrapper,
	restartsTotal prometheus.Counter,
) *ValidatingWebhookHandler {
	fluentBitConfigMapNamespacedName := types.NamespacedName{
		Name:      fluentBitConfigMap,
		Namespace: namespace,
	}
	daemonSetUtils := fluentbit.NewDaemonSetUtils(client, fluentBitConfigMapNamespacedName, restartsTotal)
	return &ValidatingWebhookHandler{
		Client:             client,
		fluentBitConfigMap: fluentBitConfigMapNamespacedName,
		parserValidator:    parserValidator,
		pipelineConfig:     pipelineConfig,
		dryRunner:          dryRunner,
		fsWrapper:          fsWrapper,
		daemonSetUtils:     daemonSetUtils,
	}
}

func (v *ValidatingWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := logf.FromContext(ctx)

	logParser := &telemetryv1alpha1.LogParser{}
	if err := v.decoder.Decode(req, logParser); err != nil {
		log.Error(err, "Failed to decode LogParser")
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := v.validateLogParser(ctx, logParser); err != nil {
		log.Error(err, "LogParser rejected")
		return admission.Response{
			AdmissionResponse: admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Code:    int32(http.StatusForbidden),
					Reason:  logpipeline.StatusReasonConfigurationError,
					Message: err.Error(),
				},
			},
		}
	}
	return admission.Allowed("LogParser validation successful")
}

func (v *ValidatingWebhookHandler) validateLogParser(ctx context.Context, logParser *telemetryv1alpha1.LogParser) error {
	log := logf.FromContext(ctx)
	err := v.parserValidator.Validate(logParser)
	if err != nil {
		return err
	}

	var pipeLine *telemetryv1alpha1.LogPipeline
	currentBaseDirectory := fluentBitConfigDirectory + uuid.New().String()

	configFiles, err := v.daemonSetUtils.GetFluentBitConfig(ctx, currentBaseDirectory, fluentBitParsersConfigMapKey, v.fluentBitConfigMap, v.pipelineConfig, pipeLine, logParser)
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

	if err = v.dryRunner.RunParser(ctx, currentBaseDirectory); err != nil {
		log.Error(err, "Failed to validate Fluent Bit config")
		return err
	}

	return nil
}

func (v *ValidatingWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
