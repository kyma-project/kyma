package serverless

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-project/kyma/components/function-controller/pkg/docker"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-v1-secret,mutating=true,failurePolicy=fail,groups="",resources=secrets,verbs=create;update,versions=v1,name=mpod.kb.io

const (
	keyDockerConfigJSON = ".dockerconfigjson"
)

type RegistryWatcher interface {
	Handle(context.Context, admission.Request) admission.Response
	InjectDecoder(*admission.Decoder) error
}

func NewRegistryWatcher(c client.Client) RegistryWatcher {
	return &registryWatcher{}
}

type registryWatcher struct {
	Decoder *admission.Decoder
}

func (r *registryWatcher) Handle(ctx context.Context, req admission.Request) admission.Response {
	var secret corev1.Secret
	if err := r.Decoder.Decode(req, &secret); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	registryCfgMarshaler, err := docker.NewRegistryCfgMarshaler(secret.Data)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	secret.Data[keyDockerConfigJSON], err = registryCfgMarshaler.MarshalJSON()
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return patchResponseFromRaw(req.Object.Raw, &secret)
}

func patchResponseFromRaw(raw []byte, secret *corev1.Secret) admission.Response {
	current, err := json.Marshal(secret)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(raw, current)
}

func (a *registryWatcher) InjectDecoder(d *admission.Decoder) error {
	a.Decoder = d
	return nil
}
