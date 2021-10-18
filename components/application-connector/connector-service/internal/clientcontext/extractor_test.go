package clientcontext

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/certificates"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/apperrors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	appName = "appName"
	tenant  = "tenant"
	group   = "group"
)

var (
	subjectDefaults = certificates.CSRSubject{
		Country:            "PL",
		Organization:       "Org",
		OrganizationalUnit: "OrgUnit",
		Province:           "Province",
		Locality:           "Gliwice",
		CommonName:         "CommonName",
	}
)

func Test_ExtractSerializableApplicationContext(t *testing.T) {

	t.Run("should return ApplicationContext", func(t *testing.T) {
		// given
		expectedSubject := certificates.CSRSubject{
			Country:            "PL",
			Organization:       tenant,
			OrganizationalUnit: group,
			Province:           "Province",
			Locality:           "Gliwice",
			CommonName:         appName,
		}

		appCtxPayload := ApplicationContext{
			Application:    appName,
			ClusterContext: ClusterContext{Group: group, Tenant: tenant},
		}

		ctx := appCtxPayload.ExtendContext(context.Background())

		extractor := NewContextExtractor(subjectDefaults)

		// when
		clientCtx, err := extractor.CreateApplicationClientContextService(ctx)
		require.NoError(t, err)

		// then
		certContext, ok := clientCtx.(*clientCertificateContext)
		require.True(t, ok)

		assert.Equal(t, expectedSubject, certContext.GetSubject())

		extractedAppCtx, ok := certContext.ClientContextService.(ApplicationContext)
		require.True(t, ok)

		assert.Equal(t, appCtxPayload, extractedAppCtx)
	})

	t.Run("should return ApplicationContext with Runtime URLs", func(t *testing.T) {
		// given
		expectedSubject := certificates.CSRSubject{
			Country:            "PL",
			Organization:       tenant,
			OrganizationalUnit: group,
			Province:           "Province",
			Locality:           "Gliwice",
			CommonName:         appName,
		}

		eventsBasedURL := "https://gateway.cool-cluster.cluster.extend.events.cx"
		metadataBasedURL := "https://gateway.cool-cluster.cluster.extend.metadata.cx"

		appCtxPayload := ApplicationContext{
			Application:    appName,
			ClusterContext: ClusterContext{Group: group, Tenant: tenant},
		}

		ctx := appCtxPayload.ExtendContext(context.Background())

		apiUrls := ApiURLs{
			EventsBaseURL:   eventsBasedURL,
			MetadataBaseURL: metadataBasedURL,
		}

		ctx = apiUrls.ExtendContext(ctx)

		expectedApplicationContext := ExtendedApplicationContext{
			ApplicationContext: appCtxPayload,
			RuntimeURLs: RuntimeURLs{
				MetadataURL:   metadataBasedURL + "/" + appName + "/v1/metadata/services",
				EventsURL:     eventsBasedURL + "/" + appName + "/v1/events",
				EventsInfoURL: eventsBasedURL + "/" + appName + "/v1/events/subscribed",
			},
		}

		extractor := NewContextExtractor(subjectDefaults)

		// when
		clientCtx, err := extractor.CreateApplicationClientContextService(ctx)
		require.NoError(t, err)

		// then
		certContext, ok := clientCtx.(*clientCertificateContext)
		require.True(t, ok)

		assert.Equal(t, expectedSubject, certContext.GetSubject())

		extractedAppCtx, ok := certContext.ClientContextService.(ExtendedApplicationContext)
		require.True(t, ok)

		assert.Equal(t, expectedApplicationContext, extractedAppCtx)
	})

	t.Run("should fail when there is no ApplicationContext", func(t *testing.T) {
		// given
		extractor := NewContextExtractor(subjectDefaults)

		// when
		_, err := extractor.CreateApplicationClientContextService(context.Background())
		require.Error(t, err)

		// then
		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})
}

func Test_ExtractSerializableClusterContext(t *testing.T) {
	t.Run("should return ClusterToken", func(t *testing.T) {
		// given
		expectedSubject := certificates.CSRSubject{
			Country:            "PL",
			Organization:       tenant,
			OrganizationalUnit: group,
			Province:           "Province",
			Locality:           "Gliwice",
			CommonName:         RuntimeDefaultCommonName,
		}

		clusterCtxPayload := ClusterContext{Group: group, Tenant: tenant}

		ctx := clusterCtxPayload.ExtendContext(context.Background())

		extractor := NewContextExtractor(subjectDefaults)

		// when
		clientCtx, err := extractor.CreateClusterClientContextService(ctx)
		require.NoError(t, err)

		// then
		certContext, ok := clientCtx.(*clientCertificateContext)
		require.True(t, ok)

		assert.Equal(t, expectedSubject, certContext.GetSubject())

		extractedClusterCtx, ok := certContext.ClientContextService.(ClusterContext)
		require.True(t, ok)

		assert.Equal(t, clusterCtxPayload, extractedClusterCtx)
	})

	t.Run("should fail when there is no ClusterContext", func(t *testing.T) {
		// given
		extractor := NewContextExtractor(subjectDefaults)

		// when
		_, err := extractor.CreateClusterClientContextService(context.Background())
		require.Error(t, err)

		// then
		assert.Equal(t, apperrors.CodeInternal, err.Code())
	})
}
