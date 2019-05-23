package certificates

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"net/http"
	"regexp"
	"strings"
)

const clientCertHeader = "X-Forwarded-Client-Cert"

type CertInfo struct {
	Hash    string
	Subject string
	URI     string
}

type ValidationInfo struct {
	Organization string
	Unit         string
	Central      bool
}

func ParseCertificateHeader(r http.Request, vi ValidationInfo) (CertInfo, apperrors.AppError) {
	certHeader := r.Header.Get(clientCertHeader)

	data := strings.Split(certHeader, ";")

	infoParts := groupData(data)

	certInfos := createCertInfos(infoParts)

	if vi.Central {
		return getCertInfoWithNonEmptySubject(certInfos)
	} else {
		return getCertInfoWithMatchingSubject(certInfos, vi.Organization, vi.Unit)
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

func GetOrganization(subject string) string {
	return getRegexMatch("O=([^,]+)", subject)
}

func GetOrganizationalUnit(subject string) string {
	return getRegexMatch("OU=([^,]+)", subject)
}

func GetCommonName(subject string) string {
	return getRegexMatch("CN=([^,]+)", subject)
}

func getRegexMatch(regex, text string) string {
	cnRegex := regexp.MustCompile(regex)
	matches := cnRegex.FindStringSubmatch(text)

	if len(matches) != 2 {
		return ""
	}

	return matches[1]
}
