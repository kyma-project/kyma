package git_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	"github.com/onsi/gomega"
)

func TestAuthOptions_ToAuthMethod(t *testing.T) {
	// given
	for testName, testData := range map[string]struct {
		authType    git.RepositoryAuthType
		credentials map[string]string

		expectedCallback  gomega.OmegaMatcher
		expectedCertCheck gomega.OmegaMatcher
		expectedErr       gomega.OmegaMatcher
	}{
		"should be ok when basic": {
			authType: git.RepositoryAuthBasic,
			credentials: map[string]string{
				git.UsernameKey: "user",
				git.PasswordKey: "password",
			},
			expectedCallback:  gomega.Not(gomega.BeNil()),
			expectedCertCheck: gomega.BeNil(),
			expectedErr:       gomega.BeNil(),
		},
		"should be ok when ssh without passphrase": {
			authType: git.RepositoryAuthSSHKey,
			credentials: map[string]string{
				git.KeyKey: testSSHPrivateKey,
			},
			expectedCallback:  gomega.Not(gomega.BeNil()),
			expectedCertCheck: gomega.Not(gomega.BeNil()),
			expectedErr:       gomega.BeNil(),
		},
		"should be ok when ssh with passphrase": {
			authType: git.RepositoryAuthSSHKey,
			credentials: map[string]string{
				git.PasswordKey: "test",
				git.KeyKey:      testSSHPrivateKeyPassphrase,
			},
			expectedCallback:  gomega.Not(gomega.BeNil()),
			expectedCertCheck: gomega.Not(gomega.BeNil()),
			expectedErr:       gomega.BeNil(),
		},
		"error when invalid auth type": {
			authType: "invalid",
			credentials: map[string]string{
				git.UsernameKey: "user",
				git.PasswordKey: "password",
			},
			expectedCallback:  gomega.BeNil(),
			expectedCertCheck: gomega.BeNil(),
			expectedErr:       gomega.HaveOccurred(),
		},
		"error when invalid key format": {
			authType: git.RepositoryAuthSSHKey,
			credentials: map[string]string{
				git.KeyKey: "invalid format",
			},
			expectedCallback:  gomega.BeNil(),
			expectedCertCheck: gomega.BeNil(),
			expectedErr:       gomega.HaveOccurred(),
		},
		"error when missing field username in basic auth": {
			authType:          git.RepositoryAuthBasic,
			credentials:       map[string]string{},
			expectedCallback:  gomega.BeNil(),
			expectedCertCheck: gomega.BeNil(),
			expectedErr:       gomega.HaveOccurred(),
		},
		"error when missing field password in basic auth": {
			authType: git.RepositoryAuthBasic,
			credentials: map[string]string{
				git.UsernameKey: "test",
			},
			expectedCallback:  gomega.BeNil(),
			expectedCertCheck: gomega.BeNil(),
			expectedErr:       gomega.HaveOccurred(),
		},
		"error when missing fields in key auth": {
			authType:          git.RepositoryAuthSSHKey,
			credentials:       map[string]string{},
			expectedCallback:  gomega.BeNil(),
			expectedCertCheck: gomega.BeNil(),
			expectedErr:       gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			g := gomega.NewWithT(t)
			options := git.AuthOptions{
				Type:        testData.authType,
				Credentials: testData.credentials,
			}

			// when
			result, err := git.GetAuth(&options)

			//then
			g.Expect(err).To(testData.expectedErr)
			g.Expect(result.CredentialsCallback).To(testData.expectedCallback)
		})
	}

	t.Run("should return nil when AuthOptions is nil", func(t *testing.T) {
		// given
		g := gomega.NewWithT(t)
		var authOptions *git.AuthOptions

		// when
		result, err := git.GetAuth(authOptions)

		// then
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.CredentialsCallback).To(gomega.BeNil())
		g.Expect(result.CertificateCheckCallback).To(gomega.BeNil())

	})
}

func TestIsAuthErr(t *testing.T) {
	// GIVEN
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "err is nil",
			err:      nil,
			expected: false,
		},
		{
			name:     "different error",
			err:      errors.New("Internet Server Error"),
			expected: false,
		},
		{
			name:     "err contains 403 error code",
			err:      errors.New("nobody expected unexpected http status code: 403 while doing inquisition"),
			expected: true,
		},
		{
			name:     "error contains too many redirects",
			err:      errors.New("too many redirects or authentication replays while cloning repository"),
			expected: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//WHEN
			result := git.IsAuthErr(testCase.err)

			//THEN
			require.Equal(t, testCase.expected, result)
		})

	}

}
