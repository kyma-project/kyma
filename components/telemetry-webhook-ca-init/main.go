package main

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/go-logr/logr"
	"math/big"
	"os"
	"telemetry-webhook-ca-init/internal"
	"time"
)

// const certDir = "/etc/webhook/certs/"
const certDir = "./bin/"

var log logr.Logger

type certificateAuthority *bytes.Buffer
type serverCertificate *bytes.Buffer
type serverPrivateKey *bytes.Buffer

func main() {
	log = internal.InitLogger()

	var caPEM, serverCertPEM, serverPrivKeyPEM *bytes.Buffer

	// CA config
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			Organization: []string{"kyma-project.io"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// CA private key
	caPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	exitOnErr(err, "failed to generate CA private key")

	// Self-signed CA certificate
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	exitOnErr(err, "failed to create self-signed CA certificate")

	// PEM encode CA certificate
	caPEM = new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	dnsNames := []string{"webhook-service", "webhook-service.default", "webhook-service.default.svc"}
	commonName := "webhook-service.default.svc"

	// server certificate config
	cert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"kyma-project.io"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	serverPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	exitOnErr(err, "failed to generate server private key")

	// server certificate signing
	serverCertBytes, err := x509.CreateCertificate(cryptorand.Reader, cert, ca, &serverPrivKey.PublicKey, caPrivKey)
	exitOnErr(err, "failed to sign server certificate")

	// PEM encode the server certificate and key
	serverCertPEM = new(bytes.Buffer)
	_ = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})

	serverPrivKeyPEM = new(bytes.Buffer)
	_ = pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})

	err = os.MkdirAll(certDir, 0777)
	exitOnErr(err, "failed to create certs directory")

	err = writeFile(certDir+"tls.crt", serverCertPEM)
	exitOnErr(err, "failed to write tls.crt")

	err = writeFile(certDir+"tls.key", serverPrivKeyPEM)
	exitOnErr(err, "failed to write tls.key")
}

func writeFile(filepath string, sCert *bytes.Buffer) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = f.Write(sCert.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func exitOnErr(err error, msg string) {
	if err != nil {
		log.Error(err, msg)
		os.Exit(1)
	}
}
