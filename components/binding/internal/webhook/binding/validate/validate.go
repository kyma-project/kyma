package validate

import (
	"context"
	"github.com/kyma-project/kyma/components/binding/internal/webhook"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	admissionTypes "k8s.io/api/admission/v1beta1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ admission.Handler = &ValidationHandler{}
var _ admission.DecoderInjector = &ValidationHandler{}

type Validator interface {
	Validate(*v1alpha1.Binding) *webhook.Error
}

type ValidationHandler struct {
	decoder *admission.Decoder

	CreateValidators []Validator
	UpdateValidators []Validator

	log log.FieldLogger
}

func NewValidationHandler(log log.FieldLogger) *ValidationHandler {
	return &ValidationHandler{
		CreateValidators: []Validator{&StaticCreate{}},
		UpdateValidators: []Validator{},
		log:              log,
	}
}

func (h *ValidationHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
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

	var err *webhook.Error

	switch req.Operation {
	case admissionTypes.Create:
		for _, v := range h.CreateValidators {
			err = v.Validate(binding)
			if err != nil {
				break
			}
		}
	case admissionTypes.Update:
		for _, v := range h.UpdateValidators {
			err = v.Validate(binding)
			if err != nil {
				break
			}
		}
	default:
		h.log.Infof("Binding validation wehbook does not support action %q", req.Operation)
		return admission.Allowed("action not taken")
	}

	if err != nil {
		switch err.Code() {
		case http.StatusForbidden:
			return admission.Denied(err.Error())
		default:
			return admission.Errored(err.Code(), err)
		}
	}

	h.log.Infof("Completed successfully validation operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)
	return admission.Allowed("Binding validation successful")
}

func (h *ValidationHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// InjectClient injects the client into the handlers
func (h *ValidationHandler) InjectClient(c client.Client) error {
	for _, v := range h.CreateValidators {
		_, err := inject.ClientInto(c, v)
		if err != nil {
			return err
		}
	}
	for _, v := range h.UpdateValidators {
		_, err := inject.ClientInto(c, v)
		if err != nil {
			return err
		}
	}

	return nil
}
