package clientcontext

import "github.com/kyma-project/kyma/components/application-connector/connector-service/internal/certificates"

type clientCertificateContext struct {
	ClientContextService
	subject certificates.CSRSubject
}

func newClientCertificateContext(clientContext ClientContextService, subject certificates.CSRSubject) *clientCertificateContext {
	return &clientCertificateContext{
		ClientContextService: clientContext,
		subject:              subject,
	}
}

func (cc *clientCertificateContext) GetSubject() certificates.CSRSubject {
	return cc.subject
}

func (cc *clientCertificateContext) ClientContext() ClientContextService {
	return cc.ClientContextService
}
