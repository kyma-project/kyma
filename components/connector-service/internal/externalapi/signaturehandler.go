package externalapi

import (
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/secrets"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache"
)

type signatureHandler struct {
	tokenCache        tokencache.TokenCache
	certUtil          certificates.CertificateUtility
	secretsRepository secrets.Repository
	host              string
	domainName        string
	csr               csrInfo
	tokenParamsParser tokens.TokenParamsParser
}

func NewSignatureHandler(tokenCache tokencache.TokenCache, certUtil certificates.CertificateUtility, secretsRepository secrets.Repository,
	host string, domainName string, subjectValues certificates.CSRSubject) SignatureHandler {
	csr := csrInfo{
		Country:            subjectValues.Country,
		Organization:       subjectValues.Organization,
		OrganizationalUnit: subjectValues.OrganizationalUnit,
		Locality:           subjectValues.Locality,
		Province:           subjectValues.Province,
	}

	return &signatureHandler{
		tokenCache:        tokenCache,
		certUtil:          certUtil,
		secretsRepository: secretsRepository,
		host:              host, domainName: domainName,
		csr: csr,
	}
}

func (sh *signatureHandler) SignCSR(w http.ResponseWriter, r *http.Request) {
	tokenParams, err := sh.tokenParamsParser(r.Context())
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	tokenRequest, appErr := sh.readCertRequest(r)
	if appErr != nil {
		httphelpers.RespondWithError(w, appErr)
		return
	}

	// --------------------------

	// get data for CN

	// --------------------------
	csr, appErr := sh.loadAndCheckCSR(tokenRequest.CSR, reName)
	if appErr != nil {
		httphelpers.RespondWithError(w, appErr)
		return
	}

	signedCrt, appErr := sh.signCSR("nginx-auth-ca", csr)
	if appErr != nil {
		httphelpers.RespondWithError(w, appErr)
		return
	}

	sh.tokenCache.Delete(reName)

	httphelpers.RespondWithBody(w, 201, certResponse{CRT: signedCrt})
}

func (sh *signatureHandler) signCSR(secretName string, csr *x509.CertificateRequest) (
	string, apperrors.AppError) {

	caCrtBytesEncoded, caKeyBytesEncoded, appErr := sh.secretsRepository.Get(secretName)
	if appErr != nil {
		return "", appErr
	}

	caCrt, appErr := sh.certUtil.LoadCert(caCrtBytesEncoded)
	if appErr != nil {
		return "", appErr
	}

	caKey, appErr := sh.certUtil.LoadKey(caKeyBytesEncoded)
	if appErr != nil {
		return "", appErr
	}

	signedCrt, appErr := sh.certUtil.CreateCrtChain(caCrt, csr, caKey)
	if appErr != nil {
		return "", appErr
	}

	return signedCrt, nil
}

func (sh *signatureHandler) readCertRequest(r *http.Request) (*certRequest, apperrors.AppError) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, apperrors.Internal("Error while reading request body: %s", err)
	}
	defer r.Body.Close()

	var tokenRequest certRequest
	err = json.Unmarshal(b, &tokenRequest)
	if err != nil {
		return nil, apperrors.Internal("Error while unmarshalling request body: %s", err)
	}

	return &tokenRequest, nil
}

func (sh *signatureHandler) loadAndCheckCSR(encodedData string, reName string) (*x509.CertificateRequest, apperrors.AppError) {
	csr, appErr := sh.certUtil.LoadCSR(encodedData)
	if appErr != nil {
		return nil, appErr
	}

	subjectValues := certificates.CSRSubject{
		CName:              reName,
		Country:            sh.csr.Country,
		Organization:       sh.csr.Organization,
		OrganizationalUnit: sh.csr.OrganizationalUnit,
		Locality:           sh.csr.Locality,
		Province:           sh.csr.Province,
	}

	appErr = sh.certUtil.CheckCSRValues(csr, subjectValues)
	if appErr != nil {
		return nil, appErr
	}

	return csr, nil
}
