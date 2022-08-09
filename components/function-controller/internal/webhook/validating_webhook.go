package webhook

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	v1 "k8s.io/api/admission/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type ValidatingWebHook struct {
	configv1alpha1 *serverlessv1alpha1.ValidationConfig
	configv1alpha2 *serverlessv1alpha2.ValidationConfig
	client         ctrlclient.Client
	decoder        *admission.Decoder
}

func NewValidatingHook(configv1alpha1 *serverlessv1alpha1.ValidationConfig, configv1alpha2 *serverlessv1alpha2.ValidationConfig, client ctrlclient.Client) *ValidatingWebHook {
	return &ValidatingWebHook{
		configv1alpha1: configv1alpha1,
		configv1alpha2: configv1alpha2,
		client:         client,
	}
}
func (w *ValidatingWebHook) Handle(_ context.Context, req admission.Request) admission.Response {
	// We don't currently have any delete validation logic
	if req.Operation == v1.Delete {
		return admission.Allowed("")
	}

	if req.Kind.Kind == "Function" {
		return w.handleFunctionValidation(req)
	}

	return admission.Errored(http.StatusBadRequest, fmt.Errorf("invalid kind: %v", req.RequestKind.Kind))
}

func (w *ValidatingWebHook) InjectDecoder(decoder *admission.Decoder) error {
	w.decoder = decoder
	return nil
}

func (w *ValidatingWebHook) handleFunctionValidation(req admission.Request) admission.Response {
	switch req.Kind.Version {
	case serverlessv1alpha1.FunctionVersion:
		{
			fn := &serverlessv1alpha1.Function{}
			if err := w.decoder.Decode(req, fn); err != nil {
				return admission.Errored(http.StatusBadRequest, err)
			}
			if err := fn.Validate(w.configv1alpha1); err != nil {
				return admission.Denied(fmt.Sprintf("validation failed: %s", err.Error()))
			}
		}
	case serverlessv1alpha2.FunctionVersion:
		{
			fn := &serverlessv1alpha2.Function{}
			if err := w.decoder.Decode(req, fn); err != nil {
				return admission.Errored(http.StatusBadRequest, err)
			}
			if err := fn.Validate(w.configv1alpha2); err != nil {
				return admission.Denied(fmt.Sprintf("validation failed: %s", err.Error()))
			}
		}
	default:
		return admission.Errored(http.StatusBadRequest, errors.Errorf("Invalid resource version provided: %s", req.Kind.Version))
	}
	return admission.Allowed("")
}
