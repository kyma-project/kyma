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

type CA []byte
type CAKey []byte
type CABundle struct {
	CA    CA
	CAKey CAKey
}

type ServerCert []byte
type ServerPrivateKey []byte
type ServerCertAndKey struct {
	Cert ServerCert
	Key  ServerPrivateKey
}

func GenerateCACert() (*CABundle, error) {
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
		return nil, fmt.Errorf("failed to generate ca private key: %v", err)
	}

	caSignedBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create self-signed ca certificate: %v", err)
	}

	encodedCA := new(bytes.Buffer)
	err = pem.Encode(encodedCA, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caSignedBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode signed ca certificate: %v", err)
	}

	encodedCAPrivateKey := new(bytes.Buffer)
	err = pem.Encode(encodedCAPrivateKey, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode ca private key: %v", err)
	}

	return &CABundle{
		CA:    encodedCA.Bytes(),
		CAKey: encodedCAPrivateKey.Bytes(),
	}, nil
}

func GenerateServerCertAndKey(ca *CABundle, serviceName, namespace string) (*ServerCertAndKey, error) {
	caCert, err := x509.ParseCertificate(ca.CA)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ca: %v", err)
	}

	caKey, err := x509.ParsePKCS1PrivateKey(ca.CAKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ca key: %v", err)
	}

	dnsNames := []string{
		serviceName,
		fmt.Sprintf("%s.%s", serviceName, namespace),
		fmt.Sprintf("%s.%s.svc", serviceName, namespace),
	}

	serverCert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   dnsNames[2],
			Organization: []string{"kyma-project.io"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(0, 2, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	serverPrivateKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate server private key: %v", err)
	}

	serverCertSignedBytes, err := x509.CreateCertificate(cryptorand.Reader, serverCert, caCert, &serverPrivateKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign server certificate: %v", err)
	}

	encodedServerCert := new(bytes.Buffer)
	err = pem.Encode(encodedServerCert, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertSignedBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to pem encode server certificate: %v", err)
	}

	encodedServerPrivateKey := new(bytes.Buffer)
	err = pem.Encode(encodedServerPrivateKey, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to pem encode server private key: %v", err)
	}

	return &ServerCertAndKey{
		Cert: encodedServerCert.Bytes(),
		Key:  encodedServerPrivateKey.Bytes(),
	}, nil
}
