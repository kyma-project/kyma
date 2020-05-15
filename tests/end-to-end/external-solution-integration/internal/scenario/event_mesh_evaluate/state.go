package event_mesh_evaluate

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/common/resilient"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/event_mesh_prepare"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

type e2EEventMeshState struct {
	scenario.E2EState
	ServiceClassID string
	registryClient *testkit.RegistryClient
	dataStore      *testkit.DataStore
}

// // SetServiceClassID allows to set ServiceClassID so it can be shared between steps
// func (s *e2EEventMeshState) SetServiceClassID(serviceID string) {
// 	s.ServiceClassID = serviceID
//
// }

// // GetServiceClassID allows to get ServiceClassID so it can be shared between steps
// func (s *e2EEventMeshState) GetServiceClassID() string {
// 	return s.ServiceClassID
// }

func (s *e2EEventMeshState) SetDataStore(dataStore *testkit.DataStore) {
	s.dataStore = dataStore
}

func (s *e2EEventMeshState) GetDataStore() *testkit.DataStore {
	return s.dataStore
}

// SetGatewayClientCerts allows to set application gateway client certificates so they can be used by later steps
func (s *e2EEventMeshState) SetGatewayClientCerts([]tls.Certificate) {
	var err error
	s.AppName, err = s.GetDataStore().Load(event_mesh_prepare.AppNameKey)
	if err != nil {
		panic(err)
	}
	certsJson, err := s.GetDataStore().Load(event_mesh_prepare.CertKey)
	if err != nil {
		panic(err)
	}
	var pemcerts []event_mesh_prepare.PEMCertificate
	if err := json.Unmarshal([]byte(certsJson), &pemcerts); err != nil {
		panic(err)
	}

	certs := make([]tls.Certificate, len(pemcerts))
	for i, pemcert := range pemcerts {
		cert, err := event_mesh_prepare.PEMToCertificate(pemcert)
		if err != nil {
			panic(err)
		}
		certs[i] = cert
	}

	metadataURL := fmt.Sprintf("https://gateway.%s/%s/v1/metadata/services", s.Domain, s.AppName)

	if err != nil {
		panic(err)
	}

	httpClient := internal.NewHTTPClient(
		internal.WithSkipSSLVerification(s.SkipSSLVerify),
	)
	httpClient.Transport.(*http.Transport).TLSClientConfig.Certificates = certs

	s.RegistryClient = testkit.NewRegistryClient(metadataURL, resilient.WrapHttpClient(httpClient))
	s.EventSender = testkit.NewEventSender(httpClient, s.Domain, s.AppName)
}
