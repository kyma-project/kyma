package v1alpha1

import (
	"context"
	"fmt"
	"regexp"

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

	if isRepoURLIsSSH(in.URL) {
		return err.Also(in.Auth.validateSSHAuth(fmt.Sprintf("%s.auth", path)))
	}

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

func (in *RepositoryAuth) validateSSHAuth(path string) *apis.FieldError {
	if in == nil {
		return apis.ErrMissingField(path)
	}
	var err *apis.FieldError
	if in.Type != RepositoryAuthSSHKey {
		err = apis.ErrGeneric(fmt.Sprintf("invalid value for git ssh, expected %s, current: %s",
			RepositoryAuthSSHKey, in.Type), fmt.Sprintf("%s.type", path))
	}

	return err.Also(validateIfMissingFields(property{
		name:  fmt.Sprintf("%s.secretName", path),
		value: in.SecretName,
	}))
}
func isRepoURLIsSSH(repoURL string) bool {
	exp, err := regexp.Compile(`((git|ssh?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)?`)
	if err != nil {
		panic(err)
	}

	return exp.MatchString(repoURL)
}
