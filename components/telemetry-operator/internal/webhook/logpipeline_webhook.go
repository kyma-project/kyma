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

	"github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fileutils"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-telemetry-kyma-project-io-v1alpha1-logpipeline,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.kyma-project.io,resources=logpipelines,verbs=create;update,versions=v1alpha1,name=vlogpipeline.kb.io,admissionReviewVersions=v1
type logPipelineValidator struct {
	client.Client
	fluentBitConfigMap types.NamespacedName
	decoder            *admission.Decoder
}

func NewLogPipeLineValidator(client client.Client, fluentBitConfigMap string, namespace string) *logPipelineValidator {
	return &logPipelineValidator{
		Client: client,
		fluentBitConfigMap: types.NamespacedName{
			Name:      fluentBitConfigMap,
			Namespace: namespace,
		},
	}
}

func (v *logPipelineValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := log.FromContext(ctx)

	logPipeline := &v1alpha1.LogPipeline{}
	if err := v.decoder.Decode(req, logPipeline); err != nil {
		// TODO log
		log.Error(err, "Failed to decode log pipeline")
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := v.validateLogPipeline(ctx, logPipeline); err != nil {
		return admission.Denied(err.Error())
	}

	return admission.Allowed("LogPipeline validation successful")
}

func (v *logPipelineValidator) validateLogPipeline(ctx context.Context, logPipeline *v1alpha1.LogPipeline) error {
	// Create or update existing fluentbit config
	var generalConfig corev1.ConfigMap
	if err := v.Get(ctx, v.fluentBitConfigMap, &generalConfig); err != nil {
		return err
	}

	sectionsConfig := fluentbit.MergeFluentBitConfig(logPipeline)

	var logPipelines v1alpha1.LogPipelineList
	if err := v.List(ctx, &logPipelines); err != nil {
		return err
	}
	parsersConfig := fluentbit.MergeFluentBitParsersConfig(&logPipelines)

	//filesConfig := logPipeline.Spec.Files  // TODO

	// write the fluentbit config file
	baseDirectory := "dry-run/"
	// 	base/fluent-bit.conf and base/custom_parsers.conf
	for key, data := range generalConfig.Data {
		err := fileutils.Write(baseDirectory, key, []byte(data))
		if err != nil {
			return err
		}
	}
	//	base/dynamic/dynamic.conf
	err := fileutils.Write(fmt.Sprintf("%s/dynamic", baseDirectory), fmt.Sprintf("%s.conf", logPipeline.Name), []byte(sectionsConfig))
	if err != nil {
		return err
	}
	//	base/dynamic-parsers/parsers.conf
	err = fileutils.Write(fmt.Sprintf("%s/dynamic-parsers", baseDirectory), "parsers.conf", []byte(parsersConfig))
	if err != nil {
		return err
	}

	//	base/parsers.conf

	// Write to the filesystem
	// write parses.conf
	// write to dynamic/

	// Validate it with dry run

	// Delete the fluentbit config file
	return nil
}

func (v *logPipelineValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
