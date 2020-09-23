package v1alpha1

import (
	"context"
	"fmt"

	"knative.dev/pkg/apis"
)

func (in *GitRepository) Validate(_ context.Context) *apis.FieldError {
	return in.Spec.validate("spec")
}

func (in *GitRepositorySpec) validate(path string) *apis.FieldError {
	err := validateIfMissingFields(property{
		name:  fmt.Sprintf("%s.url", path),
		value: in.URL,
	})

	if in.Auth == nil { // no-auth authentication method
		return err
	}

	return err.Also(in.Auth.validate(fmt.Sprintf("%s.auth", path)))
}

func (in *RepositoryAuth) validate(path string) *apis.FieldError {
	return validateIfMissingFields(property{
		name:  fmt.Sprintf("%s.secretName", path),
		value: in.SecretName,
	}, property{
		name:  fmt.Sprintf("%s.type", path),
		value: string(in.Type),
	})
}
