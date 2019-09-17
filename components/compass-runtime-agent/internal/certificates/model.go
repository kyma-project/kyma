package certificates

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"

	"github.com/pkg/errors"
)

type Credentials struct {
	ClientCredentials
	CACertificates []*x509.Certificate
}

type ClientCredentials struct {
	ClientKey         *rsa.PrivateKey
	CertificateChain  []*x509.Certificate
	ClientCertificate *x509.Certificate
}

type PemEncodedCredentials struct {
	ClientKey         []byte
	CertificateChain  []byte
	ClientCertificate []byte
	CACertificates    []byte
}

func (c ClientCredentials) AsTLSCertificate() tls.Certificate {
	var rawCerts [][]byte

	for _, cert := range c.CertificateChain {
		rawCerts = append(rawCerts, cert.Raw)
	}

	return tls.Certificate{
		PrivateKey:  c.ClientKey,
		Certificate: rawCerts,
	}
}

func (c Credentials) AsPemEncoded() PemEncodedCredentials {
	return PemEncodedCredentials{
		ClientKey:         pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(c.ClientKey)}),
		CertificateChain:  toPem(c.CertificateChain...),
		ClientCertificate: toPem(c.ClientCertificate),
		CACertificates:    toPem(c.CACertificates...),
	}
}

func toPem(certificates ...*x509.Certificate) []byte {
	certChainPem := make([]byte, 0)
	for _, cert := range certificates {
		certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
		certChainPem = append(certChainPem, certBytes...)
	}

	return certChainPem
}

func (c PemEncodedCredentials) AsClientCredentials() (ClientCredentials, error) {
	certificateChain, err := decodeCertificates(c.CertificateChain)
	if err != nil {
		return ClientCredentials{}, errors.Wrap(err, "Failed to decode certificate chain")
	}

	clientCertificate, err := decodeCertificate(c.ClientCertificate)
	if err != nil {
		return ClientCredentials{}, errors.Wrap(err, "Failed to decode client certificate")
	}

	clientKey, err := getClientPrivateKey(c.ClientKey)
	if err != nil {
		return ClientCredentials{}, errors.Wrap(err, "Failed to decode client key")
	}

	return ClientCredentials{
		ClientKey:         clientKey,
		CertificateChain:  certificateChain,
		ClientCertificate: clientCertificate,
	}, nil
}

func (c PemEncodedCredentials) AsCredentials() (Credentials, error) {
	clientCredentials, err := c.AsClientCredentials()
	if err != nil {
		return Credentials{}, err
	}

	caCerts, err := decodeCertificates(c.CACertificates)
	if err != nil {
		return Credentials{}, errors.Wrap(err, "Failed to decode CA certificate")
	}

	return Credentials{
		ClientCredentials: clientCredentials,
		CACertificates:    caCerts,
	}, nil
}
