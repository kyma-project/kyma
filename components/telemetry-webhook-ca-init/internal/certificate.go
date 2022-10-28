package internal

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

type CertificateAuthority *bytes.Buffer
type ServerCertificate *bytes.Buffer
type ServerPrivateKey *bytes.Buffer

type CABundle struct {
	CA            CertificateAuthority
	ServerCert    ServerCertificate
	ServerPrivKey ServerPrivateKey
}

func CreateCABundle(caName string) (*CABundle, error) {
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

	caPrivateKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CA private key: %v", err)
	}

	caSignedBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create self-signed CA certificate: %v", err)
	}

	encodedCA := new(bytes.Buffer)
	_ = pem.Encode(encodedCA, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caSignedBytes,
	})

	dnsNames := []string{"webhook-service", "webhook-service.default", "webhook-service.default.svc"}
	commonName := "webhook-service.default.svc"

	serverCert := &x509.Certificate{
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

	serverPrivateKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate server private key: %v", err)
	}

	serverCertSignedBytes, err := x509.CreateCertificate(cryptorand.Reader, serverCert, ca, &serverPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign server certificate: %v", err)
	}

	encodedServerCert := new(bytes.Buffer)
	err = pem.Encode(encodedServerCert, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertSignedBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to PEM encode server certificate: %v", err)
	}

	encodedServerPrivateKey := new(bytes.Buffer)
	err = pem.Encode(encodedServerPrivateKey, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to PEM encode server private key: %v", err)
	}

	return &CABundle{
		CA:            encodedCA,
		ServerCert:    encodedServerCert,
		ServerPrivKey: encodedServerPrivateKey,
	}, nil
}
