package director

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/internal"
)

// NothingDoerService is a wrapper around the Director GraphQL client to support business logic
type NothingDoerService struct {
	directorCli              APIPackageInstanceAuthClient
	operationPollingInterval time.Duration
}

// NewNothingDoerService returns new NothingDoerService instance
func NewNothingDoerService() *NothingDoerService {
	return &NothingDoerService{}
}

// EnsureAPIPackageCredentials ensures that the API Package credentials were created for the given service instance ID
func (s *NothingDoerService) EnsureAPIPackageCredentials(_ context.Context, _, _, _ string, _ map[string]interface{}) error {
	return nil
}

// GetAPIPackageCredentials returns API Package credentials associated with the given service instance id
func (s *NothingDoerService) GetAPIPackageCredentials(_ context.Context, _, _, _ string) (internal.APIPackageCredential, error) {
	return internal.APIPackageCredential{}, nil
}

// EnsureAPIPackageCredentialsDeleted ensures that the given API Package credential associated with the given service instance id
// is removed.
func (s *NothingDoerService) EnsureAPIPackageCredentialsDeleted(_ context.Context, _ string, _ string, _ string) error {
	return nil
}
