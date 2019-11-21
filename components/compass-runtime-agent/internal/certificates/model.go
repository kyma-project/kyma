package certificates

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	gqlschema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
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

func NewCredentials(key *rsa.PrivateKey, certificateResponse gqlschema.CertificationResult) (Credentials, error) {
	pemCertChain, err := base64.StdEncoding.DecodeString(certificateResponse.CertificateChain)
	if err != nil {
		return Credentials{}, errors.Wrap(err, "Failed to decode base 64 certificate chain")
	}
	certificateChain, err := decodeCertificates(pemCertChain)
	if err != nil {
		return Credentials{}, errors.Wrap(err, "Failed to decode certificate chain")
	}
	pemClientCert, err := base64.StdEncoding.DecodeString(certificateResponse.ClientCertificate)
	if err != nil {
		return Credentials{}, errors.Wrap(err, "Failed to decode base 64 client certificate")
	}
	clientCert, err := decodeCertificate(pemClientCert)
	if err != nil {
		return Credentials{}, errors.Wrap(err, "Failed to decode client certificate")
	}
	pemCACert, err := base64.StdEncoding.DecodeString(certificateResponse.CaCertificate)
	if err != nil {
		return Credentials{}, errors.Wrap(err, "Failed to decode base 64 CA certificate")
	}
	caCerts, err := decodeCertificates(pemCACert)
	if err != nil {
		return Credentials{}, errors.Wrap(err, "Failed to decode CA certificate")
	}

	return Credentials{
		ClientCredentials: ClientCredentials{
			ClientKey:         key,
			CertificateChain:  certificateChain,
			ClientCertificate: clientCert,
		},
		CACertificates: caCerts,
	}, nil
}

func ParsePrivateKey(clusterKey []byte) (*rsa.PrivateKey, error) {
	if clusterKey == nil {
		return nil, errors.New("Private key data is empty")
	}

	block, _ := pem.Decode(clusterKey)
	if block == nil {
		return nil, errors.New("Failed to decode client key pem")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

type PemEncodedCredentials struct {
	ClientKey         []byte
	CertificateChain  []byte
	ClientCertificate []byte
	CACertificates    []byte
}

func (c ClientCredentials) AsTLSCertificate() *tls.Certificate {
	var rawCerts [][]byte

	for _, cert := range c.CertificateChain {
		rawCerts = append(rawCerts, cert.Raw)
	}

	return &tls.Certificate{
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

	clientKey, err := ParsePrivateKey(c.ClientKey)
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
