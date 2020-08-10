package git_test

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	"github.com/onsi/gomega"
)

func TestAuthOptions_ToAuthMethod(t *testing.T) {
	// given
	for testName, testData := range map[string]struct {
		authType    git.RepositoryAuthType
		credentials map[string]string

		expectedAuthMethod gomega.OmegaMatcher
		expectedErr        gomega.OmegaMatcher
	}{
		"should be ok when basic": {
			authType: git.RepositoryAuthBasic,
			credentials: map[string]string{
				git.UsernameKey: "user",
				git.PasswordKey: "password",
			},
			expectedAuthMethod: gomega.Equal(&http.BasicAuth{
				Username: "user",
				Password: "password",
			}),
			expectedErr: gomega.BeNil(),
		},
		"error when invalid auth type": {
			authType: "invalid",
			credentials: map[string]string{
				git.UsernameKey: "user",
				git.PasswordKey: "password",
			},
			expectedAuthMethod: gomega.BeNil(),
			expectedErr:        gomega.HaveOccurred(),
		},
		"error when invalid key format": {
			authType: git.RepositoryAuthSSHKey,
			credentials: map[string]string{
				git.KeyKey: "invalid format",
			},
			expectedAuthMethod: gomega.BeNil(),
			expectedErr:        gomega.HaveOccurred(),
		},
		"error when missing fields in basic auth": {
			authType:           git.RepositoryAuthBasic,
			credentials:        map[string]string{},
			expectedAuthMethod: gomega.BeNil(),
			expectedErr:        gomega.HaveOccurred(),
		},
		"error when missing fields in key auth": {
			authType:           git.RepositoryAuthSSHKey,
			credentials:        map[string]string{},
			expectedAuthMethod: gomega.BeNil(),
			expectedErr:        gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			g := gomega.NewWithT(t)
			options := git.AuthOptions{
				Type:        testData.authType,
				Credentials: testData.credentials,
			}

			// when
			result, err := options.ToAuthMethod()

			//then
			g.Expect(result).To(testData.expectedAuthMethod)
			g.Expect(err).To(testData.expectedErr)
		})
	}

	t.Run("should return nil when AuthOptions is nil", func(t *testing.T) {
		// given
		g := gomega.NewWithT(t)
		var auth *git.AuthOptions

		// when
		result, err := auth.ToAuthMethod()

		// then
		g.Expect(result).To(gomega.BeNil())
		g.Expect(err).To(gomega.BeNil())

	})
}
