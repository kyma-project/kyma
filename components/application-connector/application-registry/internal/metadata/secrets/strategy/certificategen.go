package strategy

import (
	"crypto/x509/pkix"
	"encoding/base64"

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

func (svc *certificateGen) ToCredentials(secretData SecretData, appCredentials *applications.Credentials) (model.CredentialsWithCSRF, apperrors.AppError) {
	commonName, cert := svc.readCertificateGenMap(secretData)

	return model.CredentialsWithCSRF{
		CertificateGen: &model.CertificateGen{
			CommonName:  commonName,
			Certificate: cert,
		},
		CSRFInfo: convertToModelCSRInfo(appCredentials),
	}, nil
}

func (svc *certificateGen) CredentialsProvided(credentials *model.CredentialsWithCSRF) bool {
	return svc.certificateGenCredentialsProvided(credentials)
}

func (svc *certificateGen) CreateSecretData(credentials *model.CredentialsWithCSRF) (SecretData, apperrors.AppError) {
	keyCertPair, err := svc.certificateGenerator(pkix.Name{CommonName: credentials.CertificateGen.CommonName})
	if err != nil {
		return nil, err.Append("Failed to generate key and certificate pair")
	}

	return svc.makeCertificateGenMap(credentials.CertificateGen.CommonName, keyCertPair.PrivateKey, keyCertPair.Certificate), nil
}

func (svc *certificateGen) ToCredentialsInfo(credentials *model.CredentialsWithCSRF, secretName string) applications.Credentials {
	applicationCredentials := applications.Credentials{
		Type:       applications.CredentialsCertificateGenType,
		SecretName: secretName,
		CSRFInfo:   toAppCSRFInfo(credentials),
	}

	return applicationCredentials
}

func (svc *certificateGen) ShouldUpdate(currentData SecretData, newData SecretData) bool {
	return string(currentData[CertificateGenCNKey]) != string(newData[CertificateGenCNKey])
}

func (svc *certificateGen) certificateGenCredentialsProvided(credentials *model.CredentialsWithCSRF) bool {
	return credentials != nil && credentials.CertificateGen != nil && credentials.CertificateGen.CommonName != ""
}

func (svc *certificateGen) makeCertificateGenMap(commonName string, key, certificate []byte) map[string][]byte {
	return map[string][]byte{
		CertificateGenCNKey:         []byte(commonName),
		CertificateGenPrivateKeyKey: key,
		CertificateGenCertKey:       certificate,
	}
}

func (svc *certificateGen) readCertificateGenMap(data map[string][]byte) (commonName, certificate string) {
	return string(data[CertificateGenCNKey]), encodeCertificateToString(data[CertificateGenCertKey])
}

func encodeCertificateToString(cert []byte) string {
	return base64.StdEncoding.EncodeToString(cert)
}
