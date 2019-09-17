package compass

import "fmt"

func Configuration() string {
	return fmt.Sprintf(`query {
	result: configuration() {
		%s
	}
}`, configurationResult())
}

func SignCSR(csr string) string {
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
