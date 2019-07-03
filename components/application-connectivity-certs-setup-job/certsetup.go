package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/pkg/errors"
)

const (
	rsaKeySize             = 4096
	caCertificateSecretKey = "cacert"

	connectorCertSecretKey = "ca.crt"
	connectorKeySecretKey  = "ca.key"
)

type certSetupHandler struct {
	secretRepository SecretRepository
	options          *options
}

func NewCertificateSetupHandler(options *options, secretRepo SecretRepository) *certSetupHandler {
	return &certSetupHandler{
		secretRepository: secretRepo,
		options:          options,
	}
}

func (csh *certSetupHandler) SetupApplicationConnectorCertificate() error {
	// TODO - should we validate those provided?
	// TODO - should check if secrets already populated?
	if csh.certAndKeyProvided() {
		return csh.populateSecrets([]byte(csh.options.caKey), []byte(csh.options.caCertificate))
	}

	key, certificate, err := csh.generateKeyAndCertificate()
	if err != nil {
		return errors.Wrap(err, "Failed to generate key and certificate")
	}

	certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw})
	keyBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	return csh.populateSecrets(keyBytes, certBytes)
}

func (csh *certSetupHandler) certAndKeyProvided() bool {
	return csh.options.caKey != "" && csh.options.caCertificate != ""
}

func (csh *certSetupHandler) generateCertificate(key *rsa.PrivateKey) (*x509.Certificate, error) {
	currentTime := time.Now()

	certTemplate := &x509.Certificate{
		SerialNumber:       big.NewInt(2),
		NotBefore:          currentTime,
		SignatureAlgorithm: x509.SHA256WithRSA,
		IsCA:               true,
		Subject: pkix.Name{
			CommonName: "Kyma",
		},
		NotAfter:              currentTime.Add(csh.options.generatedValidityTime),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, &key.PublicKey, key)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate certificate")
	}

	certificate, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse certificate")
	}

	return certificate, nil
}

func (csh *certSetupHandler) generateKeyAndCertificate() (*rsa.PrivateKey, *x509.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to generate private key")
	}

	certificate, err := csh.generateCertificate(key)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to generate certificate")
	}

	return key, certificate, nil
}

func (csh *certSetupHandler) populateSecrets(pemKey, pemCert []byte) error {
	caCertSecretData := map[string][]byte{
		caCertificateSecretKey: pemCert,
	}

	err := csh.secretRepository.Upsert(csh.options.caCertificateSecret, caCertSecretData)
	if err != nil {
		return errors.Wrap(err, "Failed to update CA certificate secret")
	}

	connectorSecretData := map[string][]byte{
		connectorCertSecretKey: pemCert,
		connectorKeySecretKey:  pemKey,
	}

	err = csh.secretRepository.Upsert(csh.options.connectorCertificateSecret, connectorSecretData)
	if err != nil {
		return errors.Wrap(err, "Failed to update Connector Service secret")
	}

	return nil
}
