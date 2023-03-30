package webhook

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	v1 "k8s.io/api/admission/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type ValidatingWebHook struct {
	configv1alpha2 *serverlessv1alpha2.ValidationConfig
	client         ctrlclient.Client
	decoder        *admission.Decoder
	log            *zap.SugaredLogger
}

func NewValidatingWebhook(configV1alpha2 *serverlessv1alpha2.ValidationConfig, client ctrlclient.Client, log *zap.SugaredLogger) *ValidatingWebHook {
	return &ValidatingWebHook{
		configv1alpha2: configV1alpha2,
		client:         client,
		log:            log,
	}
}
func (w *ValidatingWebHook) Handle(_ context.Context, req admission.Request) admission.Response {
	log := w.log.With("name", req.Name, "namespace", req.Namespace, "kind", req.Kind.Kind)
	log.Debug("starting validation")

	// We don't currently have any delete validation logic
	if req.Operation == v1.Delete {
		res := admission.Allowed("")
		log.Debug("validation finished for deletion")
		return res
	}

	if req.Kind.Kind == "Function" {
		res := w.handleFunctionValidation(req)
		log.Debug("validation finished for function")
		return res
	}

	log.Debug("request object invalid kind")
	return admission.Errored(http.StatusBadRequest, fmt.Errorf("invalid kind: %v", req.Kind.Kind))
}

func (w *ValidatingWebHook) InjectDecoder(decoder *admission.Decoder) error {
	w.decoder = decoder
	return nil
}

func (w *ValidatingWebHook) handleFunctionValidation(req admission.Request) admission.Response {
	switch req.Kind.Version {
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
