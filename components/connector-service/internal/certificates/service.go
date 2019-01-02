package certificates

import (
	"crypto/x509"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/secrets"
)

type Service interface {
	SignCSR(encodedCSR string, identifier string) (string, apperrors.AppError)
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

func (svc *certificateService) SignCSR(encodedCSR string, identifier string) (string, apperrors.AppError) {

	csr, err := svc.certUtil.LoadCSR(encodedCSR)
	if err != nil {
		return "", err
	}

	err = svc.checkCSR(csr, identifier)
	if err != nil {
		return "", err
	}

	caCrtBytesEncoded, caKeyBytesEncoded, err := svc.secretsRepository.Get(svc.caSecretName)
	if err != nil {
		return "", err
	}

	caCrt, err := svc.certUtil.LoadCert(caCrtBytesEncoded)
	if err != nil {
		return "", err
	}

	caKey, err := svc.certUtil.LoadKey(caKeyBytesEncoded)
	if err != nil {
		return "", err
	}

	signedCrt, err := svc.certUtil.CreateCrtChain(caCrt, csr, caKey)
	if err != nil {
		return "", err
	}

	return signedCrt, nil
}

func (svc *certificateService) checkCSR(csr *x509.CertificateRequest, appName string) apperrors.AppError {
	subjectValues := CSRSubject{
		CName:              appName,
		Country:            svc.csrSubject.Country,
		Organization:       svc.csrSubject.Organization,
		OrganizationalUnit: svc.csrSubject.OrganizationalUnit,
		Locality:           svc.csrSubject.Locality,
		Province:           svc.csrSubject.Province,
	}

	return svc.certUtil.CheckCSRValues(csr, subjectValues)
}
