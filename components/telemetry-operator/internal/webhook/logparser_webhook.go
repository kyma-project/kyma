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
	"github.com/google/uuid"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fs"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/validation"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//+kubebuilder:webhook:path=/validate-logparser,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.kyma-project.io,resources=logparsers,verbs=create;update,versions=v1alpha1,name=vlogparser.kb.io,admissionReviewVersions=v1
type LogparserValidator struct {
	client.Client

	fluentBitConfigMap types.NamespacedName
	parserValidator    validation.ParserValidator
	pipelineConfig     fluentbit.PipelineConfig
	configValidator    validation.ConfigValidator
	fsWrapper          fs.Wrapper

	decoder *admission.Decoder
}

func NewLogParserValidator(
	client client.Client,
	fluentBitConfigMap string,
	namespace string,
	parserValidator validation.ParserValidator,
	pipelineConfig fluentbit.PipelineConfig,
	configValidator validation.ConfigValidator,
	fsWrapper fs.Wrapper,

) *LogparserValidator {
	return &LogparserValidator{
		Client: client,
		fluentBitConfigMap: types.NamespacedName{
			Name:      fluentBitConfigMap,
			Namespace: namespace,
		},
		parserValidator: parserValidator,
		pipelineConfig:  pipelineConfig,
		configValidator: configValidator,
		fsWrapper:       fsWrapper,
	}
}

func (v *LogparserValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := logf.FromContext(ctx)

	log.Info("I am invoked !!!!!")
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
					Reason:  StatusReasonConfigurationError,
					Message: err.Error(),
				},
			},
		}
	}
	return admission.Allowed("LogParser validation successful")
}

func (v *LogparserValidator) validateLogParser(ctx context.Context, logParser *telemetryv1alpha1.LogParser) error {
	log := logf.FromContext(ctx)
	err := v.parserValidator.Validate(logParser)
	if err != nil {
		return err
	}

	utils := kubernetes.NewUtils(v.Client)
	var pipeLine *telemetryv1alpha1.LogPipeline
	currentBaseDirectory := fluentBitConfigDirectory + uuid.New().String()

	configFiles, err := utils.GetFluentBitConfig(ctx, currentBaseDirectory, fluentBitParsersConfigMapKey, v.fluentBitConfigMap, v.pipelineConfig, pipeLine, logParser)
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
	if err = v.configValidator.Validate(ctx, fmt.Sprintf("%s/fluent-bit.conf", currentBaseDirectory)); err != nil {
		log.Error(err, "Failed to validate Fluent Bit config")
		return err
	}

	return nil
}

func (v *LogparserValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
