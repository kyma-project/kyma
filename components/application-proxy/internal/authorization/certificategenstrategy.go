package authorization

import (
	"crypto/tls"
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
		return apperrors.Internal("Failed to prepare certificate, %s", err.Error())
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

func (b certificateGenStrategy) prepareCertificate() (tls.Certificate, error) {
	return tls.X509KeyPair(b.certificate, b.privateKey)
}
