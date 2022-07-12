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
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/validation"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
//var logparserlog = logf.Log.WithName("logparser-resource")

//func (r *v1alpha1.LogParser) SetupWebhookWithManager(mgr ctrl.Manager) error {
//	return ctrl.NewWebhookManagedBy(mgr).
//		For(r).
//		Complete()
//}

//// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
//
//// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//
//var _ webhook.Validator = &v1alpha1.LogParser{}
//
//// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
//func (r *v1alpha1.LogParser) ValidateCreate() error {
//	logparserlog.Info("validate create", "name", r.Name)
//
//	// TODO(user): fill in your validation logic upon object creation.
//	return nil
//}
//
//// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
//func (r *v1alpha1.LogParser) ValidateUpdate(old runtime.Object) error {
//	logparserlog.Info("validate update", "name", r.Name)
//
//	// TODO(user): fill in your validation logic upon object update.
//	return nil
//}
//
//// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
//func (r *v1alpha1.LogParser) ValidateDelete() error {
//	logparserlog.Info("validate delete", "name", r.Name)
//
//	// TODO(user): fill in your validation logic upon object deletion.
//	return nil
//}

//+kubebuilder:webhook:path=/validate-logparser,mutating=false,failurePolicy=fail,sideEffects=None,groups=telemetry.kyma-project.io,resources=logparsers,verbs=create;update,versions=v1alpha1,name=vlogparser.kb.io,admissionReviewVersions=v1
type LogparserValidator struct {
	client.Client

	fluentBitConfigMap types.NamespacedName
	parserValidator    validation.ParserValidator

	decoder *admission.Decoder
}

func NewLogParserValidator(client client.Client, fluentBitConfigMap string, namespace string, parserValidator validation.ParserValidator) *LogparserValidator {
	fmt.Println("aaaaaaaaaaaaaaaaaaaaaaaa")
	return &LogparserValidator{
		Client: client,
		fluentBitConfigMap: types.NamespacedName{
			Name:      fluentBitConfigMap,
			Namespace: namespace,
		},
		parserValidator: parserValidator,
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
	return v.validateLogParser(ctx, logParser)
}
