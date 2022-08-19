package v1alpha1

import (
	"fmt"
	"regexp"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func (in *GitRepository) Validate() error {
	return in.Spec.validate("spec")
}

func (in *GitRepositorySpec) validate(path string) error {
	allErrs := []string{}
	if err := validateIfMissingFields(property{
		name:  fmt.Sprintf("%s.url", path),
		value: in.URL,
	}); err != nil {
		allErrs = append(allErrs, err.Error())
	}

	if isRepoURLIsSSH(in.URL) {
		if err := in.Auth.validateSSHAuth(fmt.Sprintf("%s.auth", path)); err != nil {
			return returnAllErrs("", append(allErrs, err.Error()))
		}
		return returnAllErrs("", allErrs)
	}

	if in.Auth == nil { // no-auth authentication method
		return returnAllErrs("", allErrs)
	}

	if err := in.Auth.validate(fmt.Sprintf("%s.auth", path)); err != nil {
		return returnAllErrs("", append(allErrs, err.Error()))
	}

	return returnAllErrs("", allErrs)
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
		return field.Required(field.NewPath(path), "missing required field")
	}
	allErrs := []string{}
	if in.Type != RepositoryAuthSSHKey {
		allErrs = append(allErrs, fmt.Sprintf("invalid value for spec.auth.type, expected %s, current: %s",
			RepositoryAuthSSHKey, in.Type))
	}
	if err := validateIfMissingFields(property{
		name:  fmt.Sprintf("%s.secretName", path),
		value: in.SecretName,
	}); err != nil {
		allErrs = append(allErrs, err.Error())
	}
	return returnAllErrs("", allErrs)
}
func isRepoURLIsSSH(repoURL string) bool {
	exp, err := regexp.Compile(`((git|ssh?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(/)?`)
	if err != nil {
		panic(err)
	}

	return exp.MatchString(repoURL)
}
