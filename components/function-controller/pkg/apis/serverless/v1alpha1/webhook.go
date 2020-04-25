package v1alpha1

import (
	"context"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/webhook/resourcesemantics"
)

var _ resourcesemantics.GenericCRD = (*Function)(nil)

func (in *Function) SetDefaults(ctx context.Context) {
	one := int32(1)
	if in.Spec.MinReplicas == nil {
		in.Spec.MinReplicas = &one
	}
}

func (in *Function) Validate(ctx context.Context) *apis.FieldError {
	maxReplicas := in.Spec.MaxReplicas
	minReplicas := in.Spec.MinReplicas
	if maxReplicas != nil && minReplicas != nil && *maxReplicas < *minReplicas {
		return apis.ErrInvalidValue("maxReplicas is less then minReplicas", "spec.maxReplicas")
	}

	return nil
}
