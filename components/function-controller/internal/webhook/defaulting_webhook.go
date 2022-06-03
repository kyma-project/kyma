package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type DefaultingWebHook struct {
	config  *serverlessv1alpha1.DefaultingConfig
	client  ctrlclient.Client
	decoder *admission.Decoder
}

func NewDefaultingWebhook(config *serverlessv1alpha1.DefaultingConfig, client ctrlclient.Client) *DefaultingWebHook {
	return &DefaultingWebHook{
		config: config,
		client: client,
	}
}
func (w *DefaultingWebHook) Handle(_ context.Context, req admission.Request) admission.Response {
	if req.RequestKind.Kind == "Function" {
		return w.handleFunctionDefaulting(req)
	}
	if req.RequestKind.Kind == "GitRepository" {
		return w.handleGitRepoDefaulting()
	}
	return admission.Errored(http.StatusBadRequest, fmt.Errorf("invalid kind: %v", req.RequestKind.Kind))
}

func (w *DefaultingWebHook) InjectDecoder(decoder *admission.Decoder) error {
	w.decoder = decoder
	return nil
}

func (w *DefaultingWebHook) handleFunctionDefaulting(req admission.Request) admission.Response {
	f := &serverlessv1alpha1.Function{}
	if err := w.decoder.Decode(req, f); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	// apply defaults
	f.Default(w.config)

	fBytes, err := json.Marshal(f)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, fBytes)
}

func (w *DefaultingWebHook) handleGitRepoDefaulting() admission.Response {
	return admission.Allowed("")
}
