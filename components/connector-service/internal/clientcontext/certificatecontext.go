package clientcontext

type clientCertificateContext struct {
	clientContextService
	subject string
}

func newClientCertificateContext(clientContext clientContextService, subject string) *clientCertificateContext {
	return &clientCertificateContext{
		clientContextService: clientContext,
		subject:              subject, // TODO - should it be certificates.CSRSubject?
	}
}

func (cc *clientCertificateContext) GetSubject() string {
	return cc.subject
}
