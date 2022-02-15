package v1alpha1

import (
	"context"
	"testing"

	"github.com/onsi/gomega"
)

func TestGitRepositoryValidation(t *testing.T) {
	for testName, testData := range map[string]struct {
		givenFunc              GitRepository
		expectedError          gomega.OmegaMatcher
		specifiedExpectedError gomega.OmegaMatcher
	}{
		"should be valid - no auth": {
			givenFunc: GitRepository{
				Spec: GitRepositorySpec{
					URL: "https://github.com/kyma-project/kyma.git",
				},
			},
			expectedError: gomega.BeNil(),
		},
		"should be valid - auth": {
			givenFunc: GitRepository{
				Spec: GitRepositorySpec{
					URL: "https://github.com/kyma-project/kyma.git",
					Auth: &RepositoryAuth{
						Type:       RepositoryAuthBasic,
						SecretName: "some_name",
					},
				},
			},
			expectedError: gomega.BeNil(),
		},
		"should be invalid - empty URL and SecretName": {
			givenFunc: GitRepository{
				Spec: GitRepositorySpec{
					URL: "",
					Auth: &RepositoryAuth{
						Type:       RepositoryAuthBasic,
						SecretName: "",
					},
				},
			},
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring("spec.url"),
				gomega.ContainSubstring("spec.auth.secretName"),
			),
			expectedError: gomega.HaveOccurred(),
		},
		"should be invalid - missing URL and SecretName": {
			givenFunc: GitRepository{
				Spec: GitRepositorySpec{
					Auth: &RepositoryAuth{
						Type: RepositoryAuthSSHKey,
					},
				},
			},
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring("spec.url"),
				gomega.ContainSubstring("spec.auth.secretName"),
			),
			expectedError: gomega.HaveOccurred(),
		},
		"should be invalid - missing Type": {
			givenFunc: GitRepository{
				Spec: GitRepositorySpec{
					URL: "some_url",
					Auth: &RepositoryAuth{
						SecretName: "some_name",
					},
				},
			},
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring("spec.auth.type"),
			),
			expectedError: gomega.HaveOccurred(),
		},
		"should be valid git ssh": {
			givenFunc: GitRepository{
				Spec: GitRepositorySpec{
					URL: "git@github.com:kyma-project/kyma.git",
					Auth: &RepositoryAuth{
						Type:       RepositoryAuthSSHKey,
						SecretName: "my-secret",
					},
				},
			},
			expectedError: gomega.BeNil(),
		},
		"should be invalid git ssh, no auth provided": {
			givenFunc: GitRepository{
				Spec: GitRepositorySpec{
					URL: "git@github.com:kyma-project/kyma.git",
				},
			},
			specifiedExpectedError: gomega.ContainSubstring("spec.auth"),
			expectedError:          gomega.HaveOccurred(),
		},
		"should be invalid git ssh, auth type is not key and secret name is empty": {
			givenFunc: GitRepository{
				Spec: GitRepositorySpec{
					URL: "git@github.com:kyma-project/kyma.git",
					Auth: &RepositoryAuth{
						Type: RepositoryAuthBasic,
					},
				},
			},
			specifiedExpectedError: gomega.And(
				gomega.ContainSubstring("spec.auth.type"),
				gomega.ContainSubstring("spec.auth.secretName"),
				gomega.ContainSubstring("invalid value for git ssh")),
			expectedError: gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)

			// when
			err := testData.givenFunc.Validate(context.Background())
			g.Expect(err).To(testData.expectedError)
			if testData.specifiedExpectedError != nil {
				g.Expect(err.Error()).To(testData.specifiedExpectedError)
			}
		})
	}
}

func TestSSHRegex(t *testing.T) {
	for testName, testData := range map[string]struct {
		givenURL string
		isSSH    bool
	}{
		"should success": {
			givenURL: "git@github.com:kyma-project/kyma.git",
			isSSH:    true,
		},
		"should success with protocol": {
			givenURL: "ssh://ssh@github.com/test.git",
			isSSH:    true,
		},
		"should success with ~": {
			givenURL: "ssh://user@host.xz/~user/path/to/repo.git/",
			isSSH:    true,
		},
		"should success with git protocol": {
			givenURL: "git://host.xz/path/to/repo.git/",
			isSSH:    true,
		},
		"should success with port": {
			givenURL: "ssh://ssh@github.com:2500/test.git",
			isSSH:    true,
		},
		"should not success": {
			givenURL: "https://github.com/kyma-project/kyma.git",
			isSSH:    false,
		},
		"should not success with @": {
			givenURL: "https://fix@me.plz.git",
			isSSH:    false,
		},
	} {
		t.Run(testName, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)

			output := isRepoURLIsSSH(testData.givenURL)

			//THEN
			g.Expect(output).Should(gomega.Equal(testData.isSSH))
		})
	}
}
