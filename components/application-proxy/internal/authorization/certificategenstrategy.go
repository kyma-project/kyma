package authorization

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httputil"

	"github.com/kyma-project/kyma/components/application-proxy/internal/apperrors"
)

type certificateGenStrategy struct {
	certificate []byte
	privateKey  []byte
}

func newCertificateGenStrategy(certificate, privateKey []byte) certificateGenStrategy {
	return certificateGenStrategy{
		certificate: certificate,
		privateKey:  privateKey,
	}
}

func (b certificateGenStrategy) AddAuthorization(r *http.Request, proxy *httputil.ReverseProxy) apperrors.AppError {
	cert, err := b.prepareCertificate()
	if err != nil {
		return err
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	proxy.Transport = transport

	return nil
}

func (b certificateGenStrategy) Invalidate() {
}

func (b certificateGenStrategy) prepareCertificate() (tls.Certificate, apperrors.AppError) {
	key, err := parseKey(b.privateKey)
	if err != nil {
		return tls.Certificate{}, apperrors.Internal("Failed to parse private key, %s", err.Error())
	}

	certificate, _ := pem.Decode(b.certificate)
	if certificate == nil {
		return tls.Certificate{}, apperrors.Internal("Empty certificate pem block")
	}

	return tls.Certificate{
		PrivateKey:  key,
		Certificate: [][]byte{certificate.Bytes},
	}, nil
}

func parseKey(encodedData []byte) (*rsa.PrivateKey, error) {
	pemBlock, _ := pem.Decode(encodedData)
	if pemBlock == nil {
		return nil, apperrors.Internal("Empty key pem block")
	}

	return x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
}
