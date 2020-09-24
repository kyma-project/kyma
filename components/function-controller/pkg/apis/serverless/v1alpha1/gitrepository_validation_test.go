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
					URL: "some_url",
				},
			},
			expectedError: gomega.BeNil(),
		},
		"should be valid - auth": {
			givenFunc: GitRepository{
				Spec: GitRepositorySpec{
					URL: "some_url",
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
