package externalapi

import (
	"errors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates/revocationlist/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRevocationHandler(t *testing.T) {

	urlRevocation := "/v1/applications/certificates/revocation"
	hashedTestCert := "f21139ef2b82d02ee73a56c5c73c053fbafa3480a0b35459cba276b0667c57fc"

	t.Run("should revoke certificate and return http code 201", func(t *testing.T) {
		//given
		testCert := "testCert"

		revocationListRepository := &mocks.RevocationListRepository{}
		revocationListRepository.On("Insert", hashedTestCert).Return(nil)

		handler := NewRevocationHandler(revocationListRepository)

		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, urlRevocation, nil)
		req.Header.Set(CertificateHeader, testCert)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusCreated, rr.Code)
		revocationListRepository.AssertExpectations(t)
	})

	t.Run("should return http code 201 when certificate already revoked", func(t *testing.T) {
		//given
		testCert := "testCert"

		revocationListRepository := &mocks.RevocationListRepository{}
		revocationListRepository.On("Insert", hashedTestCert).Return(nil)

		handler := NewRevocationHandler(revocationListRepository)

		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, urlRevocation, nil)
		req.Header.Set(CertificateHeader, testCert)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusCreated, rr.Code)
		revocationListRepository.AssertExpectations(t)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusCreated, rr.Code)
		revocationListRepository.AssertExpectations(t)
	})

	t.Run("should return http code 403 when certificate not passed", func(t *testing.T) {
		//given
		testCert := ""

		revocationListRepository := &mocks.RevocationListRepository{}

		handler := NewRevocationHandler(revocationListRepository)

		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, urlRevocation, nil)
		req.Header.Set(CertificateHeader, testCert)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusForbidden, rr.Code)
		revocationListRepository.AssertNotCalled(t, "Insert", mock.AnythingOfType("string"))
	})

	t.Run("should return http code 500 when certificate revocation not persisted", func(t *testing.T) {
		//given
		testCert := "testCert"

		revocationListRepository := &mocks.RevocationListRepository{}
		revocationListRepository.On("Insert", hashedTestCert).Return(errors.New("Error"))

		handler := NewRevocationHandler(revocationListRepository)

		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, urlRevocation, nil)
		req.Header.Set(CertificateHeader, testCert)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		revocationListRepository.AssertExpectations(t)
	})
}
