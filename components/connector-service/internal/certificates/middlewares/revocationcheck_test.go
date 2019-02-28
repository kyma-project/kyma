package middlewares

import (
	"errors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates/revocationlist/mocks"
	"github.com/kyma-project/kyma/components/connector-service/internal/externalapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRevocationCheckMiddleware(t *testing.T) {

	t.Run("should return http code 403 when certificate hash on revocation list", func(t *testing.T) {
		// given
		testCert := "testCert"
		hashedTestCert := "f21139ef2b82d02ee73a56c5c73c053fbafa3480a0b35459cba276b0667c57fc"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("POST", "/", nil)
		req.Header.Set(externalapi.CertificateHeader, testCert)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		repository := &mocks.RevocationListRepository{}
		repository.On("Contains", hashedTestCert).Return(true, nil)

		middleware := NewRevocationCheckMiddleware(repository)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("should allow certificate when hash not on revocation list", func(t *testing.T) {
		// given
		testCert := "testCert"
		hashedTestCert := "f21139ef2b82d02ee73a56c5c73c053fbafa3480a0b35459cba276b0667c57fc"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("POST", "/", nil)
		req.Header.Set(externalapi.CertificateHeader, testCert)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		repository := &mocks.RevocationListRepository{}
		repository.On("Contains", hashedTestCert).Return(false, nil)

		middleware := NewRevocationCheckMiddleware(repository)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("should return http code 500 when error occured on contains method call", func(t *testing.T) {
		// given
		testCert := "testCert"
		hashedTestCert := "f21139ef2b82d02ee73a56c5c73c053fbafa3480a0b35459cba276b0667c57fc"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("POST", "/", nil)
		req.Header.Set(externalapi.CertificateHeader, testCert)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		repository := &mocks.RevocationListRepository{}
		repository.On("Contains", hashedTestCert).Return(false, errors.New("Some error"))

		middleware := NewRevocationCheckMiddleware(repository)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("should return http code 500 when error occured during hash calculation", func(t *testing.T) {
		// given
		testCert := "testCert%WrongEscape%"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req, err := http.NewRequest("POST", "/", nil)
		req.Header.Set(externalapi.CertificateHeader, testCert)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		repository := &mocks.RevocationListRepository{}

		middleware := NewRevocationCheckMiddleware(repository)

		// when
		resultHandler := middleware.Middleware(handler)
		resultHandler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
