package binding

import (
	"context"
	"encoding/json"
	"github.com/kyma-project/kyma/components/binding/internal/webhook"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	admissionTypes "k8s.io/api/admission/v1beta1"
	"net/http"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ admission.Handler = &MutationHandler{}
var _ admission.DecoderInjector = &MutationHandler{}

type MutationHandler struct {
	decoder *admission.Decoder
	log     log.FieldLogger
}

func NewMutationHandler(log log.FieldLogger) *MutationHandler {
	return &MutationHandler{
		log: log,
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

	mutated := binding.DeepCopy()
	switch req.Operation {
	case admissionTypes.Create:
		if err := h.mutateOnCreate(ctx, mutated); err != nil {
			h.log.Errorf("cannot mutate Binding: %s", err)
			return admission.Errored(http.StatusInternalServerError, err)
		}
	default:
		log.Infof("Binding mutation wehbook does not support action %q", req.Operation)
		return admission.Allowed("action not taken")
	}

	rawBinding, err := json.Marshal(binding)
	if err != nil {
		h.log.Errorf("cannot marshal mutated binding: %s", err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	h.log.Infof("finish handling binding: %s", req.UID)
	return admission.PatchResponseFromRaw(req.Object.Raw, rawBinding)
}

func (h *MutationHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *MutationHandler) mutateOnCreate(ctx context.Context, binding *v1alpha1.Binding) error {
	if binding.Finalizers == nil {
		binding.Finalizers = make([]string, 0)
	}
	binding.Finalizers = append(binding.Finalizers, v1alpha1.BindingFinalizer)
	return nil
}
