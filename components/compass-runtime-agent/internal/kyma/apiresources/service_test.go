package apiresources

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	k8sconstsmocks "kyma-project.io/compass-runtime-agent/internal/k8sconsts/mocks"
	accessservicemock "kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/accessservice/mocks"
	istiomocks "kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/istio/mocks"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/clusterassetgroup"
	rafteremock "kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/mocks"
	secretmock "kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/secrets/mocks"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/secrets/model"
)

func TestService(t *testing.T) {

	jsonApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	eventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")

	specFormatJSON := clusterassetgroup.SpecFormatJSON

	t.Run("should create API resources", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		rafterMock := &rafteremock.Service{}

		credentials := model.CredentialsWithCSRF{
			Basic: &model.Basic{
				Username: "admin",
				Password: "nimda",
			},
		}

		assets := []clusterassetgroup.Asset{
			{
				Name:    "id",
				Type:    clusterassetgroup.OpenApiType,
				Format:  specFormatJSON,
				Content: jsonApiSpec,
			},
		}

		accessServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		istioServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		secretServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", mock.MatchedBy(getCredentialsMatcher(&credentials))).Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")
		rafterMock.On("Put", "serviceID", assets).Return(nil)

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, rafterMock)

		err := service.CreateApiResources("appName", types.UID("appUUID"), "serviceID", &credentials, assets)

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		rafterMock.AssertExpectations(t)
	})

	t.Run("should create Event API resources", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		rafterMock := &rafteremock.Service{}

		assets := []clusterassetgroup.Asset{
			{
				Name:    "id",
				Type:    clusterassetgroup.AsyncApi,
				Format:  specFormatJSON,
				Content: eventsSpec,
			},
		}
		rafterMock.On("Put", "serviceID", assets).Return(nil)

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, rafterMock)

		err := service.CreateEventApiResources("appName", "serviceID", assets)

		// then
		require.NoError(t, err)
		rafterMock.AssertExpectations(t)
	})

	t.Run("should not create secret if credentials not provided", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		rafterMock := &rafteremock.Service{}

		assets := []clusterassetgroup.Asset{
			{
				Name:    "id",
				Type:    clusterassetgroup.OpenApiType,
				Format:  specFormatJSON,
				Content: jsonApiSpec,
			},
		}

		accessServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		istioServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")
		rafterMock.On("Put", "serviceID", assets).Return(nil)

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, rafterMock)

		err := service.CreateApiResources("appName", types.UID("appUUID"), "serviceID", nil, assets)

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertNotCalled(t, "Create")
		istioServiceMock.AssertExpectations(t)
		rafterMock.AssertExpectations(t)
	})

	t.Run("should not interrupt execution when error occurs on creation", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		rafterMock := &rafteremock.Service{}

		credentials := model.CredentialsWithCSRF{
			Basic: &model.Basic{
				Username: "admin",
				Password: "nimda",
			},
		}

		assets := []clusterassetgroup.Asset{
			{
				Name:    "id",
				Type:    clusterassetgroup.OpenApiType,
				Format:  specFormatJSON,
				Content: jsonApiSpec,
			},
		}

		accessServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(apperrors.Internal("some error"))
		istioServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(apperrors.Internal("just another error"))
		secretServiceMock.On("Create", "appName", types.UID("appUUID"), "serviceID", &credentials).Return(apperrors.Internal("some other error"))
		rafterMock.On("Put", "serviceID", assets).Return(apperrors.Internal("some other error"))
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, rafterMock)

		err := service.CreateApiResources("appName", types.UID("appUUID"), "serviceID", &credentials, assets)

		// then
		require.Error(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		rafterMock.AssertExpectations(t)
	})

	t.Run("should update API resources", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		rafterMock := &rafteremock.Service{}

		credentials := model.CredentialsWithCSRF{
			Basic: &model.Basic{
				Username: "admin",
				Password: "nimda",
			},
		}

		assets := []clusterassetgroup.Asset{
			{
				Name:    "id",
				Type:    clusterassetgroup.OpenApiType,
				Format:  specFormatJSON,
				Content: jsonApiSpec,
			},
		}

		accessServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		istioServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		secretServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", mock.MatchedBy(getCredentialsMatcher(&credentials))).Return(nil)
		rafterMock.On("Put", "serviceID", assets).Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, rafterMock)

		err := service.UpdateApiResources("appName", types.UID("appUUID"), "serviceID", &credentials, assets)

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		rafterMock.AssertExpectations(t)
	})

	t.Run("should update Event API resources", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		rafterMock := &rafteremock.Service{}

		assets := []clusterassetgroup.Asset{
			{
				Name:    "id",
				Type:    clusterassetgroup.AsyncApi,
				Format:  specFormatJSON,
				Content: eventsSpec,
			},
		}

		rafterMock.On("Put", "serviceID", assets).Return(nil)

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, rafterMock)

		err := service.UpdateEventApiResources("appName", "serviceID", assets)

		// then
		require.NoError(t, err)
		rafterMock.AssertExpectations(t)
	})

	t.Run("should not update secret if credentials not provided", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		rafterMock := &rafteremock.Service{}

		assets := []clusterassetgroup.Asset{
			{
				Name:    "id",
				Type:    clusterassetgroup.OpenApiType,
				Format:  specFormatJSON,
				Content: jsonApiSpec,
			},
		}
		accessServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		istioServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(nil)
		secretServiceMock.On("Delete", "secretName").Return(nil)
		rafterMock.On("Put", "serviceID", assets).Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")
		nameResolver.On("GetCredentialsSecretName", "appName", "serviceID").Return("secretName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, rafterMock)

		err := service.UpdateApiResources("appName", types.UID("appUUID"), "serviceID", nil, assets)

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertNotCalled(t, "Upsert")
		istioServiceMock.AssertExpectations(t)
		rafterMock.AssertExpectations(t)
	})

	t.Run("should not interrupt execution when error occurs on update with credentials", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		rafterMock := &rafteremock.Service{}

		credentials := model.CredentialsWithCSRF{
			Basic: &model.Basic{
				Username: "admin",
				Password: "nimda",
			},
		}

		assets := []clusterassetgroup.Asset{
			{
				Name:    "id",
				Type:    clusterassetgroup.OpenApiType,
				Format:  specFormatJSON,
				Content: jsonApiSpec,
			},
		}

		accessServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(apperrors.Internal("some error"))
		istioServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(apperrors.Internal("just another error"))
		secretServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", &credentials).Return(apperrors.Internal("some other error"))
		rafterMock.On("Put", "serviceID", assets).Return(apperrors.Internal("some other error"))
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, rafterMock)

		err := service.UpdateApiResources("appName", types.UID("appUUID"), "serviceID", &credentials, assets)

		// then
		require.Error(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		rafterMock.AssertExpectations(t)
	})

	t.Run("should not interrupt execution when error occurs on update without credentials", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		rafterMock := &rafteremock.Service{}

		assets := []clusterassetgroup.Asset{
			{
				Name:    "id",
				Type:    clusterassetgroup.OpenApiType,
				Format:  specFormatJSON,
				Content: jsonApiSpec,
			},
		}

		accessServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(apperrors.Internal("some error"))
		istioServiceMock.On("Upsert", "appName", types.UID("appUUID"), "serviceID", "resourceName").Return(apperrors.Internal("just another error"))
		secretServiceMock.On("Delete", "secretName").Return(apperrors.Internal("some other error"))
		rafterMock.On("Put", "serviceID", assets).Return(apperrors.Internal("some other error"))
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")
		nameResolver.On("GetCredentialsSecretName", "appName", "serviceID").Return("secretName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, rafterMock)

		err := service.UpdateApiResources("appName", types.UID("appUUID"), "serviceID", nil, assets)

		// then
		require.Error(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		rafterMock.AssertExpectations(t)
	})

	t.Run("should delete API resources with credentials", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		rafterMock := &rafteremock.Service{}

		accessServiceMock.On("Delete", "resourceName").Return(nil)
		istioServiceMock.On("Delete", "resourceName").Return(nil)
		secretServiceMock.On("Delete", "secretName").Return(nil)
		rafterMock.On("Delete", "serviceID").Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, rafterMock)

		err := service.DeleteApiResources("appName", "serviceID", "secretName")

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		rafterMock.AssertExpectations(t)
	})

	t.Run("should delete API resources without credentials", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		rafterMock := &rafteremock.Service{}

		accessServiceMock.On("Delete", "resourceName").Return(nil)
		istioServiceMock.On("Delete", "resourceName").Return(nil)
		rafterMock.On("Delete", "serviceID").Return(nil)
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, rafterMock)

		err := service.DeleteApiResources("appName", "serviceID", "")

		// then
		require.NoError(t, err)
		accessServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		rafterMock.AssertExpectations(t)
	})

	t.Run("should not interrupt execution when error occurs on delete", func(t *testing.T) {
		// given
		accessServiceMock := &accessservicemock.AccessServiceManager{}
		secretServiceMock := &secretmock.Service{}
		nameResolver := &k8sconstsmocks.NameResolver{}
		istioServiceMock := &istiomocks.Service{}
		rafterMock := &rafteremock.Service{}

		accessServiceMock.On("Delete", "resourceName").Return(apperrors.Internal("some error"))
		istioServiceMock.On("Delete", "resourceName").Return(apperrors.Internal("some error"))
		secretServiceMock.On("Delete", "secretName").Return(apperrors.Internal("some error"))
		rafterMock.On("Delete", "serviceID").Return(apperrors.Internal("some error"))
		nameResolver.On("GetResourceName", "appName", "serviceID").Return("resourceName")

		// when
		service := NewService(accessServiceMock, secretServiceMock, nameResolver, istioServiceMock, rafterMock)

		err := service.DeleteApiResources("appName", "serviceID", "secretName")

		// then
		require.Error(t, err)
		accessServiceMock.AssertExpectations(t)
		secretServiceMock.AssertExpectations(t)
		istioServiceMock.AssertExpectations(t)
		rafterMock.AssertExpectations(t)
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
