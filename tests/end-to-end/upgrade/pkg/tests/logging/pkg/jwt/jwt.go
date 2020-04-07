package jwt

import (
	"fmt"

	dex "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/fetch-dex-token"
)

// GetToken retrieves jwt token from dex package
func GetToken(idpConfig dex.IdProviderConfig) (string, error) {
	return dex.Authenticate(idpConfig)
}

// SetAuthHeader sets authorization header with JWT
func SetAuthHeader(token string) string {
	return fmt.Sprintf("Authorization: Bearer %s", token)
}
