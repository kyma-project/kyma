package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/sirupsen/logrus"

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
	logrus.Infoln("Checking if certificate and key provided...")
	if csh.certAndKeyProvided() {
		caKey, caCert, err := csh.validateProvidedCertAndKey(csh.options.caKey, csh.options.caCertificate)
		if err == nil {
			logrus.Infoln("Valid certificate and key provided. Skipping generation.")
			return csh.populateSecrets(caKey, caCert)
		}

		logrus.Warningf("Certificate or key is invalid: %s", err.Error())
	}

	namespacedName := csh.options.connectorCertificateSecret
	namespacedName.Name += "-backup"
	provided, err := csh.backupCertAndKeyProvided(namespacedName)
	if err != nil {
		return errors.Wrap(err, "Failed to check if certificate backup exists")
	}

	if provided {
		secretData, err := csh.secretRepository.Get(namespacedName)
		if err != nil && !k8serrors.IsNotFound(err) {
			return errors.Wrap(err, "Failed to get backup data")
		}

		caKey, caCert, err := csh.validateProvidedCertAndKeyBytes(
			secretData[connectorKeySecretKey],
			secretData[connectorCertSecretKey])
		if err == nil {
			logrus.Infoln("Valid backup certificate and key provided. Skipping generation.")
			return csh.populateSecrets(caKey, caCert)
		}

		logrus.Warningf("Backup certificate or key is invalid or not found: %s", err.Error())
	}

	logrus.Infoln("Checking if certificate and key exists in Secrets...")
	certsExists, err := csh.certificatesExists()
	if err != nil {
		return errors.Wrap(err, "Failed to check if certificates exists")
	}

	if certsExists {
		logrus.Info("Certificate and key already exist in the Secrets. Skipping generation.")
		return nil
	}

	logrus.Infoln("New key and certificate will be generated.")
	key, certificate, err := csh.generateKeyAndCertificate()
	if err != nil {
		return errors.Wrap(err, "Failed to generate key and certificate")
	}

	certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw})
	keyBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	return csh.populateSecrets(keyBytes, certBytes)
}

func (csh *certSetupHandler) certificatesExists() (bool, error) {
	connectorCertsProvided, err := csh.secretRepository.ValuesProvided(csh.options.connectorCertificateSecret, []string{connectorKeySecretKey, connectorCertSecretKey})
	if err != nil {
		return false, errors.Wrap(err, "Failed to check if Connector Service certs exist in the Secret")
	}

	caCertProvided, err := csh.secretRepository.ValuesProvided(csh.options.caCertificateSecret, []string{caCertificateSecretKey})
	if err != nil {
		return false, errors.Wrap(err, "Failed to check if CA cert exists in the Secret")
	}

	return connectorCertsProvided && caCertProvided, nil
}

func (csh *certSetupHandler) certAndKeyProvided() bool {
	return csh.options.caKey != "" && csh.options.caCertificate != ""
}

func (csh *certSetupHandler) backupCertAndKeyProvided(namespacedName types.NamespacedName) (bool, error) {

	backupConnectorCertsProvided, err := csh.secretRepository.
		ValuesProvided(namespacedName, []string{connectorKeySecretKey, connectorCertSecretKey})

	if err != nil {
		return false, errors.Wrap(err, "Error while loading backup certs or backup not found")
	}

	return backupConnectorCertsProvided, nil
}

func (csh *certSetupHandler) validateProvidedCertAndKey(caKeyInput string, caCertificate string) ([]byte, []byte, error) {
	caKey, err := base64.StdEncoding.DecodeString(caKeyInput)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Failed to decode base64 key: %s", err.Error()))
	}

	caCert, err := base64.StdEncoding.DecodeString(caCertificate)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Failed to decode base64 certificate: %s", err.Error()))
	}

	return csh.validateProvidedCertAndKeyBytes(caCert, caKey)
}

func (csh *certSetupHandler) validateProvidedCertAndKeyBytes(caCert []byte, caKey []byte) ([]byte, []byte, error) {
	_, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Failed to parse key and certificate, key or certificate is invalid: %s", err.Error()))
	}

	return caKey, caCert, nil
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
	logrus.Info("Generating certificate and key")

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
	logrus.Info("Populating secrets with key and certificate")

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
