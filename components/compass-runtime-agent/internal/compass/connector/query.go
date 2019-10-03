package connector

import "fmt"

type queryProvider struct{}

func (qp queryProvider) configuration() string {
	return fmt.Sprintf(`query {
	result: configuration() {
		%s
	}
}`, configurationResult())
}

func (qp queryProvider) signCSR(csr string) string {
	return fmt.Sprintf(`mutation {
	result: signCertificateSigningRequest(csr: "%s") {
		%s
	}
}`, csr, certificationResult())
}

func configurationResult() string {
	return `token { token }
		certificateSigningRequestInfo { subject keyAlgorithm }
		managementPlaneInfo { 
			directorURL
			certificateSecuredConnectorURL
		}`
}

func certificationResult() string {
	return `certificateChain
		caCertificate
		clientCertificate`
}
