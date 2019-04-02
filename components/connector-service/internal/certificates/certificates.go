package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

type CertificateUtility interface {
	LoadCert(encodedData []byte) (*x509.Certificate, apperrors.AppError)
	LoadKey(encodedData []byte) (*rsa.PrivateKey, apperrors.AppError)
	LoadCSR(encodedData []byte) (*x509.CertificateRequest, apperrors.AppError)
	CheckCSRValues(csr *x509.CertificateRequest, subject CSRSubject) apperrors.AppError
	SignCSR(caCrt *x509.Certificate, csr *x509.CertificateRequest, caKey *rsa.PrivateKey) ([]byte, apperrors.AppError)
	AddCertificateHeaderAndFooter(crtRaw []byte) []byte
}

type certificateUtility struct {
	certificateValidityTime time.Duration
}

func NewCertificateUtility(certificateValidityTime time.Duration) CertificateUtility {
	return &certificateUtility{
		certificateValidityTime: certificateValidityTime,
	}
}

func (cu *certificateUtility) LoadCert(encodedData []byte) (*x509.Certificate, apperrors.AppError) {

	pemBlock, _ := pem.Decode(encodedData)
	if pemBlock == nil {
		return nil, apperrors.Internal("Error while decoding pem block")
	}

	caCRT, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, apperrors.Internal("Error while parsing certificate: %s", err)
	}

	return caCRT, nil
}

func (cu *certificateUtility) LoadKey(encodedData []byte) (*rsa.PrivateKey, apperrors.AppError) {

	pemBlock, _ := pem.Decode(encodedData)
	if pemBlock == nil {
		return nil, apperrors.Internal("Error while decoding pem block.")
	}

	if caPrivateKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes); err == nil {
		return caPrivateKey, nil
	}

	caPrivateKey, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, apperrors.Internal("Error while parsing private key: %s", err)
	}

	return caPrivateKey.(*rsa.PrivateKey), nil
}

func (cu *certificateUtility) LoadCSR(encodedData []byte) (*x509.CertificateRequest, apperrors.AppError) {

	pemBlock, _ := pem.Decode(encodedData)
	if pemBlock == nil {
		return nil, apperrors.BadRequest("Error while decoding pem block.")
	}

	clientCSR, err := x509.ParseCertificateRequest(pemBlock.Bytes)
	if err != nil {
		return nil, apperrors.BadRequest("Error while parsing CSR: %s", err)
	}

	err = clientCSR.CheckSignature()
	if err != nil {
		return nil, apperrors.BadRequest("CSR signature invalid: %s", err)
	}

	return clientCSR, nil
}

func (cu *certificateUtility) CheckCSRValues(csr *x509.CertificateRequest, subject CSRSubject) apperrors.AppError {
	if csr.Subject.CommonName != subject.CommonName {
		return apperrors.WrongInput("CSR: Invalid common name provided.")
	}

	if csr.Subject.Country == nil {
		return apperrors.WrongInput("CSR: No country provided.")
	} else if csr.Subject.Country[0] != subject.Country {
		return apperrors.WrongInput("CSR: Invalid country provided.")
	}

	if csr.Subject.Organization == nil {
		return apperrors.WrongInput("CSR: No organization provided.")
	} else if csr.Subject.Organization[0] != subject.Organization {
		return apperrors.WrongInput("CSR: Invalid organization provided.")
	}

	if csr.Subject.OrganizationalUnit == nil {
		return apperrors.WrongInput("CSR: No organizational unit provided.")
	} else if csr.Subject.OrganizationalUnit[0] != subject.OrganizationalUnit {
		return apperrors.WrongInput("CSR: Invalid organizational unit provided.")
	}

	if csr.Subject.Locality == nil {
		return apperrors.WrongInput("CSR: No locality provided.")
	} else if csr.Subject.Locality[0] != subject.Locality {
		return apperrors.WrongInput("CSR: Invalid locality provided.")
	}

	if csr.Subject.Province == nil {
		return apperrors.WrongInput("CSR: No province provided.")
	} else if csr.Subject.Province[0] != subject.Province {
		return apperrors.WrongInput("CSR: Invalid province provided.")
	}
	return nil
}

func (cu *certificateUtility) SignCSR(caCrt *x509.Certificate, csr *x509.CertificateRequest, caKey *rsa.PrivateKey) ([]byte, apperrors.AppError) {
	clientCRTTemplate := cu.prepareCRTTemplate(csr)

	clientCrtRaw, err := x509.CreateCertificate(rand.Reader, &clientCRTTemplate, caCrt, csr.PublicKey, caKey)
	if err != nil {
		return nil, apperrors.Internal("Error while creating certificate: %s", err)
	}

	return clientCrtRaw, nil
}

func (cu *certificateUtility) prepareCRTTemplate(csr *x509.CertificateRequest) x509.Certificate {
	return x509.Certificate{
		SignatureAlgorithm: csr.SignatureAlgorithm,

		SerialNumber: big.NewInt(2),
		Subject:      csr.Subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(cu.certificateValidityTime),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
}

func (cu *certificateUtility) AddCertificateHeaderAndFooter(crtRaw []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: crtRaw})
}
