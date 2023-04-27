package webhook

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-project/kyma/components/function-controller/internal/docker"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-v1-secret,mutating=true,failurePolicy=fail,groups="",resources=secrets,verbs=create;update,versions=v1,name=mutating.secret.serverless.k8s.io,sideEffects=none,admissionReviewVersions=v1,versions=v1

const (
	keyDockerConfigJSON = ".dockerconfigjson"
)

type RemoteRegistryCfgUpdater interface {
	Handle(context.Context, admission.Request) admission.Response
	InjectDecoder(*admission.Decoder) error
}

func NewRegistryWatcher(log *zap.SugaredLogger) RemoteRegistryCfgUpdater {
	return &registryWatcher{
		log: log,
	}
}

type registryWatcher struct {
	Decoder *admission.Decoder
	log     *zap.SugaredLogger
}

func (r *registryWatcher) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := r.log.With("name", req.Name, "namespace", req.Namespace, "kind", req.Kind.Kind)
	log.Debug("starting admission")

	var secret corev1.Secret
	if err := r.Decoder.Decode(req, &secret); err != nil {
		log.Debugf("failed to decode secret: %s", err.Error())
		return admission.Errored(http.StatusBadRequest, err)
	}

	registryCfgMarshaler, err := docker.NewRegistryCfgMarshaler(secret.Data)
	if err != nil {
		log.Debugf("failed to create marshaler: %s", err.Error())
		return admission.Errored(http.StatusInternalServerError, err)
	}

	secret.Data[keyDockerConfigJSON], err = registryCfgMarshaler.MarshalJSON()
	if err != nil {
		log.Debugf("failed to marshal: %s", err.Error())
		return admission.Errored(http.StatusInternalServerError, err)
	}

	res := patchResponseFromRaw(req.Object.Raw, &secret)
	log.Debug("admission finalized")
	return res
}

func patchResponseFromRaw(raw []byte, secret *corev1.Secret) admission.Response {
	current, err := json.Marshal(secret)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(raw, current)
}

func (r *registryWatcher) InjectDecoder(d *admission.Decoder) error {
	r.Decoder = d
	return nil
}
