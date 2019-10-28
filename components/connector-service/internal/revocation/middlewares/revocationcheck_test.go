package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	certmock "github.com/kyma-project/kyma/components/connector-service/internal/certificates/mocks"

	"github.com/kyma-project/kyma/components/connector-service/internal/revocation/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRevocationCheckMiddleware(t *testing.T) {

	hash := "f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad"

	certInfo := certificates.CertInfo{Hash: hash, Subject: ""}

	testCertHeader := "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"CN=test-application,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE\";" +
		"URI=spiffe://cluster.local/ns/kyma-integration/sa/default;Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"

	headerParser := &certmock.HeaderParser{}

	t.Run("should return http code 403 when certificate fingerprint present on revocation list", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)
		req.Header.Set(certificates.ClientCertHeader, testCertHeader)

		rr := httptest.NewRecorder()

		repository := &mocks.RevocationListRepository{}
		repository.On("Contains", hash).Return(true, nil)

		headerParser.On("ParseCertificateHeader", *req).Return(certInfo, nil)

		middleware := NewRevocationCheckMiddleware(repository, headerParser)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("should allow certificate when hash not on revocation list", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)
		req.Header.Set(certificates.ClientCertHeader, testCertHeader)

		rr := httptest.NewRecorder()

		repository := &mocks.RevocationListRepository{}
		repository.On("Contains", hash).Return(false, nil)
		headerParser.On("ParseCertificateHeader", *req).Return(certInfo, nil)

		middleware := NewRevocationCheckMiddleware(repository, headerParser)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should return http code 500 when error occurred on contains method call", func(t *testing.T) {
		// given
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)
		req.Header.Set(certificates.ClientCertHeader, testCertHeader)

		rr := httptest.NewRecorder()

		repository := &mocks.RevocationListRepository{}
		repository.On("Contains", hash).Return(false, errors.New("Some error"))
		headerParser.On("ParseCertificateHeader", *req).Return(certInfo, nil)

		middleware := NewRevocationCheckMiddleware(repository, headerParser)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("should return http code 400 when certificate contains corrupted data", func(t *testing.T) {
		// given
		testCertHeader := "Hash=;URI=spiffe://cluster.local/ns/kyma-integration/sa/default;Subject=\"\""
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("POST", "/", nil)
		require.NoError(t, err)
		req.Header.Set(certificates.ClientCertHeader, testCertHeader)

		headerParser.On("ParseCertificateHeader", *req).Return(certificates.CertInfo{}, apperrors.BadRequest("Something wrong with cert"))

		rr := httptest.NewRecorder()

		repository := &mocks.RevocationListRepository{}

		middleware := NewRevocationCheckMiddleware(repository, headerParser)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
