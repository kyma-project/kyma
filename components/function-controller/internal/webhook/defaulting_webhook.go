package webhook

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"net/http"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type DefaultingWebHook struct {
	configAlphaV2 *serverlessv1alpha2.DefaultingConfig
	configAlphaV1 *serverlessv1alpha1.DefaultingConfig
	client        ctrlclient.Client
	decoder       *admission.Decoder
	log           *zap.SugaredLogger
}

func NewDefaultingWebhook(configV1Alpha1 *serverlessv1alpha1.DefaultingConfig, configV1Alpha2 *serverlessv1alpha2.DefaultingConfig, client ctrlclient.Client, log *zap.SugaredLogger) *DefaultingWebHook {
	return &DefaultingWebHook{
		configAlphaV1: configV1Alpha1,
		configAlphaV2: configV1Alpha2,
		client:        client,
		log:           log,
	}
}

func (w *DefaultingWebHook) Handle(_ context.Context, req admission.Request) admission.Response {
	log := w.log.With("name", req.Name, "namespace", req.Namespace, "kind", req.Kind.Kind)
	log.Debug("starting defaulting")

	if req.Kind.Kind == "Function" {
		res := w.handleFunctionDefaulting(req)
		log.Debug("defaulting finished for function")
		return res
	}
	if req.Kind.Kind == "GitRepository" {
		res := w.handleGitRepoDefaulting()
		log.Debug("defaulting finished for gitrepository")
		return res
	}

	log.Debug("request object invalid kind")
	return admission.Errored(http.StatusBadRequest, fmt.Errorf("invalid kind: %v", req.Kind.Kind))
}

func (w *DefaultingWebHook) InjectDecoder(decoder *admission.Decoder) error {
	w.decoder = decoder
	return nil
}

func (w *DefaultingWebHook) handleFunctionDefaulting(req admission.Request) admission.Response {
	var f interface{}
	switch req.Kind.Version {
	case serverlessv1alpha1.FunctionVersion:
		{
			fn := &serverlessv1alpha1.Function{}
			if err := w.decoder.Decode(req, fn); err != nil {
				return admission.Errored(http.StatusBadRequest, err)
			}
			fn.Default(w.configAlphaV1)
			f = fn
		}
	case serverlessv1alpha2.FunctionVersion:
		{
			fn := &serverlessv1alpha2.Function{}
			if err := w.decoder.Decode(req, fn); err != nil {
				return admission.Errored(http.StatusBadRequest, err)
			}
			fn.Default(w.configAlphaV2)
			f = fn
		}
	default:
		return admission.Errored(http.StatusBadRequest, errors.Errorf("Invalid resource version provided: %s", req.Kind.Version))
	}

	fBytes, err := json.Marshal(f)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, fBytes)
}

func (w *DefaultingWebHook) handleGitRepoDefaulting() admission.Response {
	return admission.Allowed("")
}
