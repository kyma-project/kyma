package v1alpha1

import (
	"context"
	"fmt"
	"regexp"

	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"knative.dev/pkg/apis"
)

func (in *GitRepository) Validate(_ context.Context) error {
	return in.Spec.validate("spec")
}

func (in *GitRepositorySpec) validate(path string) error {
	var allErrs []error
	if err := validateIfMissingFields(property{
		name:  fmt.Sprintf("%s.url", path),
		value: in.URL,
	}); err != nil {
		allErrs = append(allErrs, err)
	}
	if isRepoURLIsSSH(in.URL) {
		if err := in.Auth.validateSSHAuth(fmt.Sprintf("%s.auth", path)); err != nil {
			allErrs = append(allErrs, err)
		}
		return errors.NewAggregate(allErrs)
	}

	if in.Auth == nil { // no-auth authentication method
		return errors.NewAggregate(allErrs)
	}
	if err := in.Auth.validate(fmt.Sprintf("%s.auth", path)); err != nil {
		allErrs = append(allErrs, err)
	}
	return errors.NewAggregate(allErrs)
}

func (in *RepositoryAuth) validate(path string) error {
	return validateIfMissingFields(property{
		name:  fmt.Sprintf("%s.secretName", path),
		value: in.SecretName,
	}, property{
		name:  fmt.Sprintf("%s.type", path),
		value: string(in.Type),
	})
}

func (in *RepositoryAuth) validateSSHAuth(path string) error {
	if in == nil {
		return apis.ErrMissingField(path)
	}
	var allErrs []error
	if in.Type != RepositoryAuthSSHKey {
		allErrs = append(allErrs,
			field.Invalid(
				field.NewPath(fmt.Sprintf("%s.type", path)),
				in.Type,
				fmt.Sprintf("invalid value for git ssh, expected %s, current: %s",
					RepositoryAuthSSHKey, in.Type),
			),
		)
	}
	if err := validateIfMissingFields(property{
		name:  fmt.Sprintf("%s.secretName", path),
		value: in.SecretName,
	}); err != nil {
		allErrs = append(allErrs, err)
	}
	return errors.NewAggregate(allErrs)
}
func isRepoURLIsSSH(repoURL string) bool {
	exp, err := regexp.Compile(`((git|ssh?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)?`)
	if err != nil {
		panic(err)
	}

	return exp.MatchString(repoURL)
}
