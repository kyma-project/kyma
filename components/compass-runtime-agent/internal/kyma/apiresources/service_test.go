package apiresources

import (
	"testing"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	k8sconstsmocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/k8sconsts/mocks"
	accessservicemock "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/accessservice/mocks"
	secretmock "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets/mocks"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets/model"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
)

func TestService(t *testing.T) {

	t.Run("should create API resources", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}

		credentials := model.CredentialsWithCSRF{
			Basic: &model.Basic{
				Username: "admin",
				Password: "nimda",
			},
		}

		accessServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		secretServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", mock.MatchedBy(getCredentialsMatcher(&credentials))).Return(applications.Credentials{}, nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")
		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver)

		err := service.CreateApiResources("appName", types.UID("appUUID"), "serviceID", &credentials, nil)

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
	})

	t.Run("should not create secret if credentials not provided", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}

		accessServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver)

		err := service.CreateApiResources("appName", types.UID("appUUID"), "serviceID", nil, nil)

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertNotCalled(t, "Create")
	})

	t.Run("should not interrupt execution when error occurs on creation", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}

		credentials := model.CredentialsWithCSRF{
			Basic: &model.Basic{
				Username: "admin",
				Password: "nimda",
			},
		}
		accessServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(apperrors.Internal("some error"))
		secretServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", &credentials).Return(applications.Credentials{}, apperrors.Internal("some other error"))
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver)

		err := service.CreateApiResources("appName", types.UID("appUUID"), "serviceID", &credentials, nil)

		// then
		require.Error(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
	})

	t.Run("should update API resources", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}

		credentials := model.CredentialsWithCSRF{
			Basic: &model.Basic{
				Username: "admin",
				Password: "nimda",
			},
		}

		accessServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		secretServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", mock.MatchedBy(getCredentialsMatcher(&credentials))).Return(applications.Credentials{}, nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver)

		err := service.UpdateApiResources("appName", types.UID("appUUID"), "serviceID", &credentials, nil)

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
	})

	t.Run("should not update secret if credentials not provided", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}

		accessServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		secretServiceMock.On("Delete", "secretName").Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")
		nameResolver.On("GetCredentialsSecretName", "appName", "serviceID").Return("secretName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver)

		err := service.UpdateApiResources("appName", types.UID("appUUID"), "serviceID", nil, nil)

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertNotCalled(t, "Upsert")
	})

	t.Run("should not interrupt execution when error occurs on update with credentials", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}

		credentials := model.CredentialsWithCSRF{
			Basic: &model.Basic{
				Username: "admin",
				Password: "nimda",
			},
		}
		accessServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(apperrors.Internal("some error"))
		secretServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", &credentials).Return(applications.Credentials{}, apperrors.Internal("some other error"))
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver)

		err := service.UpdateApiResources("appName", types.UID("appUUID"), "serviceID", &credentials, nil)

		// then
		require.Error(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
	})

	t.Run("should not interrupt execution when error occurs on update without credentials", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}

		accessServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(apperrors.Internal("some error"))
		secretServiceMock.On("Delete", "secretName").Return(apperrors.Internal("some other error"))
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")
		nameResolver.On("GetCredentialsSecretName", "appName", "serviceID").Return("secretName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver)

		err := service.UpdateApiResources("appName", types.UID("appUUID"), "serviceID", nil, nil)

		// then
		require.Error(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
	})

	t.Run("should delete API resources with credentials", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}

		accessServiceMock.On("Delete", "resourceName").Return(nil)
		secretServiceMock.On("Delete", "secretName").Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver)

		err := service.DeleteApiResources("appName", "serviceID", "secretName")

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
	})

	t.Run("should delete API resources without credentials", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}

		accessServiceMock.On("Delete", "resourceName").Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver)

		err := service.DeleteApiResources("appName", "serviceID", "")

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
	})

	t.Run("should not interrupt execution when error occurs on delete", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}

		accessServiceMock.On("Delete", "resourceName").Return(apperrors.Internal("some error"))
		secretServiceMock.On("Delete", "secretName").Return(apperrors.Internal("some error"))
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver)

		err := service.DeleteApiResources("appName", "serviceID", "secretName")

		// then
		require.Error(t, err)
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
