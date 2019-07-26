package certificates

import "crypto/x509"

type Credentials struct {
	CertificateChain  []*x509.Certificate
	ClientCertificate *x509.Certificate
	CACertificates    []*x509.Certificate
}
