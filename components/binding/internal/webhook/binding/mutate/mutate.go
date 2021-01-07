package mutate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-project/kyma/components/binding/internal/storage"
	"github.com/kyma-project/kyma/components/binding/internal/webhook"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	log "github.com/sirupsen/logrus"
	admissionTypes "k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"net/http"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ admission.Handler = &MutationHandler{}
var _ admission.DecoderInjector = &MutationHandler{}

type MutationHandler struct {
	decoder     *admission.Decoder
	client      webhook.Client
	dc          dynamic.Interface
	log         log.FieldLogger
	kindStorage storage.KindStorage
}

func NewMutationHandler(kindStorage storage.KindStorage, client webhook.Client, dc dynamic.Interface, log log.FieldLogger) *MutationHandler {
	return &MutationHandler{
		kindStorage: kindStorage,
		client:      client,
		dc:          dc,
		log:         log,
	}
}

func (h *MutationHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	h.log.Infof("start handling binding: %s", req.UID)

	binding := &v1alpha1.Binding{}
	if err := webhook.MatchKinds(binding, req.Kind); err != nil {
		h.log.Errorf("kind does not match: %s", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := h.decoder.Decode(req, binding); err != nil {
		h.log.Errorf("cannot decode Binding: %s", err)
		return admission.Errored(http.StatusBadRequest, err)
	}
	h.log = h.log.WithField("binding", fmt.Sprintf("%s/%s", binding.Name, binding.Namespace))

	b := binding.DeepCopy()
	switch req.Operation {
	case admissionTypes.Create:
		h.mutateOnCreate(ctx, b)
	case admissionTypes.Update:
		h.mutateOnUpdate(ctx, b)
	default:
		log.Infof("Binding mutation webhook does not support action %q", req.Operation)
		return admission.Allowed("action not taken")
	}
	if reflect.DeepEqual(b, binding) {
		return admission.Allowed("action not taken")
	}

	rawBinding, err := json.Marshal(b)
	if err != nil {
		h.log.Errorf("cannot marshal mutated binding: %s", err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	h.log.Infof("finish handling binding: %s", req.UID)
	return admission.PatchResponseFromRaw(req.Object.Raw, rawBinding)
}

func (h *MutationHandler) mutateOnCreate(ctx context.Context, binding *v1alpha1.Binding) {
	h.injectFinalizer(binding)
	h.validateSpec(ctx, binding)
}

func (h *MutationHandler) mutateOnUpdate(ctx context.Context, binding *v1alpha1.Binding) {
	h.validateSpec(ctx, binding)
}

func (h *MutationHandler) validateSpec(ctx context.Context, binding *v1alpha1.Binding) {
	h.removeValidationLabel(binding)

	if err := h.validateSource(ctx, binding); err != nil {
		h.log.Warnf("while validating source: %d %s", err.Code(), err.Error())
		return
	}
	if err := h.validateTarget(ctx, binding); err != nil {
		h.log.Warnf("while validating target: %d %s", err.Code(), err.Error())
		return
	}
	h.injectValidationLabel(binding)
}

// checks if source of data to inject, exists on cluster
func (h *MutationHandler) validateSource(ctx context.Context, binding *v1alpha1.Binding) *webhook.Error {
	switch binding.Spec.Source.Kind {
	case v1alpha1.SourceKindConfigMap:
		_, err := h.client.FindConfigMap(ctx, binding)
		switch {
		case err == nil:
		case errors.IsNotFound(err):
			return webhook.NewErrorf(http.StatusBadRequest, "config map %s/%s not exist: %s", binding.Spec.Source.Name, binding.Namespace, err)
		default:
			return webhook.NewErrorf(http.StatusInternalServerError, "while getting config map: %s", err)
		}

	case v1alpha1.SourceKindSecret:
		_, err := h.client.FindSecret(ctx, binding)
		switch {
		case err == nil:
		case errors.IsNotFound(err):
			return webhook.NewErrorf(http.StatusBadRequest, "secret %s/%s not exist: %s", binding.Spec.Source.Name, binding.Namespace, err)
		default:
			return webhook.NewErrorf(http.StatusInternalServerError, "while getting secret: %s", err)
		}

	default:
		return webhook.NewErrorf(http.StatusBadRequest, "kind %s not exist", binding.Spec.Source.Kind)
	}
	return nil
}

// checks if source of data to inject, exists on cluster
func (h *MutationHandler) validateTarget(ctx context.Context, binding *v1alpha1.Binding) *webhook.Error {
	kind, err := h.kindStorage.Get(v1alpha1.Kind(binding.Spec.Target.Kind))
	if err != nil {
		return webhook.NewErrorf(http.StatusBadRequest, "kind %s not exist: %s", binding.Spec.Target.Kind, err)
	}
	_, err = h.dc.Resource(kind.Schema).Namespace(binding.Namespace).Get(ctx, binding.Spec.Target.Name, v1.GetOptions{})
	switch {
	case err == nil:
	case errors.IsNotFound(err):
		return webhook.NewErrorf(http.StatusBadRequest, "resource %s of kind %s is not found", binding.Spec.Target.Name, binding.Spec.Target.Kind)
	default:
		return webhook.NewErrorf(http.StatusInternalServerError, "while getting resource: %s", err)
	}
	return nil
}

// finalizer prevents k8s from the binding deletion
func (h *MutationHandler) injectFinalizer(binding *v1alpha1.Binding) {
	if binding.Finalizers == nil {
		binding.Finalizers = make([]string, 0)
	}
	binding.Finalizers = append(binding.Finalizers, v1alpha1.BindingFinalizer)
	h.log.Infof("Applied %s finalizer to binding", v1alpha1.BindingFinalizer)
}

// if validation label exists on the binding it means that error happened during the validation
func (h *MutationHandler) injectValidationLabel(binding *v1alpha1.Binding) {
	if binding.Labels == nil {
		binding.Labels = make(map[string]string, 0)
	}
	binding.Labels[v1alpha1.BindingValidatedLabelKey] = "true"
	h.log.Infof("Applied %s label to binding", v1alpha1.BindingValidatedLabelKey)
}

// if validation label exists on the binding it means that error happened during the validation
func (h *MutationHandler) removeValidationLabel(binding *v1alpha1.Binding) {
	if binding.Labels == nil {
		return
	}
	delete(binding.Labels, v1alpha1.BindingValidatedLabelKey)
	h.log.Infof("Removed %s label from binding", v1alpha1.BindingValidatedLabelKey)
}

func (h *MutationHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}
