package webhook

import (
	"context"
	"fmt"
	"net/http"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	v1 "k8s.io/api/admission/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type ValidatingWebHook struct {
	config  *serverlessv1alpha2.ValidationConfig
	client  ctrlclient.Client
	decoder *admission.Decoder
}

func NewValidatingHook(config *serverlessv1alpha2.ValidationConfig, client ctrlclient.Client) *ValidatingWebHook {
	return &ValidatingWebHook{
		config: config,
		client: client,
	}
}
func (w *ValidatingWebHook) Handle(_ context.Context, req admission.Request) admission.Response {
	// We don't currently have any delete validation logic
	if req.Operation == v1.Delete {
		return admission.Allowed("")
	}

	if req.RequestKind.Kind == "Function" {
		return w.handleFunctionValidation(req)
	}

	return admission.Errored(http.StatusBadRequest, fmt.Errorf("invalid kind: %v", req.RequestKind.Kind))
}

func (w *ValidatingWebHook) InjectDecoder(decoder *admission.Decoder) error {
	w.decoder = decoder
	return nil
}

func (w *ValidatingWebHook) handleFunctionValidation(req admission.Request) admission.Response {
	f := &serverlessv1alpha2.Function{}
	if err := w.decoder.Decode(req, f); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := f.Validate(w.config); err != nil {
		return admission.Denied(fmt.Sprintf("validation failed: %s", err.Error()))
	}
	return admission.Allowed("")
}
