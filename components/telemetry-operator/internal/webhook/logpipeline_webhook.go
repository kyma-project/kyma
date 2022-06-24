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

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/validation"

	"github.com/google/uuid"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fs"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/secret"
)

const (
	fluentBitConfigDirectory       = "/tmp/dry-run"
	StatusReasonConfigurationError = "InvalidConfiguration"
	fluentBitParsersConfigMapKey   = "parsers.conf"
)

//+kubebuilder:webhook:path=/validate-logpipeline,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.kyma-project.io,resources=logpipelines,verbs=create;update,versions=v1alpha1,name=vlogpipeline.kb.io,admissionReviewVersions=v1
type LogPipelineValidator struct {
	client.Client

	fluentBitConfigMap types.NamespacedName
	variablesValidator validation.VariablesValidator
	configValidator    validation.ConfigValidator
	pluginValidator    validation.PluginValidator
	emitterConfig      fluentbit.EmitterConfig
	fsWrapper          fs.Wrapper

	decoder *admission.Decoder
}

func NewLogPipeLineValidator(
	client client.Client,
	fluentBitConfigMap string,
	namespace string,
	variablesValidator validation.VariablesValidator,
	configValidator validation.ConfigValidator,
	pluginValidator validation.PluginValidator,
	emitterConfig fluentbit.EmitterConfig,
	fsWrapper fs.Wrapper) *LogPipelineValidator {

	return &LogPipelineValidator{
		Client: client,
		fluentBitConfigMap: types.NamespacedName{
			Name:      fluentBitConfigMap,
			Namespace: namespace,
		},
		variablesValidator: variablesValidator,
		configValidator:    configValidator,
		pluginValidator:    pluginValidator,
		fsWrapper:          fsWrapper,
		emitterConfig:      emitterConfig,
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

	secretsFound := v.validateSecrets(ctx, logPipeline)
	if !secretsFound {
		warnMsg := "One or more secrets do not exist, the log pipeline will stay in pending state until the secrets are available"
		return admission.Response{
			AdmissionResponse: admissionv1.AdmissionResponse{
				Allowed:  true,
				Warnings: []string{warnMsg},
			},
		}
	}

	if logPipeline.Spec.EnableUnsupportedPlugins {
		warnMsg := "'enableUnsupportedPlugin' is enabled which would allow unsupported plugins to be used!"
		return admission.Response{
			AdmissionResponse: admissionv1.AdmissionResponse{
				Allowed:  true,
				Warnings: []string{warnMsg},
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

	configFiles, err := v.getFluentBitConfig(ctx, currentBaseDirectory, logPipeline)
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

	if err = v.variablesValidator.Validate(ctx, logPipeline, &logPipelines); err != nil {
		log.Error(err, "Failed to validate variables")
		return err
	}

	if err = v.pluginValidator.Validate(logPipeline, &logPipelines); err != nil {
		log.Error(err, "Failed to validate plugins")
		return err
	}

	if err = v.configValidator.Validate(ctx, fmt.Sprintf("%s/fluent-bit.conf", currentBaseDirectory)); err != nil {
		log.Error(err, "Failed to validate Fluent Bit config")
		return err
	}

	return nil
}

func (v *LogPipelineValidator) getFluentBitConfig(ctx context.Context, currentBaseDirectory string, logPipeline *telemetryv1alpha1.LogPipeline) ([]fs.File, error) {
	var configFiles []fs.File
	fluentBitSectionsConfigDirectory := currentBaseDirectory + "/dynamic"
	fluentBitParsersConfigDirectory := currentBaseDirectory + "/dynamic-parsers"
	fluentBitFilesDirectory := currentBaseDirectory + "/files"

	var generalCm corev1.ConfigMap
	if err := v.Get(ctx, v.fluentBitConfigMap, &generalCm); err != nil {
		return nil, err
	}
	for key, data := range generalCm.Data {
		configFiles = append(configFiles, fs.File{
			Path: currentBaseDirectory,
			Name: key,
			Data: data,
		})
	}

	for _, file := range logPipeline.Spec.Files {
		configFiles = append(configFiles, fs.File{
			Path: fluentBitFilesDirectory,
			Name: file.Name,
			Data: file.Content,
		})
	}

	sectionsConfig, err := fluentbit.MergeSectionsConfig(logPipeline, v.emitterConfig)
	if err != nil {
		return []fs.File{}, err
	}
	configFiles = append(configFiles, fs.File{
		Path: fluentBitSectionsConfigDirectory,
		Name: logPipeline.Name + ".conf",
		Data: sectionsConfig,
	})

	var logPipelines telemetryv1alpha1.LogPipelineList
	if err := v.List(ctx, &logPipelines); err != nil {
		return nil, err
	}
	parsersConfig := fluentbit.MergeParsersConfig(&logPipelines)
	configFiles = append(configFiles, fs.File{
		Path: fluentBitParsersConfigDirectory,
		Name: fluentBitParsersConfigMapKey,
		Data: parsersConfig,
	})

	return configFiles, nil
}

func (v *LogPipelineValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}

func (v *LogPipelineValidator) validateSecrets(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) bool {
	secVal := secret.NewSecretHelper(v.Client)
	return secVal.ValidateSecretsExist(ctx, logPipeline)
}
