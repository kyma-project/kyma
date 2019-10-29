package externalapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	certmocks "github.com/kyma-project/kyma/components/connector-service/internal/certificates/mocks"

	"github.com/kyma-project/kyma/components/connector-service/internal/revocation/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRevocationHandler(t *testing.T) {

	urlRevocation := "/v1/applications/certificates/revocations"
	hash := "f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad"
	certInfo := certificates.CertInfo{Hash: hash, Subject: ""}

	testCertHeader := "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"CN=test-application,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE\";" +
		"URI=spiffe://cluster.local/ns/kyma-integration/sa/default;Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account"

	headerParser := &certmocks.HeaderParser{}

	t.Run("should revoke certificate and return http code 201", func(t *testing.T) {
		//given
		revocationListRepository := &mocks.RevocationListRepository{}
		revocationListRepository.On("Insert", hash).Return(nil)

		handler := NewRevocationHandler(revocationListRepository, headerParser)

		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, urlRevocation, nil)
		req.Header.Set(certificates.ClientCertHeader, testCertHeader)
		headerParser.On("ParseCertificateHeader", *req).Return(certInfo, nil)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusCreated, rr.Code)
		revocationListRepository.AssertExpectations(t)
	})

	t.Run("should return http code 201 when certificate already revoked", func(t *testing.T) {
		//given
		revocationListRepository := &mocks.RevocationListRepository{}
		revocationListRepository.On("Insert", hash).Return(nil)

		handler := NewRevocationHandler(revocationListRepository, headerParser)

		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, urlRevocation, nil)
		req.Header.Set(certificates.ClientCertHeader, testCertHeader)
		headerParser.On("ParseCertificateHeader", *req).Return(certInfo, nil)

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

	t.Run("should return http code 400 when certificate not passed", func(t *testing.T) {
		//given
		testCert := ""

		revocationListRepository := &mocks.RevocationListRepository{}

		handler := NewRevocationHandler(revocationListRepository, headerParser)

		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, urlRevocation, nil)
		req.Header.Set(certificates.ClientCertHeader, testCert)

		headerParser.On("ParseCertificateHeader", *req).Return(certificates.CertInfo{}, apperrors.BadRequest("Cert header is empty"))

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		revocationListRepository.AssertNotCalled(t, "Insert", mock.AnythingOfType("string"))
	})

	t.Run("should return http code 500 when certificate revocation not persisted", func(t *testing.T) {
		//given
		revocationListRepository := &mocks.RevocationListRepository{}
		revocationListRepository.On("Insert", hash).Return(errors.New("Error"))

		handler := NewRevocationHandler(revocationListRepository, headerParser)

		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, urlRevocation, nil)
		req.Header.Set(certificates.ClientCertHeader, testCertHeader)
		headerParser.On("ParseCertificateHeader", *req).Return(certInfo, nil)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		revocationListRepository.AssertExpectations(t)
	})
}
