package certificaterequest

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/connectorservice"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ccMocks "github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/centralconnection/mocks"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificaterequest/mocks"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates"
	certMocks "github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates/mocks"
	connectorMocks "github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/connectorservice/mocks"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/pkg/apis/applicationconnector/v1alpha1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	certificateRequestName = "cert-request"
	csrInfoURL             = "https://connector-service.kyma.local"
	managementInfoURL      = "https://connector-service.kyma.local/management/info"
)

var (
	emptyConnection = connectorservice.EstablishedConnection{}
)

func TestController_Reconcile(t *testing.T) {
	certs := certificates.Certificates{CRTChain: []byte("cert-chain")}

	establishedConnection := connectorservice.EstablishedConnection{
		ManagementInfoURL: managementInfoURL,
		Certificates:      certs,
	}

	namespacedName := types.NamespacedName{
		Name: certificateRequestName,
	}

	request := reconcile.Request{
		NamespacedName: namespacedName,
	}

	assertErrorStatus := func(args mock.Arguments) {
		certReqInstance := args.Get(1).(*v1alpha1.CertificateRequest)
		assert.NotEmpty(t, certReqInstance.Status.Error)
	}

	t.Run("should request and save certificates", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CertificateRequest")).
			Run(setupCertificateRequestInstance).Return(nil)
		client.On("Delete", context.Background(), mock.AnythingOfType("*v1alpha1.CertificateRequest")).Return(nil)
		connectorClient := &connectorMocks.InitialConnectionClient{}
		connectorClient.On("Establish", csrInfoURL).Return(establishedConnection, nil)
		preserver := &certMocks.Preserver{}
		preserver.On("PreserveCertificates", certs).Return(nil)
		centralConnectionClient := &ccMocks.Client{}
		centralConnectionClient.On("Upsert", certificateRequestName, mock.AnythingOfType("v1alpha1.CentralConnectionSpec")).Return(nil, nil)

		certificateRequestController := newCertificatesRequestController(client, connectorClient, preserver, centralConnectionClient)

		// when
		result, err := certificateRequestController.Reconcile(request)

		// then
		require.NoError(t, err)
		assert.Empty(t, result)
		assertExpectations(t, &client.Mock, &connectorClient.Mock, &preserver.Mock)
	})

	t.Run("should not fetch certificate if error status present on CR", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CertificateRequest")).
			Run(setupCertificateRequestWithErrorStatus).Return(nil)
		connectorClient := &connectorMocks.InitialConnectionClient{}
		preserver := &certMocks.Preserver{}
		centralConnectionClient := &ccMocks.Client{}

		certificateRequestController := newCertificatesRequestController(client, connectorClient, preserver, centralConnectionClient)

		// when
		result, err := certificateRequestController.Reconcile(request)

		// then
		require.NoError(t, err)
		assert.Empty(t, result)
		assertExpectations(t, &client.Mock, &connectorClient.Mock, &preserver.Mock)
	})

	t.Run("should not fail if instance not found", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CertificateRequest")).
			Return(k8sErrors.NewNotFound(schema.GroupResource{}, "error"))
		connectorClient := &connectorMocks.InitialConnectionClient{}
		preserver := &certMocks.Preserver{}
		centralConnectionClient := &ccMocks.Client{}

		certificateRequestController := newCertificatesRequestController(client, connectorClient, preserver, centralConnectionClient)

		// when
		result, err := certificateRequestController.Reconcile(request)

		// then
		require.NoError(t, err)
		assert.Empty(t, result)
		assertExpectations(t, &client.Mock, &connectorClient.Mock, &preserver.Mock)
	})

	t.Run("should return error if failed to get instance", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CertificateRequest")).
			Return(errors.New("error"))
		connectorClient := &connectorMocks.InitialConnectionClient{}
		preserver := &certMocks.Preserver{}
		centralConnectionClient := &ccMocks.Client{}

		certificateRequestController := newCertificatesRequestController(client, connectorClient, preserver, centralConnectionClient)

		// when
		result, err := certificateRequestController.Reconcile(request)

		// then
		require.Error(t, err)
		assert.Empty(t, result)
		assertExpectations(t, &client.Mock, &connectorClient.Mock, &preserver.Mock)
	})

	t.Run("should set error status when failed to request certificate", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CertificateRequest")).
			Run(setupCertificateRequestInstance).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CertificateRequest")).
			Run(assertErrorStatus).Return(nil)
		connectorClient := &connectorMocks.InitialConnectionClient{}
		connectorClient.On("Establish", csrInfoURL).Return(emptyConnection, errors.New("error"))
		preserver := &certMocks.Preserver{}
		centralConnectionClient := &ccMocks.Client{}

		certificateRequestController := newCertificatesRequestController(client, connectorClient, preserver, centralConnectionClient)

		// when
		result, err := certificateRequestController.Reconcile(request)

		// then
		require.NoError(t, err)
		assert.Empty(t, result)
		assertExpectations(t, &client.Mock, &connectorClient.Mock, &preserver.Mock)
	})

	t.Run("should set error status when failed to preserve certificate", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CertificateRequest")).
			Run(setupCertificateRequestInstance).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CertificateRequest")).
			Run(assertErrorStatus).Return(nil)
		connectorClient := &connectorMocks.InitialConnectionClient{}
		connectorClient.On("Establish", csrInfoURL).Return(establishedConnection, nil)
		preserver := &certMocks.Preserver{}
		preserver.On("PreserveCertificates", certs).Return(errors.New("error"))
		centralConnectionClient := &ccMocks.Client{}
		centralConnectionClient.On("Upsert", certificateRequestName, mock.AnythingOfType("v1alpha1.CentralConnectionSpec")).Return(nil, nil)

		certificateRequestController := newCertificatesRequestController(client, connectorClient, preserver, centralConnectionClient)

		// when
		result, err := certificateRequestController.Reconcile(request)

		// then
		require.NoError(t, err)
		assert.Empty(t, result)
		assertExpectations(t, &client.Mock, &connectorClient.Mock, &preserver.Mock)
	})

}

func getCertRequestFromArgs(args mock.Arguments) *v1alpha1.CertificateRequest {
	certReqInstance := args.Get(2).(*v1alpha1.CertificateRequest)
	certReqInstance.Name = certificateRequestName
	return certReqInstance
}

func setupCertificateRequestInstance(args mock.Arguments) {
	certReqInstance := getCertRequestFromArgs(args)
	certReqInstance.Spec.CSRInfoURL = csrInfoURL
}

func setupCertificateRequestWithErrorStatus(args mock.Arguments) {
	certReqInstance := getCertRequestFromArgs(args)
	certReqInstance.Status.Error = "Error"
}

func assertExpectations(t *testing.T, mocks ...*mock.Mock) {
	for _, m := range mocks {
		m.AssertExpectations(t)
	}
}
