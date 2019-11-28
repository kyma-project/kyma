package apiresources

import (
	"testing"

	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/assetstore/docstopic"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	k8sconstsmocks "kyma-project.io/compass-runtime-agent/internal/k8sconsts/mocks"
	accessservicemock "kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/accessservice/mocks"
	assetstoremock "kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/assetstore/mocks"
	istiomocks "kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/istio/mocks"
	secretmock "kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/secrets/mocks"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/secrets/model"
)

func TestService(t *testing.T) {

	jsonApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	eventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")

	specFormatJSON := docstopic.SpecFormatJSON

	t.Run("should create API resources", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		assetStoreMock := assetstoremock.Service{}

		credentials := model.CredentialsWithCSRF{
			Basic: &model.Basic{
				Username: "admin",
				Password: "nimda",
			},
		}

		accessServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		istioServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		secretServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", mock.MatchedBy(getCredentialsMatcher(&credentials))).Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")
		assetStoreMock.On("Put", "serviceID", docstopic.OpenApiType, jsonApiSpec, specFormatJSON, docstopic.ApiSpec).Return(nil)

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, assetStoreMock)

		err := service.CreateApiResources("appName", types.UID("appUUID"), "serviceID", &credentials, jsonApiSpec, specFormatJSON, docstopic.OpenApiType)

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		assetStoreMock.AssertExpectations(t)
	})

	t.Run("should create Event API resources", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		assetStoreMock := assetstoremock.Service{}

		assetStoreMock.On("Put", "serviceID", docstopic.AsyncApi, eventsSpec, specFormatJSON, docstopic.EventApiSpec).Return(nil)

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, assetStoreMock)

		err := service.CreateEventApiResources("appName", "serviceID", eventsSpec, specFormatJSON, docstopic.AsyncApi)

		// then
		require.NoError(t, err)
		assetStoreMock.AssertExpectations(t)
	})

	t.Run("should not create secret if credentials not provided", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		assetStoreMock := assetstoremock.Service{}

		accessServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		istioServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")
		assetStoreMock.On("Put", "serviceID", docstopic.OpenApiType, jsonApiSpec, specFormatJSON, docstopic.ApiSpec).Return(nil)

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, assetStoreMock)

		err := service.CreateApiResources("appName", types.UID("appUUID"), "serviceID", nil, jsonApiSpec, specFormatJSON, docstopic.OpenApiType)

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertNotCalled(t, "Create")
		istioServiceMock.AssertExpectations(t)
		assetStoreMock.AssertExpectations(t)
	})

	t.Run("should not interrupt execution when error occurs on creation", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		assetStoreMock := assetstoremock.Service{}

		credentials := model.CredentialsWithCSRF{
			Basic: &model.Basic{
				Username: "admin",
				Password: "nimda",
			},
		}
		accessServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(apperrors.Internal("some error"))
		istioServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(apperrors.Internal("just another error"))
		secretServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", &credentials).Return(apperrors.Internal("some other error"))
		assetStoreMock.On("Put", "serviceID", docstopic.OpenApiType, jsonApiSpec, specFormatJSON, docstopic.ApiSpec).Return(apperrors.Internal("some other error"))
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, assetStoreMock)

		err := service.CreateApiResources("appName", types.UID("appUUID"), "serviceID", &credentials, jsonApiSpec, specFormatJSON, docstopic.OpenApiType)

		// then
		require.Error(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		assetStoreMock.AssertExpectations(t)
	})

	t.Run("should update API resources", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		assetStoreMock := assetstoremock.Service{}

		credentials := model.CredentialsWithCSRF{
			Basic: &model.Basic{
				Username: "admin",
				Password: "nimda",
			},
		}

		accessServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		istioServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		secretServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", mock.MatchedBy(getCredentialsMatcher(&credentials))).Return(nil)
		assetStoreMock.On("Put", "serviceID", docstopic.OpenApiType, jsonApiSpec, specFormatJSON, docstopic.ApiSpec).Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, assetStoreMock)

		err := service.UpdateApiResources("appName", types.UID("appUUID"), "serviceID", &credentials, jsonApiSpec, specFormatJSON, docstopic.OpenApiType)

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		assetStoreMock.AssertExpectations(t)
	})

	t.Run("should update Event API resources", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		assetStoreMock := assetstoremock.Service{}

		assetStoreMock.On("Put", "serviceID", docstopic.AsyncApi, eventsSpec, specFormatJSON, docstopic.EventApiSpec).Return(nil)

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, assetStoreMock)

		err := service.UpdateEventApiResources("appName", "serviceID", eventsSpec, specFormatJSON, docstopic.AsyncApi)

		// then
		require.NoError(t, err)
		assetStoreMock.AssertExpectations(t)
	})

	t.Run("should not update secret if credentials not provided", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		assetStoreMock := assetstoremock.Service{}

		accessServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		istioServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		secretServiceMock.On("Delete", "secretName").Return(nil)
		assetStoreMock.On("Put", "serviceID", docstopic.OpenApiType, jsonApiSpec, specFormatJSON, docstopic.ApiSpec).Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")
		nameResolver.On("GetCredentialsSecretName", "appName", "serviceID").Return("secretName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, assetStoreMock)

		err := service.UpdateApiResources("appName", types.UID("appUUID"), "serviceID", nil, jsonApiSpec, specFormatJSON, docstopic.OpenApiType)

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertNotCalled(t, "Upsert")
		istioServiceMock.AssertExpectations(t)
		assetStoreMock.AssertExpectations(t)
	})

	t.Run("should not interrupt execution when error occurs on update with credentials", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		assetStoreMock := assetstoremock.Service{}

		credentials := model.CredentialsWithCSRF{
			Basic: &model.Basic{
				Username: "admin",
				Password: "nimda",
			},
		}
		accessServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(apperrors.Internal("some error"))
		istioServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(apperrors.Internal("just another error"))
		secretServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", &credentials).Return(apperrors.Internal("some other error"))
		assetStoreMock.On("Put", "serviceID", docstopic.OpenApiType, jsonApiSpec, specFormatJSON, docstopic.ApiSpec).Return(apperrors.Internal("some other error"))
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, assetStoreMock)

		err := service.UpdateApiResources("appName", types.UID("appUUID"), "serviceID", &credentials, jsonApiSpec, specFormatJSON, docstopic.OpenApiType)

		// then
		require.Error(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		assetStoreMock.AssertExpectations(t)
	})

	t.Run("should not interrupt execution when error occurs on update without credentials", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		assetStoreMock := assetstoremock.Service{}

		accessServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(apperrors.Internal("some error"))
		istioServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(apperrors.Internal("just another error"))
		secretServiceMock.On("Delete", "secretName").Return(apperrors.Internal("some other error"))
		assetStoreMock.On("Put", "serviceID", docstopic.OpenApiType, jsonApiSpec, specFormatJSON, docstopic.ApiSpec).Return(apperrors.Internal("some other error"))
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")
		nameResolver.On("GetCredentialsSecretName", "appName", "serviceID").Return("secretName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, assetStoreMock)

		err := service.UpdateApiResources("appName", types.UID("appUUID"), "serviceID", nil, jsonApiSpec, specFormatJSON, docstopic.OpenApiType)

		// then
		require.Error(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		assetStoreMock.AssertExpectations(t)
	})

	t.Run("should delete API resources with credentials", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		assetStoreMock := assetstoremock.Service{}

		accessServiceMock.On("Delete", "resourceName").Return(nil)
		istioServiceMock.On("Delete", "resourceName").Return(nil)
		secretServiceMock.On("Delete", "secretName").Return(nil)
		assetStoreMock.On("Delete", "serviceID").Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, assetStoreMock)

		err := service.DeleteApiResources("appName", "serviceID", "secretName")

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		assetStoreMock.AssertExpectations(t)
	})

	t.Run("should delete API resources without credentials", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		assetStoreMock := assetstoremock.Service{}

		accessServiceMock.On("Delete", "resourceName").Return(nil)
		istioServiceMock.On("Delete", "resourceName").Return(nil)
		assetStoreMock.On("Delete", "serviceID").Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, assetStoreMock)

		err := service.DeleteApiResources("appName", "serviceID", "")

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		assetStoreMock.AssertExpectations(t)
	})

	t.Run("should not interrupt execution when error occurs on delete", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		assetStoreMock := assetstoremock.Service{}

		accessServiceMock.On("Delete", "resourceName").Return(apperrors.Internal("some error"))
		istioServiceMock.On("Delete", "resourceName").Return(apperrors.Internal("some error"))
		secretServiceMock.On("Delete", "secretName").Return(apperrors.Internal("some error"))
		assetStoreMock.On("Delete", "serviceID").Return(apperrors.Internal("some error"))
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, assetStoreMock)

		err := service.DeleteApiResources("appName", "serviceID", "secretName")

		// then
		require.Error(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		assetStoreMock.AssertExpectations(t)
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
