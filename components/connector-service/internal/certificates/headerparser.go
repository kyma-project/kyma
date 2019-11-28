package certificates

import (
	"net/http"
	"regexp"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

const ClientCertHeader = "X-Forwarded-Client-Cert"

type HeaderParser interface {
	ParseCertificateHeader(r http.Request) (CertInfo, apperrors.AppError)
}

type SubjectVerification func(i CertInfo) bool

type headerParser struct {
	country             string
	locality            string
	organization        string
	organizationalUnit  string
	province            string
	central             bool
	subjectVerification SubjectVerification
}

type CertInfo struct {
	Hash    string
	Subject string
}

func NewHeaderParser(country, province, locality, organization, unit string, central bool) HeaderParser {

	headerParser := headerParser{
		country:            country,
		locality:           locality,
		organization:       organization,
		organizationalUnit: unit,
		province:           province,
		central:            central,
	}

	if central {
		headerParser.subjectVerification = headerParser.isSubjectMatchingCentral
	} else {
		headerParser.subjectVerification = headerParser.isSubjectMatching
	}

	return &headerParser
}

func (hp *headerParser) ParseCertificateHeader(r http.Request) (CertInfo, apperrors.AppError) {
	certHeader := r.Header.Get(ClientCertHeader)

	if certHeader == "" {
		return CertInfo{}, apperrors.BadRequest("Certificate header is empty")
	}

	subjectRegex := regexp.MustCompile(`Subject="(.*?)"`)

	subjects := extractFromHeader(certHeader, subjectRegex)

	hashRegex := regexp.MustCompile(`Hash=([0-9a-f]*)`)

	hashes := extractFromHeader(certHeader, hashRegex)

	certInfos := createCertInfos(subjects, hashes)

	return hp.getCertInfoWithMatchingSubject(certInfos)
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
	certInfos := make([]CertInfo, len(subjects))
	for i := 0; i < len(subjects); i++ {
		certInfo := newCertInfo(subjects[i], hashes[i])
		certInfos[i] = certInfo
	}
	return certInfos
}

func (hp *headerParser) getCertInfoWithMatchingSubject(infos []CertInfo) (CertInfo, apperrors.AppError) {
	for _, i := range infos {
		if hp.subjectVerification(i) {
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

func (hp *headerParser) isSubjectMatching(i CertInfo) bool {
	return GetOrganization(i.Subject) == hp.organization && GetOrganizationalUnit(i.Subject) == hp.organizationalUnit &&
		GetCountry(i.Subject) == hp.country && GetLocality(i.Subject) == hp.locality && GetProvince(i.Subject) == hp.province
}

func (hp *headerParser) isSubjectMatchingCentral(i CertInfo) bool {
	return GetCountry(i.Subject) == hp.country && GetLocality(i.Subject) == hp.locality && GetProvince(i.Subject) == hp.province
}
