package authorization

import (
	"crypto/tls"
	"net/http"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"
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

func (b certificateGenStrategy) AddAuthorization(r *http.Request, setter clientcert.SetClientCertificateFunc) apperrors.AppError {
	cert, err := b.prepareCertificate()
	if err != nil {
		return apperrors.Internal("Failed to prepare certificate, %s", err.Error())
	}
	setter(&cert)
	return nil
}

func (b certificateGenStrategy) Invalidate() {
}

func (b certificateGenStrategy) prepareCertificate() (tls.Certificate, error) {
	return tls.X509KeyPair(b.certificate, b.privateKey)
}
