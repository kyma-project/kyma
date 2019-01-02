package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"
)

func MakeCertInfo(info certificates.CSRSubject, reName string) CertificateInfo {
	subject := fmt.Sprintf("OU=%s,O=%s,L=%s,ST=%s,C=%s,CN=%s", info.OrganizationalUnit, info.Organization, info.Locality, info.Province, info.Country, reName)

	return CertificateInfo{
		Subject:      subject,
		Extensions:   "",
		KeyAlgorithm: "rsa2048",
	}
}

// TODO: improve
func RespondWithError(w http.ResponseWriter, apperr apperrors.AppError) {
	log.Error(apperr.Error())
	statusCode, responseBody := httperrors.AppErrorToResponse(apperr)

	respond(w, statusCode)
	json.NewEncoder(w).Encode(responseBody)
}

func RespondWithBody(w http.ResponseWriter, statusCode int, responseBody interface{}) {
	respond(w, statusCode)
	json.NewEncoder(w).Encode(responseBody)
}

func respond(w http.ResponseWriter, statusCode int) {
	w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)
	w.WriteHeader(statusCode)
}

func ReadRequestBody(r *http.Request, outStruct interface{}) apperrors.AppError {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return apperrors.Internal("Error while reading request body: %s", err)
	}
	defer r.Body.Close()

	err = json.Unmarshal(b, outStruct)
	if err != nil {
		return apperrors.Internal("Error while unmarshalling request body: %s", err)
	}

	return nil
}
