package certificates

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
)

type Credentials struct {
	PrivateKey        *rsa.PrivateKey
	CertificateChain  []*x509.Certificate
	ClientCertificate *x509.Certificate
	CACertificates    []*x509.Certificate
}

func (c Credentials) AsTLSCertificate() tls.Certificate {
	var rawCerts [][]byte

	for _, cert := range c.CertificateChain {
		rawCerts = append(rawCerts, cert.Raw)
	}

	return tls.Certificate{
		PrivateKey:  c.PrivateKey,
		Certificate: rawCerts,
	}
}
