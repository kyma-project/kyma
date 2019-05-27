package certificates

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"net/http"
	"regexp"
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
}

func (hp *HeaderParser) ParseCertificateHeader(r http.Request) (CertInfo, apperrors.AppError) {
	certHeader := r.Header.Get(ClientCertHeader)

	if certHeader == "" {
		return CertInfo{}, apperrors.BadRequest("Certificate header is empty")
	}

	subjectRegex := regexp.MustCompile(`Subject="(.*?)"`)

	subjects := extractFromHeader(certHeader, subjectRegex)

	hashRegex := regexp.MustCompile(`Hash=([0-9a-f]*)`)

	hashes := extractFromHeader(certHeader, hashRegex)

	certInfos := createCertInfos(subjects, hashes)

	if hp.Central {
		return getCertInfoWithNonEmptySubject(certInfos)
	} else {
		return getCertInfoWithMatchingSubject(certInfos, hp.Organization, hp.Unit)
	}
}

func extractFromHeader(certHeader string, regex *regexp.Regexp) []string {
	var matchedStrings []string

	matches := regex.FindAllStringSubmatch(certHeader, -1)

	for _, match := range matches {
		hash := get(match, 1)
		matchedStrings = append(matchedStrings, hash)
	}

	return matchedStrings
}

func get(array []string, index int) string {
	if len(array) > index {
		return array[index]
	}
	return ""
}

func createCertInfos(subjects, hashes []string) []CertInfo {
	certInfos := make([]CertInfo, 0)
	for i := 0; i < len(subjects); i++ {
		certInfo := newCertInfo(subjects[i], hashes[i])
		certInfos = append(certInfos, certInfo)
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

func newCertInfo(subject, hash string) CertInfo {
	certInfo := CertInfo{
		Hash:    hash,
		Subject: subject,
	}
	return certInfo
}

func isSubjectMatching(i CertInfo, organization string, unit string) bool {
	return GetOrganization(i.Subject) == organization && GetOrganizationalUnit(i.Subject) == unit
}
