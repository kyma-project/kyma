package externalapi

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"

	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
)

type signatureHandler struct {
	tokenRemover             tokens.Remover
	connectorClientExtractor clientcontext.ConnectorClientExtractor
	certificateService       certificates.Service
	host                     string
}

func NewSignatureHandler(tokenRemover tokens.Remover, certificateService certificates.Service, connectorClientExtractor clientcontext.ConnectorClientExtractor, host string) SignatureHandler {
	return &signatureHandler{
		tokenRemover:             tokenRemover,
		connectorClientExtractor: connectorClientExtractor,
		certificateService:       certificateService,
		host:                     host,
	}
}

func (sh *signatureHandler) SignCSR(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	connectorClientContext, err := sh.connectorClientExtractor(r.Context())
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	signingRequest, err := sh.readCertRequest(r)
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	rawCSR, err := decodeStringFromBase64(signingRequest.CSR)
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	clientCert, err := sh.certificateService.SignCSR(rawCSR, connectorClientContext.GetCommonName())
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	sh.tokenRemover.Delete(token)

	httphelpers.RespondWithBody(w, 201, certResponse{CRT: clientCert})
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

func decodeStringFromBase64(string string) ([]byte, apperrors.AppError) {
	bytes, err := base64.StdEncoding.DecodeString(string)
	if err != nil {
		return nil, apperrors.BadRequest("There was an error while parsing the base64 content. An incorrect value was provided.")
	}

	return bytes, nil
}
