package v1alpha1

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"

	"knative.dev/pkg/apis"
)

func (in *GitRepository) Validate() error {
	return in.Spec.validate("spec")
}

func (in *GitRepositorySpec) validate(path string) error {
	if err := validateIfMissingFields(property{
		name:  fmt.Sprintf("%s.url", path),
		value: in.URL,
	}); err != nil {
		return errors.Wrap(err, "missing required fields: %v")
	}
	if isRepoURLIsSSH(in.URL) {
		if err := in.Auth.validateSSHAuth(fmt.Sprintf("%s.auth", path)); err != nil {
			return errors.Wrap(err, "invalid ssh auth")
		}
		return nil
	}

	if in.Auth == nil { // no-auth authentication method
		return nil
	}
	if err := in.Auth.validate(fmt.Sprintf("%s.auth", path)); err != nil {
		return errors.Wrap(err, "invalid auth")

	}
	return nil
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
	if in.Type != RepositoryAuthSSHKey {
		return fmt.Errorf("invalid value for git ssh, expected %s, current: %s",
			RepositoryAuthSSHKey, in.Type)
	}
	if err := validateIfMissingFields(property{
		name:  fmt.Sprintf("%s.secretName", path),
		value: in.SecretName,
	}); err != nil {
		return errors.Wrap(err, "missing required fields: %v")

	}
	return nil
}
func isRepoURLIsSSH(repoURL string) bool {
	exp, err := regexp.Compile(`((git|ssh?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)?`)
	if err != nil {
		panic(err)
	}

	return exp.MatchString(repoURL)
}
