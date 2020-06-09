package event_mesh_evaluate

import (
	"crypto/tls"
	"net/http"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

type e2EEventMeshState struct {
	scenario.E2EState
}

// SetGatewayClientCerts allows to set application gateway client certificates so they can be used by later steps
func (s *e2EEventMeshState) SetGatewayClientCerts(certs []tls.Certificate) {
	httpClient := internal.NewHTTPClient(
		internal.WithSkipSSLVerification(s.SkipSSLVerify),
	)
	httpClient.Transport.(*http.Transport).TLSClientConfig.Certificates = certs
	s.EventSender = testkit.NewEventSender(httpClient, s.Domain, s.AppName)
}

func (s *e2EEventMeshState) SetApplicationName(name string) {
	s.AppName = name
}
