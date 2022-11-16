package internal

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	b64 "encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

type CABundle struct {
	CA    string
	CAKey string
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

	encodedCA := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caSignedBytes,
	})
	//if err != nil {
	//	return nil, fmt.Errorf("failed to encode signed ca certificate: %v", err)
	//}

	//encodedCAPrivateKey := b64.URLEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(caPrivateKey))
	//encodedCAPrivateKey := new(bytes.Buffer)
	encodedCAPrivateKey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey),
	})
	//if err != nil {
	//	return nil, fmt.Errorf("failed to encode ca private key: %v", err)
	//}

	return &CABundle{
		CA:    b64.URLEncoding.EncodeToString(encodedCA),
		CAKey: b64.URLEncoding.EncodeToString(encodedCAPrivateKey),
	}, nil
}

func GenerateServerCertAndKey(ca *CABundle, serviceName, namespace string) (*ServerCertAndKey, error) {
	decodedKey, err := b64.URLEncoding.DecodeString(ca.CA)
	pemBlock, _ := pem.Decode(decodedKey)
	if pemBlock == nil {
		return nil, fmt.Errorf("pem.Decode failed")
	}

	caCert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ca: %v", err)
	}

	//caCert, err := x509.ParsePKIXPublicKey(decodedKey)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to parse ca: %v", err)
	//}

	fmt.Print(caCert)

	decodedKey, err = b64.URLEncoding.DecodeString(ca.CAKey)
	pemBlock, _ = pem.Decode(decodedKey)
	if pemBlock == nil {
		return nil, fmt.Errorf("pem.Decode failed")
	}
	//der, err := x509.DecryptPEMBlock(pemBlock, []byte("ca private key password"))
	//if err != nil {
	//	panic(err)
	//}
	caPrivateKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		panic(err)
	}

	fmt.Print(caPrivateKey)
	//caKey, err := x509.ParsePKCS1PrivateKey([]byte(ca.CAKey))
	//if err != nil {
	//	return nil, fmt.Errorf("failed to parse ca key: %v", err)
	//}

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

	serverCertSignedBytes, err := x509.CreateCertificate(cryptorand.Reader, serverCert, caCert, &serverPrivateKey.PublicKey, caPrivateKey)
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
