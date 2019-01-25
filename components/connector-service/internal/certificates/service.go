package certificates

import (
	"crypto/x509"
	"encoding/base64"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/secrets"
)

type Service interface {
	// SignCSR takes encoded CSR, validates subject and generates Certificate based on CA stored in secret
	// returns encoded certificate chain
	SignCSR(encodedCSR []byte, commonName string) (EncodedCertificateChain, apperrors.AppError)
}

type certificateService struct {
	secretsRepository secrets.Repository
	certUtil          CertificateUtility
	caSecretName      string
	csrSubject        CSRSubject
}

func NewCertificateService(secretRepository secrets.Repository, certUtil CertificateUtility, caSecretName string, csrSubject CSRSubject) Service {
	return &certificateService{
		secretsRepository: secretRepository,
		certUtil:          certUtil,
		caSecretName:      caSecretName,
		csrSubject:        csrSubject,
	}
}

func (svc *certificateService) SignCSR(encodedCSR []byte, commonName string) (EncodedCertificateChain, apperrors.AppError) {
	csr, err := svc.certUtil.LoadCSR(encodedCSR)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	err = svc.checkCSR(csr, commonName)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	return svc.signCSR(csr)
}

func (svc *certificateService) signCSR(csr *x509.CertificateRequest) (EncodedCertificateChain, apperrors.AppError) {
	caCrtBytesEncoded, caKeyBytesEncoded, err := svc.secretsRepository.Get(svc.caSecretName)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	caCrt, err := svc.certUtil.LoadCert(caCrtBytesEncoded)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	caKey, err := svc.certUtil.LoadKey(caKeyBytesEncoded)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	signedCrt, err := svc.certUtil.SignCSR(caCrt, csr, caKey)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	certChain := svc.certUtil.CreateCrtChain(caCrt.Raw, signedCrt)

	return encodeCertificateChain(certChain, signedCrt, caCrt.Raw), nil
}

func (svc *certificateService) checkCSR(csr *x509.CertificateRequest, commonName string) apperrors.AppError {

	subjectValues := CSRSubject{
		CommonName:         commonName,
		Country:            svc.csrSubject.Country,
		Organization:       svc.csrSubject.Organization,
		OrganizationalUnit: svc.csrSubject.OrganizationalUnit,
		Locality:           svc.csrSubject.Locality,
		Province:           svc.csrSubject.Province,
	}

	return svc.certUtil.CheckCSRValues(csr, subjectValues)
}

func encodeCertificateChain(certChain, clientCRT, caCRT []byte) EncodedCertificateChain {
	return EncodedCertificateChain{
		CertificateChain:  encodeStringBase64(certChain),
		ClientCertificate: encodeStringBase64(clientCRT),
		CaCertificate:     encodeStringBase64(caCRT),
	}
}

func encodeStringBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}
