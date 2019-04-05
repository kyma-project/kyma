package centralconnection

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"testing"
	"time"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime/schema"

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
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
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

	clientCert = []byte(`-----BEGIN CERTIFICATE-----
MIICIzCCAYwCCQDDkk/CKHDcZjANBgkqhkiG9w0BAQUFADASMRAwDgYDVQQKEwdB
Y21lIENvMCAXDTE5MDMyOTEzMjU1M1oYDzIxMTkwMzA1MTMyNTUzWjAUMRIwEAYD
VQQDDAlsb2NhbGhvc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDr
D1AQps/wSEUTrMA8hWUXLBkQqpIgAFQPOg9kctISsiu782mIjz6Om/8LF4vuTTf5
s/wGdR8nfWGhMvdAMUi9trn5GKLEN2McDB5Pa7T7kA/54/DHkkgBJ2lgIoOxh6Dy
jsHQQ5hM/BPSCZ/xpXys5o0PVm8nroftGGq6Ij/pXfhpZ4PsgximBCyu0pKrfCs3
ogJkPGyf1uSmfLzWHxb8C34+UFMbPly585Tpqaey44y8bB6jL30b6Nqg8AzkY4D9
3YBUGJnw5Tn1k/lSvwSuXwW+n7Kc8cRoJYYJVXlEZE+y9Bqqe9hqSAlcy8GfozBR
/cCio1OhLzsqrV0KZ5qNAgMBAAEwDQYJKoZIhvcNAQEFBQADgYEAi9t6j7ahK9vZ
VsfqyMGcgeIrI2mzI8oDAHb0xkrKiQpOAGoq9ejBujwDI3L2g2MToHhB0aataCmC
oiCU2Sf1LDG70bnyd0eLKshNEFjHEsVHJkzPwxeOFsM7xuKCZQ4uvnFBZyyQmuyY
QbIjsJhuMRQuka2NB6eGq4qFaHHbkzc=
-----END CERTIFICATE-----`)
	caCert    = []byte("ca-cert")
	certChain = []byte("cert-chain")
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
			connection := args.Get(1).(*v1alpha1.CentralConnection)

			assert.NotEmpty(t, connection.Status.CertificateStatus.NotBefore)
			assert.NotEmpty(t, connection.Status.CertificateStatus.NotAfter)

			assert.Equal(t, connection.Status.SynchronizationStatus.LastSync, connection.Status.SynchronizationStatus.LastSuccess)
		}

		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(setupCentralConnectionInstance).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(checkStatus).Return(nil)

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
		assertExpectations(t, client.Mock, certProvider.Mock, certPreserver.Mock, mutualTLSClient.Mock, mTLSClientProvider.Mock)
	})

	t.Run("should not take action if connection deleted", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Return(k8sErrors.NewNotFound(schema.GroupResource{}, "error"))

		certProvider := &certMocks.Provider{}
		certPreserver := &certMocks.Preserver{}
		mTLSClientProvider := &connectorMocks.MutualTLSClientProvider{}

		connectionController := newCentralConnectionController(client, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.NoError(t, err)
		assertExpectations(t, client.Mock, certProvider.Mock, certPreserver.Mock)
	})

	t.Run("should set error status when failed to read client certificate", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(setupCentralConnectionInstance).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(assertErrorStatus(t)).Return(nil)

		certProvider := &certMocks.Provider{}
		certProvider.On("GetClientCredentials").Return(nil, nil, errors.New("error"))

		certPreserver := &certMocks.Preserver{}
		mTLSClientProvider := &connectorMocks.MutualTLSClientProvider{}

		connectionController := newCentralConnectionController(client, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.Error(t, err)
		assertExpectations(t, client.Mock, certProvider.Mock, certPreserver.Mock)
	})

	t.Run("should set error status when failed to get management info", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(setupCentralConnectionInstance).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(assertErrorStatus(t)).Return(nil)

		certProvider := &certMocks.Provider{}
		certProvider.On("GetClientCredentials").Return(privateKey, certificate, nil)

		certPreserver := &certMocks.Preserver{}

		mutualTLSClient := &connectorMocks.MutualTLSConnectorClient{}
		mutualTLSClient.On("GetManagementInfo", managementInfoURL).
			Return(connectorservice.ManagementInfo{}, errors.New("error"))

		mTLSClientProvider := &connectorMocks.MutualTLSClientProvider{}
		mTLSClientProvider.On("CreateClient", privateKey, certificate).Return(mutualTLSClient)

		connectionController := newCentralConnectionController(client, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.Error(t, err)
		assertExpectations(t, client.Mock, certProvider.Mock, certPreserver.Mock, mutualTLSClient.Mock, mTLSClientProvider.Mock)
	})

	t.Run("should set error status when failed to renew certificate", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(setupCentralConnectionInstance).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(assertErrorStatus(t)).Return(nil)

		certProvider := &certMocks.Provider{}
		certProvider.On("GetClientCredentials").Return(privateKey, certificate, nil)

		certPreserver := &certMocks.Preserver{}

		mutualTLSClient := &connectorMocks.MutualTLSConnectorClient{}
		mutualTLSClient.On("GetManagementInfo", managementInfoURL).Return(managementInfo, nil)
		mutualTLSClient.On("RenewCertificate", renewalURL).Return(certificates.Certificates{}, errors.New("error"))

		mTLSClientProvider := &connectorMocks.MutualTLSClientProvider{}
		mTLSClientProvider.On("CreateClient", privateKey, certificate).Return(mutualTLSClient)

		connectionController := newCentralConnectionController(client, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.Error(t, err)
		assertExpectations(t, client.Mock, certProvider.Mock, certPreserver.Mock, mutualTLSClient.Mock, mTLSClientProvider.Mock)
	})

	t.Run("should set error status when failed to preserve certificates", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(setupCentralConnectionInstance).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(assertErrorStatus(t)).Return(nil)

		certProvider := &certMocks.Provider{}
		certProvider.On("GetClientCredentials").Return(privateKey, certificate, nil)

		certPreserver := &certMocks.Preserver{}
		certPreserver.On("PreserveCertificates", renewedCertificates).Return(errors.New("error"))

		mutualTLSClient := &connectorMocks.MutualTLSConnectorClient{}
		mutualTLSClient.On("GetManagementInfo", managementInfoURL).Return(managementInfo, nil)
		mutualTLSClient.On("RenewCertificate", renewalURL).Return(renewedCertificates, nil)

		mTLSClientProvider := &connectorMocks.MutualTLSClientProvider{}
		mTLSClientProvider.On("CreateClient", privateKey, certificate).Return(mutualTLSClient)

		connectionController := newCentralConnectionController(client, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.Error(t, err)
		assertExpectations(t, client.Mock, certProvider.Mock, certPreserver.Mock, mutualTLSClient.Mock, mTLSClientProvider.Mock)
	})

	t.Run("should set error status when failed to decode pem", func(t *testing.T) {
		// given
		invalidCerts := certificates.Certificates{
			ClientCRT: []byte("invalid cert"),
		}

		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(setupCentralConnectionInstance).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(assertErrorStatus(t)).Return(nil)

		certProvider := &certMocks.Provider{}
		certProvider.On("GetClientCredentials").Return(privateKey, certificate, nil)

		certPreserver := &certMocks.Preserver{}
		certPreserver.On("PreserveCertificates", invalidCerts).Return(nil)

		mutualTLSClient := &connectorMocks.MutualTLSConnectorClient{}
		mutualTLSClient.On("GetManagementInfo", managementInfoURL).Return(managementInfo, nil)
		mutualTLSClient.On("RenewCertificate", renewalURL).Return(invalidCerts, nil)

		mTLSClientProvider := &connectorMocks.MutualTLSClientProvider{}
		mTLSClientProvider.On("CreateClient", privateKey, certificate).Return(mutualTLSClient)

		connectionController := newCentralConnectionController(client, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.Error(t, err)
		assertExpectations(t, client.Mock, certProvider.Mock, certPreserver.Mock, mutualTLSClient.Mock, mTLSClientProvider.Mock)
	})
}

func assertErrorStatus(t *testing.T) func(args mock.Arguments) {
	return func(args mock.Arguments) {
		connection := args.Get(1).(*v1alpha1.CentralConnection)

		assert.NotEmpty(t, connection.Status.Error.Message)
		assert.NotEqual(t, connection.Status.SynchronizationStatus.LastSuccess, connection.Status.SynchronizationStatus.LastSync)
	}
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

func assertExpectations(t *testing.T, mocks ...mock.Mock) {
	for _, m := range mocks {
		m.AssertExpectations(t)
	}
}
