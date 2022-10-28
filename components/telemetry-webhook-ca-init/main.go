package main

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"flag"
	"github.com/go-logr/logr"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/big"
	"os"
	"telemetry-webhook-ca-init/internal"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const caName = "telemetry-validating-webhook-ca"

// const certDir = "/var/run/telemetry-webhook/"
var certDir string

var log logr.Logger

func main() {
	flag.StringVar(&certDir, "cert-dir", "", "Path to certificate bundle directory")
	flag.Parse()
	if err := validateFlags(); err != nil {
		panic(err.Error())
	}
	ctx := context.Background()
	log = internal.InitLogger(ctx)

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	_, err = clientset.CoreV1().Secrets("kyma-system").Get(ctx, "test", v1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

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

func validateFlags() error {
	if certDir == "" {
		return errors.New("--cert-dir flag is required")
	}
	return nil
}

func exitOnErr(err error, msg string) {
	if err != nil {
		log.Error(err, msg)
		os.Exit(1)
	}
}
