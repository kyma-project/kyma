package externalapi

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/clientcontext"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/certificates"
)

type signatureHandler struct {
	connectorClientExtractor clientcontext.ConnectorClientExtractor
	certificateService       certificates.Service
}

func NewSignatureHandler(certificateService certificates.Service, connectorClientExtractor clientcontext.ConnectorClientExtractor) SignatureHandler {
	return &signatureHandler{
		connectorClientExtractor: connectorClientExtractor,
		certificateService:       certificateService,
	}
}

func (sh *signatureHandler) SignCSR(w http.ResponseWriter, r *http.Request) {
	clientContextService, err := sh.connectorClientExtractor(r.Context())
	if err != nil {
		httphelpers.RespondWithErrorAndLog(w, err)
		return
	}

	signingRequest, err := readCertRequest(r)
	if err != nil {
		httphelpers.RespondWithErrorAndLog(w, err)
		return
	}

	rawCSR, err := decodeStringFromBase64(signingRequest.CSR)
	if err != nil {
		httphelpers.RespondWithErrorAndLog(w, err)
		return
	}

	encodedCertificatesChain, err := sh.certificateService.SignCSR(rawCSR, clientContextService.GetSubject())
	if err != nil {
		httphelpers.RespondWithErrorAndLog(w, err)
		return
	}

	httphelpers.RespondWithBody(w, http.StatusCreated, toCertResponse(encodedCertificatesChain))
}

func readCertRequest(r *http.Request) (*certRequest, apperrors.AppError) {
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

func decodeStringFromBase64(string string) ([]byte, apperrors.AppError) {
	bytes, err := base64.StdEncoding.DecodeString(string)
	if err != nil {
		return nil, apperrors.BadRequest("There was an error while parsing the base64 content. An incorrect value was provided.")
	}

	return bytes, nil
}
