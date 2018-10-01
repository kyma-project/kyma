package certificates

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

const Certificate_Validity_Days = 90

type CertificateUtility interface {
	LoadCert(encodedData []byte) (crt *x509.Certificate, appError apperrors.AppError)
	LoadKey(encodedData []byte) (key *rsa.PrivateKey, appError apperrors.AppError)
	LoadCSR(encodedData string) (csr *x509.CertificateRequest, appError apperrors.AppError)
	CheckCSRValues(csr *x509.CertificateRequest, subject CSRSubject) apperrors.AppError
	SignWithCA(caCrt *x509.Certificate, csr *x509.CertificateRequest, key *rsa.PrivateKey) (crtBase64 string, appError apperrors.AppError)
}

type certificateUtility struct {
}

type CSRSubject struct {
	CName              string
	Country            string
	Organization       string
	OrganizationalUnit string
	Locality           string
	Province           string
}

func NewCertificateUtility() CertificateUtility {
	return &certificateUtility{}
}

func (cu *certificateUtility) decodeBytesFromBase64(bytes []byte) (decodedData []byte, appError apperrors.AppError) {
	data := make([]byte, base64.StdEncoding.DecodedLen(len(bytes)))
	_, err := base64.StdEncoding.Decode(data, bytes)
	if err != nil {
		return nil, apperrors.Internal("There was an error while parsing the base64 content. An incorrect value was provided.")
	}

	return data, nil
}

func (cu *certificateUtility) decodeStringFromBase64(bytes string) (decodedData []byte, appError apperrors.AppError) {
	data, err := base64.StdEncoding.DecodeString(bytes)
	if err != nil {
		return nil, apperrors.BadRequest("There was an error while parsing the base64 content. An incorrect value was provided.")
	}

	return data, nil
}

func (cu *certificateUtility) encodeStringBase64(bytes []byte) (data string) {
	return base64.StdEncoding.EncodeToString(bytes)
}

func (cu *certificateUtility) LoadCert(encodedData []byte) (crt *x509.Certificate, appError apperrors.AppError) {

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

func (cu *certificateUtility) LoadKey(encodedData []byte) (key *rsa.PrivateKey, appError apperrors.AppError) {

	pemBlock, _ := pem.Decode(encodedData)
	if pemBlock == nil {
		return nil, apperrors.Internal("Error while decoding pem block.")
	}

	caPrivateKey, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, apperrors.Internal("Error while parsing private key: %s", err)
	}

	return caPrivateKey.(*rsa.PrivateKey), nil
}

func (cu *certificateUtility) LoadCSR(encodedData string) (csr *x509.CertificateRequest, appError apperrors.AppError) {
	decodedData, appErr := cu.decodeStringFromBase64(encodedData)
	if appErr != nil {
		return nil, appErr
	}

	pemBlock, _ := pem.Decode(decodedData)
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
	if csr.Subject.CommonName != subject.CName {
		return apperrors.WrongInput("CSR: Invalid CName provided.")
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

func (cu *certificateUtility) SignWithCA(caCrt *x509.Certificate, csr *x509.CertificateRequest, key *rsa.PrivateKey) (
	crtBase64 string, appError apperrors.AppError) {

	clientCRTTemplate := x509.Certificate{
		Signature:          csr.Signature,
		SignatureAlgorithm: csr.SignatureAlgorithm,

		PublicKeyAlgorithm: csr.PublicKeyAlgorithm,
		PublicKey:          csr.PublicKey,

		SerialNumber: big.NewInt(2),
		Issuer:       caCrt.Subject,
		Subject:      csr.Subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(Certificate_Validity_Days * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	clientCrtRaw, err := x509.CreateCertificate(rand.Reader, &clientCRTTemplate, caCrt, csr.PublicKey, key)
	if err != nil {
		return "", apperrors.Internal("Error while creating certificate: %s", err)
	}

	clientCrt := &bytes.Buffer{}

	pem.Encode(clientCrt, &pem.Block{Type: "CERTIFICATE", Bytes: clientCrtRaw})

	return cu.encodeStringBase64(clientCrt.Bytes()), nil
}
