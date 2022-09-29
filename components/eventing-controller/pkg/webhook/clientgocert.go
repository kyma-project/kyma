package webhook

import (
	"crypto/x509"
	"net"

	"k8s.io/client-go/util/cert"
)

type IClientGoCert interface {
	generateSelfSignedCertKey(host string, alternateIPs []net.IP, alternateDNS []string) ([]byte, []byte, error)
	parseCertsPEM(pemCerts []byte) ([]*x509.Certificate, error)
	newPoolFromBytes(pemBlock []byte) (*x509.CertPool, error)
}

type ClientGoCert struct{}

func (r *ClientGoCert) generateSelfSignedCertKey(host string, alternateIPs []net.IP, alternateDNS []string) ([]byte, []byte, error) {
	return cert.GenerateSelfSignedCertKey(host, alternateIPs, alternateDNS)
}

func (r *ClientGoCert) parseCertsPEM(pemCerts []byte) ([]*x509.Certificate, error) {
	return cert.ParseCertsPEM(pemCerts)
}

func (r *ClientGoCert) newPoolFromBytes(pemBlock []byte) (*x509.CertPool, error) {
	return cert.NewPoolFromBytes(pemBlock)
}
