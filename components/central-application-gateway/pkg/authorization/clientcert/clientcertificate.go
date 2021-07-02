package clientcert

import (
	"crypto/tls"
	"errors"
	"sync"
)

var ErrMissingClientCertificate = errors.New("required client certificate is missing")

type SetClientCertificateFunc func(cert *tls.Certificate)

type ClientCertificate interface {
	// GetClientCertificate implements the TLSClientConfig.GetClientCertificate function
	// which is called when a server requests a certificate from a client.
	GetClientCertificate(*tls.CertificateRequestInfo) (*tls.Certificate, error)

	// SetCertificate sets the client certificates.
	SetCertificate(certificate *tls.Certificate)
	// GetCertificate returns the client certificates.
	GetCertificate() *tls.Certificate
}

func NewClientCertificate(certificate *tls.Certificate) ClientCertificate {
	return &clientCertificate{
		certificate: certificate,
	}
}

type clientCertificate struct {
	sync.RWMutex
	certificate *tls.Certificate
}

func (c *clientCertificate) GetClientCertificate(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
	c.RLock()
	defer c.RUnlock()

	if c.certificate == nil {
		return nil, ErrMissingClientCertificate
	}
	return c.certificate, nil
}

func (c *clientCertificate) GetCertificate() *tls.Certificate {
	c.RLock()
	defer c.RUnlock()

	return c.certificate
}

func (c *clientCertificate) SetCertificate(cert *tls.Certificate) {
	c.Lock()
	defer c.Unlock()

	c.certificate = cert
}
