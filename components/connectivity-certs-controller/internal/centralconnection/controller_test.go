package centralconnection

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/connectorservice"
	connectorMocks "github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/connectorservice/mocks"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/pkg/apis/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificaterequest/mocks"
	certMocks "github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates/mocks"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	centralConnectionName = "central-kyma"

	managementInfoURL = "https://connector-service.cx/management-info"
	renewalURL        = "https://connector-service.cx/renewal"
)

var (
	connectionTime = time.Now()
	validityTime   = connectionTime.Add(2 * time.Hour)

	clientCert = []byte("client-cert")
	caCert     = []byte("ca-cert")
	certChain  = []byte("cert-chain")
)

func TestController_Reconcile(t *testing.T) {

	namespacedName := types.NamespacedName{
		Name: centralConnectionName,
	}

	request := reconcile.Request{
		NamespacedName: namespacedName,
	}

	privateKey := &rsa.PrivateKey{}
	certificate := &x509.Certificate{}

	renewedCertificates := certificates.Certificates{
		ClientCRT: clientCert,
		CaCRT:     caCert,
		CRTChain:  certChain,
	}

	managementInfo := connectorservice.ManagementInfo{
		ManagementURLs: connectorservice.ManagementURLs{
			RenewalURL: renewalURL,
		},
	}

	t.Run("should check connection and renew certificate", func(t *testing.T) {
		// given
		checkStatus := func(args mock.Arguments) {
			centralConnection := args.Get(1).(*v1alpha1.CentralConnection)

			assert.Equal(t, metav1.NewTime(connectionTime), centralConnection.Status.CertificateStatus.NotBefore)
			assert.Equal(t, metav1.NewTime(validityTime), centralConnection.Status.CertificateStatus.NotAfter)
		}

		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(setupCentralConnectionInstance).Return(nil)
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CentralConnection")).Run(checkStatus).Return(nil)

		certProvider := &certMocks.Provider{}
		certProvider.On("GetClientCredentials").Return(privateKey, certificate, nil)

		certPreserver := &certMocks.Preserver{}
		certPreserver.On("PreserveCertificates", renewedCertificates).Return(nil)

		mutualTLSClient := &connectorMocks.MutualTLSConnectorClient{}
		mutualTLSClient.On("GetManagementInfo", managementInfoURL).Return(managementInfo, nil)
		mutualTLSClient.On("RenewCertificate", renewalURL).Return(renewedCertificates, nil)

		mTLSClientProvider := &connectorMocks.MutualTLSClientProvider{}
		mTLSClientProvider.On("CreateClient", privateKey, certificate).Return(mutualTLSClient)

		connectionController := newCentralConnectionController(client, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.NoError(t, err)
	})

	//t.Run("should request and save renewedCertificates", func(t *testing.T) {
	//	// given
	//	client := &mocks.Client{}
	//	client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CertificateRequest")).
	//		Run(setupCentralConnectionInstance).Return(nil)
	//
	//	preserver := &certMocks.Provider{}
	//	preserver.On("GetClientCredentials").Return(nil) // TODO -return
	//
	//	connectionManager := &mocks.CentralConnectionManager{}
	//	connectionManager.On("Create", mock.AnythingOfType("*v1alpha1.CentralConnection")).Return(nil, nil)
	//
	//	certificateRequestController := newCertificatesRequestController(client, connectorClient, preserver, connectionManager)
	//
	//	// when
	//	result, err := certificateRequestController.Reconcile(request)
	//
	//	// then
	//	require.NoError(t, err)
	//	assert.Empty(t, result)
	//	assertExpectations(t, client.Mock, connectorClient.Mock, preserver.Mock)
	//})
}

func getCentralConnectionFromArgs(args mock.Arguments) *v1alpha1.CentralConnection {
	centralConnection := args.Get(2).(*v1alpha1.CentralConnection)
	return centralConnection
}

func setupCentralConnectionInstance(args mock.Arguments) {
	centralConnection := getCentralConnectionFromArgs(args)
	centralConnection.Name = centralConnectionName
	centralConnection.Spec = v1alpha1.CentralConnectionSpec{
		ManagementInfoURL: managementInfoURL,
		EstablishedAt:     metav1.NewTime(connectionTime),
	}
}

//
//func setupCentralConnectionInstance(args mock.Arguments) {
//	certReqInstance := getCertRequestFromArgs(args)
//	certReqInstance.Spec.CSRInfoURL = csrInfoURL
//}
//
//func setupCertificateRequestWithErrorStatus(args mock.Arguments) {
//	certReqInstance := getCertRequestFromArgs(args)
//	certReqInstance.Status.Error = "Error"
//}

func assertExpectations(t *testing.T, mocks ...mock.Mock) {
	for _, m := range mocks {
		m.AssertExpectations(t)
	}
}
