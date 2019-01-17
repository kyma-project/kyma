package strategy

import (
	"crypto/x509/pkix"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/certificates"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
)

const (
	CertificateGenPrivateKeyKey = "key"
	CertificateGenCertKey       = "crt"
	CertificateGenCNKey         = "commonName"
)

type certificateGen struct {
	certificateGenerator certificates.Generator
}

func (svc *certificateGen) ToCredentials(secretData map[string][]byte, appCredentials *applications.Credentials) model.Credentials {
	commonName := svc.readCertificateGenMap(secretData)

	return model.Credentials{
		CertificateGen: &model.CertificateGen{
			CommonName: commonName,
		},
	}
}

func (svc *certificateGen) CredentialsProvided(credentials *model.Credentials) bool {
	return svc.certificateGenCredentialsProvided(credentials)
}

func (svc *certificateGen) CreateSecretData(credentials *model.Credentials) (map[string][]byte, apperrors.AppError) {
	keyCertPair, err := svc.certificateGenerator(pkix.Name{CommonName: credentials.CertificateGen.CommonName})
	if err != nil {
		return nil, err.Append("Failed to generate key and certificate pair")
	}

	return svc.makeCertificateGenMap(credentials.CertificateGen.CommonName, keyCertPair.PrivateKey, keyCertPair.Certificate), nil
}

func (svc *certificateGen) ToAppCredentials(credentials *model.Credentials, secretName string) applications.Credentials {
	applicationCredentials := applications.Credentials{
		Type:       applications.CredentialsCertificateGenType,
		SecretName: secretName,
	}

	return applicationCredentials
}

func (svc *certificateGen) certificateGenCredentialsProvided(credentials *model.Credentials) bool {
	return credentials != nil && credentials.CertificateGen != nil && credentials.CertificateGen.CommonName != ""
}

func (svc *certificateGen) makeCertificateGenMap(commonName string, key, certificate []byte) map[string][]byte {
	return map[string][]byte{
		CertificateGenCNKey:         []byte(commonName),
		CertificateGenPrivateKeyKey: key,
		CertificateGenCertKey:       certificate,
	}
}

func (svc *certificateGen) readCertificateGenMap(data map[string][]byte) (commonName string) {
	return string(data[CertificateGenCNKey])
}
