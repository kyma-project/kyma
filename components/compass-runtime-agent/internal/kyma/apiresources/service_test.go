package apiresources

import (
	accessservicemock "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/accessservice/mocks"
	secretmock "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets/mocks"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets/model"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestService(t *testing.T) {

	t.Run("should create API resources", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}

		credentials := model.CredentialsWithCSRF{
			Basic: &model.Basic{
				Username: "admin",
				Password: "nimda",
			},
		}

		accessServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "serviceName").Return(nil)
		secretServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", mock.MatchedBy(getCredentialsMatcher(&credentials))).Return(applications.Credentials{}, nil)

		// when
		service := NewService(accessServiceMock, secretServiceMock)

		err := service.CreateApiResources("appName", types.UID("appUUID"), "serviceID", "serviceName", &credentials, nil)

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
	})
}

func getCredentialsMatcher(expected *model.CredentialsWithCSRF) func(*model.CredentialsWithCSRF) bool {
	return func(credentials *model.CredentialsWithCSRF) bool {
		if credentials == nil {
			return expected == nil
		}

		if expected == nil {
			return credentials == nil
		}

		if credentials.Basic != nil && expected.Basic != nil {
			matched := credentials.Basic.Username == expected.Basic.Username && credentials.Basic.Password == expected.Basic.Password
			if !matched {
				return false
			}
		}

		if credentials.Oauth != nil && expected.Oauth != nil {
			matched := credentials.Oauth.ClientID == expected.Oauth.ClientID && credentials.Oauth.ClientSecret == expected.Oauth.ClientSecret
			if !matched {
				return false
			}
		}

		if credentials.CSRFInfo != nil && expected.CSRFInfo != nil {
			return credentials.CSRFInfo.TokenEndpointURL == expected.CSRFInfo.TokenEndpointURL
		}

		return true
	}
}
