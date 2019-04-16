package clientcontext

import "github.com/kyma-project/kyma/components/connector-service/internal/certificates"

type clientCertificateContext struct {
	clientContextService
	subject certificates.CSRSubject
}

func newClientCertificateContext(clientContext clientContextService, subject certificates.CSRSubject) *clientCertificateContext {
	return &clientCertificateContext{
		clientContextService: clientContext,
		subject:              subject,
	}
}

func (cc *clientCertificateContext) GetSubject() certificates.CSRSubject {
	return cc.subject
}
