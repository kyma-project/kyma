package certificates

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"net/http"
	"strings"
)

const ClientCertHeader = "X-Forwarded-Client-Cert"

type HeaderParser struct {
	Organization string
	Unit         string
	Central      bool
}

type CertInfo struct {
	Hash    string
	Subject string
	URI     string
}

func (hp *HeaderParser) ParseCertificateHeader(r http.Request) (CertInfo, apperrors.AppError) {
	certHeader := r.Header.Get(ClientCertHeader)

	if certHeader == "" {
		return CertInfo{}, apperrors.BadRequest("Certificate header is empty")
	}

	data := strings.Split(certHeader, ";")

	if err := checkForCorrectness(data); err != nil {
		return CertInfo{}, err
	}

	infoParts := groupData(data)

	certInfos := createCertInfos(infoParts)

	if hp.Central {
		return getCertInfoWithNonEmptySubject(certInfos)
	} else {
		return getCertInfoWithMatchingSubject(certInfos, hp.Organization, hp.Unit)
	}
}

func groupData(split []string) [][]string {
	certs := make([][]string, 0)
	for i := 0; i < len(split); i += 3 {
		batch := split[i : i+3]
		certs = append(certs, batch)
	}
	return certs
}

func createCertInfos(infoParts [][]string) []CertInfo {
	certInfos := make([]CertInfo, 0)
	for _, i := range infoParts {
		if isInfoPartValid(i) {
			certInfo := createCertInfo(i)
			certInfos = append(certInfos, certInfo)
		}
	}

	return certInfos
}

func checkForCorrectness(data []string) apperrors.AppError {
	if len(data)%3 != 0 {
		return apperrors.BadRequest("Certificate header is empty")
	}
	return nil
}

func getCertInfoWithNonEmptySubject(infos []CertInfo) (CertInfo, apperrors.AppError) {
	for _, i := range infos {
		if i.Subject != "" {
			return i, nil
		}
	}
	return CertInfo{}, apperrors.BadRequest("Failed to get certificate subject from header.")
}

func getCertInfoWithMatchingSubject(infos []CertInfo, organization, unit string) (CertInfo, apperrors.AppError) {
	for _, i := range infos {
		if isSubjectMatching(i, organization, unit) {
			return i, nil
		}
	}
	return CertInfo{}, apperrors.BadRequest("Failed to get certificate subject from header.")
}

func createCertInfo(i []string) CertInfo {
	certInfo := CertInfo{
		Hash:    strings.Trim(i[0], "Hash="),
		Subject: strings.Trim(strings.Trim(i[1], "Subject="), "\""),
		URI:     strings.Trim(i[2], "URI="),
	}
	return certInfo
}

func isInfoPartValid(i []string) bool {
	return strings.Contains(i[0], "Hash") && strings.Contains(i[1], "Subject") && strings.Contains(i[2], "URI")
}

func isSubjectMatching(i CertInfo, organization string, unit string) bool {
	return GetOrganization(i.Subject) == organization && GetOrganizationalUnit(i.Subject) == unit
}
