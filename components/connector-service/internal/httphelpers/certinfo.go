package httphelpers

import (
	"github.com/pkg/errors"
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

func ParseCertificateHeader(r http.Request, organization, unit string) (CertInfo, error) {
	certHeader := r.Header.Get(clientCertHeader)

	data := strings.Split(certHeader, ";")

	infoParts := groupData(data)

	for _, i := range infoParts {
		if isInfoPartValid(i) {
			certInfo := createCertInfo(i)
			if isSubjectMatching(certInfo, organization, unit) {
				return certInfo, nil
			}
		}
	}

	return CertInfo{}, errors.New("unable to find matching certificate info")
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

func groupData(split []string) [][]string {
	certs := make([][]string, 0)
	for i := 0; i < len(split); i += 3 {
		batch := split[i : i+3]
		certs = append(certs, batch)
	}
	return certs
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

func getRegexMatch(regex, text string) string {
	cnRegex := regexp.MustCompile(regex)
	matches := cnRegex.FindStringSubmatch(text)

	if len(matches) != 2 {
		return ""
	}

	return matches[1]
}
