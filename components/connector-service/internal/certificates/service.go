package certificates

import (
	"context"
	"crypto/x509"
	"encoding/base64"

	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/secrets"
)

const (
	caCertificateSecretKey     = "ca.crt"
	caKeySecretKey             = "ca.key"
	rootCACertificateSecretKey = "cacert"
)

type Service interface {
	// SignCSR takes encoded CSR, validates subject and generates Certificate based on CA stored in secret
	// returns base64 encoded certificate chain
	SignCSR(encodedCSR []byte, subject CSRSubject) (EncodedCertificateChain, apperrors.AppError)
}

type certificateService struct {
	ctx                         context.Context
	secretsRepository           secrets.Repository
	certUtil                    CertificateUtility
	caSecretName                types.NamespacedName
	rootCACertificateSecretName types.NamespacedName
}

func NewCertificateService(secretRepository secrets.Repository, certUtil CertificateUtility, caSecretName, rootCACertificateSecretName types.NamespacedName) Service {
	return &certificateService{
		ctx:                         context.Background(),
		secretsRepository:           secretRepository,
		certUtil:                    certUtil,
		caSecretName:                caSecretName,
		rootCACertificateSecretName: rootCACertificateSecretName,
	}
}

func (svc *certificateService) SignCSR(encodedCSR []byte, subject CSRSubject) (EncodedCertificateChain, apperrors.AppError) {
	csr, err := svc.certUtil.LoadCSR(encodedCSR)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	err = svc.checkCSR(csr, subject)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	return svc.signCSR(csr)
}

func (svc *certificateService) signCSR(csr *x509.CertificateRequest) (EncodedCertificateChain, apperrors.AppError) {
	secretData, err := svc.secretsRepository.Get(svc.ctx, svc.caSecretName)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	caCrt, err := svc.certUtil.LoadCert(secretData[caCertificateSecretKey])
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	caKey, err := svc.certUtil.LoadKey(secretData[caKeySecretKey])
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	signedCrt, err := svc.certUtil.SignCSR(caCrt, csr, caKey)
	if err != nil {
		return EncodedCertificateChain{}, err
	}

	return svc.encodeCertificates(caCrt.Raw, signedCrt)
}

func (svc *certificateService) encodeCertificates(rawCaCertificate, rawClientCertificate []byte) (EncodedCertificateChain, apperrors.AppError) {
	caCrtBytes := svc.certUtil.AddCertificateHeaderAndFooter(rawCaCertificate)
	signedCrtBytes := svc.certUtil.AddCertificateHeaderAndFooter(rawClientCertificate)

	if svc.rootCACertificateSecretName.Name != "" {
		rootCABytes, err := svc.loadRootCACert()
		if err != nil {
			return EncodedCertificateChain{}, err
		}

		caCrtBytes = svc.createCertChain(rootCABytes, caCrtBytes)
	}

	certChain := svc.createCertChain(signedCrtBytes, caCrtBytes)

	return encodeCertificateBase64(certChain, signedCrtBytes, caCrtBytes), nil
}

func (svc *certificateService) loadRootCACert() ([]byte, apperrors.AppError) {
	secretData, err := svc.secretsRepository.Get(svc.ctx, svc.rootCACertificateSecretName)
	if err != nil {
		return nil, err
	}

	rootCACrt, err := svc.certUtil.LoadCert(secretData[rootCACertificateSecretKey])
	if err != nil {
		return nil, err
	}

	return svc.certUtil.AddCertificateHeaderAndFooter(rootCACrt.Raw), nil
}

func (svc *certificateService) checkCSR(csr *x509.CertificateRequest, expectedSubject CSRSubject) apperrors.AppError {
	return svc.certUtil.CheckCSRValues(csr, expectedSubject)
}

func (svc *certificateService) createCertChain(clientCrt, caCrt []byte) []byte {
	return append(clientCrt, caCrt...)
}

func encodeCertificateBase64(certChain, clientCRT, caCRT []byte) EncodedCertificateChain {
	return EncodedCertificateChain{
		CertificateChain:  encodeStringBase64(certChain),
		ClientCertificate: encodeStringBase64(clientCRT),
		CaCertificate:     encodeStringBase64(caCRT),
	}
}

func encodeStringBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}
