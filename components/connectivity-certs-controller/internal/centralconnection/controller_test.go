package centralconnection

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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

	clientKeyPem = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA6w9QEKbP8EhFE6zAPIVlFywZEKqSIABUDzoPZHLSErIru/Np
iI8+jpv/CxeL7k03+bP8BnUfJ31hoTL3QDFIvba5+RiixDdjHAweT2u0+5AP+ePw
x5JIASdpYCKDsYeg8o7B0EOYTPwT0gmf8aV8rOaND1ZvJ66H7RhquiI/6V34aWeD
7IMYpgQsrtKSq3wrN6ICZDxsn9bkpny81h8W/At+PlBTGz5cufOU6amnsuOMvGwe
oy99G+jaoPAM5GOA/d2AVBiZ8OU59ZP5Ur8Erl8Fvp+ynPHEaCWGCVV5RGRPsvQa
qnvYakgJXMvBn6MwUf3AoqNToS87Kq1dCmeajQIDAQABAoIBAQCIlnNN2cDGvRf2
oNFr2Y+ucV93QcZ7dfVii7haBCZx2rpzErRmN+Z/88G17k7PgGtgW+e80N3zknXi
t7zYvkqogr96MYiTQCQFLj2GpO2bqFDAQmWtciEJGp+uzx97T3aEu9N/c2fShD/4
MsOQJTtXNPkOyoj4pAA0E5Yg5roAng7edJwO9fujSbhmJtvnVLz2ZC/06TuC0tDU
cIGsJYqYYGG0lTE41iecYJs1mDwxIofWsl3GHYEU3dMILaWzfvWRAl+msx8TzyOw
4WXgOpcwV7uOtPAHN/OWpjocKbzW6ouqraL8Quft4ioYm9qeRH9mdB13d1H1Sqnc
kMK6Wrm5AoGBAP/O9vHFlbUxhKWvwqcfbb38zaVXbt0TPpnqzthpKxXLFqoPYKR0
fZQs3Gt36Z/pqNioQfzqwvthd9NgzHbNvPlPbvwqHS3C0DsZ22O0VvLPmmbybPtZ
aIExZujZ42iqXpBpv9yr9L7UGyvvAdb5D899+yXqEKb79qnU063RpUKvAoGBAOs8
XvEFQvW2C5UkGANv3zvhWPdHp99bPL/ePI2OIF35NUulbnrxZu39sCo/6hmog3kP
1YH+lT/fsMwNuvlepGztO29+CHOz6KiLGXYwCTwpLZjTREUJba6eiefhnYnGKlpn
AIy2pzs/LJUveRepojJNnthvUjKXLjiULPcri/WDAoGBAKqynK5wvpmOVYmKY0XJ
/x0MGN4AHgZ/1QI4YZafdxSv1Ivefwq+gR3jYaKE/eyrqvQIMyBmN34vaBoxOb79
QuDKVLEIGTh0Cyek9XTu3iZgyhNwKbD/1HCBWr5+xvUM2tVa+6BxTnwYZZlHf97H
i/lVg8WlDz+eWtaxIh+XCcQZAoGAe+XCQ8P3rp8Bnr3x/+1ucIWSbDvLiXLunkgZ
MJ2JIrXdgkhR1mNLSVJy9O3RCU6eYKccV2mVhpz066TXs/xLMiwJQAHrxbUed5c8
A+ntE0jFAVdU/9+la3GJRR6p8ST0rcTOn06c6jGt862bZAEusrv7TBfl/UtvRtGU
lWLURq0CgYEA/QhiaD6alc01yk2bazVt5Ofl3Eqw7yATO21EPymM51LaLMyAw6o0
nXLJnKSbO+eB020eCGMDgcR5FjE46NWcOMLdCINZ2iFkhIJo+M/lEAAXvqVLRcQr
sqXzf1JPPHnmozIitzXi9prXUE3xQ66i95l2GQq2OHHIMyHljtYk9bQ=
-----END RSA PRIVATE KEY-----`

	clientCertificatePem = `-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----`

	caCertificatePem = `-----BEGIN CERTIFICATE-----
MIIDOjCCAiICCQDUf/5116L7fTANBgkqhkiG9w0BAQsFADBfMQ8wDQYDVQQLDAZD
NGNvcmUxDDAKBgNVBAoMA1NBUDEQMA4GA1UEBwwHV2FsZG9yZjEQMA4GA1UECAwH
V2FsZG9yZjELMAkGA1UEBhMCREUxDTALBgNVBAMMBEt5bWEwHhcNMTgwNzEzMDk1
MjUxWhcNMTkwNzEzMDk1MjUxWjBfMQ8wDQYDVQQLDAZDNGNvcmUxDDAKBgNVBAoM
A1NBUDEQMA4GA1UEBwwHV2FsZG9yZjEQMA4GA1UECAwHV2FsZG9yZjELMAkGA1UE
BhMCREUxDTALBgNVBAMMBEt5bWEwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
AoIBAQD49IZuqogcaqAVSV79L7xKMI36NMy6ig+jTquecN9LRhQcalKDKxJ0BRet
bSUhftr8qcE3SaxOtPPvTLoiixjlMFaQ46ZfAx8HgGBEevb/jYoBtATXLD1K/RP/
XXmbh7moy0mxhPA5em2LU8s22EGjN9L0VjbmqER6xWlRccZ8BmAGQVOgILK98IGD
EN7EQSf6ZzLzClBS3AxGr62suP81yuXQLytNLY9xbNRPsQ7WnpPHrZM13CCb4wqb
4G5MXyLj077RdVFZV8l7P6DQ0Bb2AYWf2egYv1iEMRun2v3bzN4DX6Oup2vRD/RC
sKd/QyqWV1U9FSTgbRKAIKb1I1tZAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAMXo
tY+WqHGVXhrStebknCJ5dd8bLQwqEqCBBLDzsjP43Q1g3yXT7fTl1zIdUNYhD/x9
y02YCDJnRXR5vRivR47TXtdXJFL8d2jSBGF7q2J4qDNdHLNsEzmWYHzNYUYqBB+5
XkiqUKgKvdbaGCsHkhlmwUS3IdtxVQGtPDOzZ3/ZRwMqlhiPayFHGCpk7aGvSHA6
rU4XYOp88sPhuqmy7zafUNNlmt2XSWaNrS/Nf1WNH1GtH92uUaLh53BSP/MB5//a
u/1tNUOn8VJWVtOHtVdmMOkSf1+H3g4JOD+nq+AD2ZTgB+KRkUQph6V0bc1H9CnW
KtvlOZ1W3/EFj1Hwouw=
-----END CERTIFICATE-----`

	rootCaCertificatePem = `-----BEGIN CERTIFICATE-----
MIIDOjCCAiICCQDUf/5116L7fTANBgkqhkiG9w0BAQsFADBfMQ8wDQYDVQQLDAZD
NGNvcmUxDDAKBgNVBAoMA1NBUDEQMA4GA1UEBwwHV2FsZG9yZjEQMA4GA1UECAwH
V2FsZG9yZjELMAkGA1UEBhMCREUxDTALBgNVBAMMBEt5bWEwHhcNMTgwNzEzMDk1
MjUxWhcNMTkwNzEzMDk1MjUxWjBfMQ8wDQYDVQQLDAZDNGNvcmUxDDAKBgNVBAoM
A1NBUDEQMA4GA1UEBwwHV2FsZG9yZjEQMA4GA1UECAwHV2FsZG9yZjELMAkGA1UE
BhMCREUxDTALBgNVBAMMBEt5bWEwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
AoIBAQD49IZuqogcaqAVSV79L7xKMI36NMy6ig+jTquecN9LRhQcalKDKxJ0BRet
bSUhftr8qcE3SaxOtPPvTLoiixjlMFaQ46ZfAx8HgGBEevb/jYoBtATXLD1K/RP/
XXmbh7moy0mxhPA5em2LU8s22EGjN9L0VjbmqER6xWlRccZ8BmAGQVOgILK98IGD
EN7EQSf6ZzLzClBS3AxGr62suP81yuXQLytNLY9xbNRPsQ7WnpPHrZM13CCb4wqb
4G5MXyLj077RdVFZV8l7P6DQ0Bb2AYWf2egYv1iEMRun2v3bzN4DX6Oup2vRD/RC
sKd/QyqWV1U9FSTgbRKAIKb1I1tZAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAMXo
tY+WqHGVXhrStebknCJ5dd8bLQwqEqCBBLDzsjP43Q1g3yXT7fTl1zIdUNYhD/x9
y02YCDJnRXR5vRivR47TXtdXJFL8d2jSBGF7q2J4qDNdHLNsEzmWYHzNYUYqBB+5
XkiqUKgKvdbaGCsHkhlmwUS3IdtxVQGtPDOzZ3/ZRwMqlhiPayFHGCpk7aGvSHA6
rU4XYOp88sPhuqmy7zafUNNlmt2XSWaNrS/Nf1WNH1GtH92uUaLh53BSP/MB5//a
u/1tNUOn8VJWVtOHtVdmMOkSf1+H3g4JOD+nq+AD2ZTgB+KRkUQph6V0bc1H9CnW
KtvlOZ1W3/EFj1Hwouw=
-----END CERTIFICATE-----`

	renewedClientCertificatePem = `-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----`
)

var (
	connectionTime = time.Now()
)

func TestController_Reconcile(t *testing.T) {

	clientKey := loadPrivateKey(t, []byte(clientKeyPem))
	clientCert := loadCertificate(t, []byte(clientCertificatePem))
	caCert := []*x509.Certificate{
		loadCertificate(t, []byte(caCertificatePem)),
		loadCertificate(t, []byte(rootCaCertificatePem)),
	}

	renewedClientCert := []byte(renewedClientCertificatePem)

	certChain := []byte(renewedClientCertificatePem + caCertificatePem)

	minimalSyncTime := 300 * time.Second

	namespacedName := types.NamespacedName{
		Name: centralConnectionName,
	}

	request := reconcile.Request{
		NamespacedName: namespacedName,
	}

	renewedCertificates := certificates.Certificates{
		ClientCRT: renewedClientCert,
		CaCRT:     []byte(caCertificatePem),
		CRTChain:  certChain,
	}

	managementInfo := connectorservice.ManagementInfo{
		ManagementURLs: connectorservice.ManagementURLs{
			RenewalURL: renewalURL,
		},
	}

	certificateCredentials := connectorservice.CertificateCredentials{
		ClientKey:        clientKey,
		ClientCert:       clientCert,
		CertificateChain: caCert,
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
			Run(setupCentralConnectionToRenew).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(checkStatus).Return(nil)

		certPreserver := &certMocks.Preserver{}
		certPreserver.On("PreserveCertificates", renewedCertificates).Return(nil)

		certProvider := &certMocks.Provider{}
		certProvider.On("GetClientCredentials").Return(clientKey, clientCert, nil)
		certProvider.On("GetCertificateChain").Return(caCert, nil)

		mutualTLSClient := &connectorMocks.EstablishedConnectionClient{}
		mutualTLSClient.On("GetManagementInfo", managementInfoURL).Return(managementInfo, nil)
		mutualTLSClient.On("RenewCertificate", renewalURL).Return(renewedCertificates, nil)

		mTLSClientProvider := &connectorMocks.EstablishedConnectionClientProvider{}
		mTLSClientProvider.On("CreateClient", certificateCredentials).Return(mutualTLSClient, nil)

		connectionController := newCentralConnectionController(client, minimalSyncTime, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.NoError(t, err)
		assertExpectations(t, &client.Mock, &certPreserver.Mock, &mutualTLSClient.Mock, &mTLSClientProvider.Mock)
	})

	t.Run("should not skip connection if RenewNow set to true", func(t *testing.T) {
		// given
		checkStatus := func(args mock.Arguments) {
			connection := args.Get(1).(*v1alpha1.CentralConnection)

			assert.NotEmpty(t, connection.Status.CertificateStatus.NotBefore)
			assert.NotEmpty(t, connection.Status.CertificateStatus.NotAfter)

			assert.Equal(t, connection.Status.SynchronizationStatus.LastSync, connection.Status.SynchronizationStatus.LastSuccess)
		}

		setupConnectionWithRenewNow := func(args mock.Arguments) {
			connection := getCentralConnectionFromArgs(args)
			setupCentralConnectionToSkip(args)

			connection.Spec.RenewNow = true
		}

		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(setupConnectionWithRenewNow).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(checkStatus).Return(nil)

		certPreserver := &certMocks.Preserver{}
		certPreserver.On("PreserveCertificates", renewedCertificates).Return(nil)

		certProvider := &certMocks.Provider{}
		certProvider.On("GetClientCredentials").Return(clientKey, clientCert, nil)
		certProvider.On("GetCertificateChain").Return(caCert, nil)

		mutualTLSClient := &connectorMocks.EstablishedConnectionClient{}
		mutualTLSClient.On("GetManagementInfo", managementInfoURL).Return(managementInfo, nil)
		mutualTLSClient.On("RenewCertificate", renewalURL).Return(renewedCertificates, nil)

		mTLSClientProvider := &connectorMocks.EstablishedConnectionClientProvider{}
		mTLSClientProvider.On("CreateClient", certificateCredentials).Return(mutualTLSClient, nil)

		connectionController := newCentralConnectionController(client, minimalSyncTime, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.NoError(t, err)
		assertExpectations(t, &client.Mock, &certPreserver.Mock, &mutualTLSClient.Mock, &mTLSClientProvider.Mock)
	})

	t.Run("should check connection and skip renewal", func(t *testing.T) {
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

		certPreserver := &certMocks.Preserver{}

		certProvider := &certMocks.Provider{}
		certProvider.On("GetClientCredentials").Return(clientKey, clientCert, nil)
		certProvider.On("GetCertificateChain").Return(caCert, nil)

		mutualTLSClient := &connectorMocks.EstablishedConnectionClient{}
		mutualTLSClient.On("GetManagementInfo", managementInfoURL).Return(managementInfo, nil)

		mTLSClientProvider := &connectorMocks.EstablishedConnectionClientProvider{}
		mTLSClientProvider.On("CreateClient", certificateCredentials).Return(mutualTLSClient, nil)

		connectionController := newCentralConnectionController(client, minimalSyncTime, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.NoError(t, err)
		assertExpectations(t, &client.Mock, &certPreserver.Mock, &mutualTLSClient.Mock, &mTLSClientProvider.Mock)
	})

	t.Run("should not take action if connection deleted", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Return(k8sErrors.NewNotFound(schema.GroupResource{}, "error"))

		certPreserver := &certMocks.Preserver{}
		certProvider := &certMocks.Provider{}
		mTLSClientProvider := &connectorMocks.EstablishedConnectionClientProvider{}

		connectionController := newCentralConnectionController(client, minimalSyncTime, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.NoError(t, err)
		assertExpectations(t, &client.Mock, &certProvider.Mock, &certPreserver.Mock)
	})

	t.Run("should skip synchronization when not enough time passed", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(setupCentralConnectionToSkip).Return(nil)

		certPreserver := &certMocks.Preserver{}
		certProvider := &certMocks.Provider{}
		mTLSClientProvider := &connectorMocks.EstablishedConnectionClientProvider{}

		connectionController := newCentralConnectionController(client, minimalSyncTime, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.NoError(t, err)
		assertExpectations(t, &client.Mock, &certProvider.Mock, &certPreserver.Mock)
	})

	t.Run("should set error status when failed to get client key and certificate", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(setupCentralConnectionInstance).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(assertErrorStatus(t)).Return(nil)

		certPreserver := &certMocks.Preserver{}
		certProvider := &certMocks.Provider{}
		certProvider.On("GetClientCredentials").Return(nil, nil, errors.New("error"))

		mTLSClientProvider := &connectorMocks.EstablishedConnectionClientProvider{}

		connectionController := newCentralConnectionController(client, minimalSyncTime, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.Error(t, err)
		assertExpectations(t, &client.Mock, &certPreserver.Mock)
	})

	t.Run("should set error status when failed to get CA certificate", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(setupCentralConnectionInstance).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(assertErrorStatus(t)).Return(nil)

		certPreserver := &certMocks.Preserver{}
		certProvider := &certMocks.Provider{}
		certProvider.On("GetClientCredentials").Return(clientKey, clientCert, nil)
		certProvider.On("GetCertificateChain").Return(nil, errors.New("error"))

		mTLSClientProvider := &connectorMocks.EstablishedConnectionClientProvider{}

		connectionController := newCentralConnectionController(client, minimalSyncTime, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.Error(t, err)
		assertExpectations(t, &client.Mock, &certPreserver.Mock)
	})

	t.Run("should set error status when failed to get management info", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(setupCentralConnectionInstance).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(assertErrorStatus(t)).Return(nil)

		certPreserver := &certMocks.Preserver{}

		certProvider := &certMocks.Provider{}
		certProvider.On("GetClientCredentials").Return(clientKey, clientCert, nil)
		certProvider.On("GetCertificateChain").Return(caCert, nil)

		mutualTLSClient := &connectorMocks.EstablishedConnectionClient{}
		mutualTLSClient.On("GetManagementInfo", managementInfoURL).
			Return(connectorservice.ManagementInfo{}, errors.New("error"))

		mTLSClientProvider := &connectorMocks.EstablishedConnectionClientProvider{}
		mTLSClientProvider.On("CreateClient", certificateCredentials).Return(mutualTLSClient, nil)

		connectionController := newCentralConnectionController(client, minimalSyncTime, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.Error(t, err)
		assertExpectations(t, &client.Mock, &certPreserver.Mock, &mutualTLSClient.Mock, &mTLSClientProvider.Mock)
	})

	t.Run("should set error status when failed to renew certificate", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(setupCentralConnectionToRenew).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(assertErrorStatus(t)).Return(nil)

		certPreserver := &certMocks.Preserver{}

		certProvider := &certMocks.Provider{}
		certProvider.On("GetClientCredentials").Return(clientKey, clientCert, nil)
		certProvider.On("GetCertificateChain").Return(caCert, nil)

		mutualTLSClient := &connectorMocks.EstablishedConnectionClient{}
		mutualTLSClient.On("GetManagementInfo", managementInfoURL).Return(managementInfo, nil)
		mutualTLSClient.On("RenewCertificate", renewalURL).Return(certificates.Certificates{}, errors.New("error"))

		mTLSClientProvider := &connectorMocks.EstablishedConnectionClientProvider{}
		mTLSClientProvider.On("CreateClient", certificateCredentials).Return(mutualTLSClient, nil)

		connectionController := newCentralConnectionController(client, minimalSyncTime, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.Error(t, err)
		assertExpectations(t, &client.Mock, &certPreserver.Mock, &mutualTLSClient.Mock, &mTLSClientProvider.Mock)
	})

	t.Run("should set error status when failed to preserve certificates", func(t *testing.T) {
		// given
		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(setupCentralConnectionToRenew).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(assertErrorStatus(t)).Return(nil)

		certPreserver := &certMocks.Preserver{}
		certPreserver.On("PreserveCertificates", renewedCertificates).Return(errors.New("error"))

		certProvider := &certMocks.Provider{}
		certProvider.On("GetClientCredentials").Return(clientKey, clientCert, nil)
		certProvider.On("GetCertificateChain").Return(caCert, nil)

		mutualTLSClient := &connectorMocks.EstablishedConnectionClient{}
		mutualTLSClient.On("GetManagementInfo", managementInfoURL).Return(managementInfo, nil)
		mutualTLSClient.On("RenewCertificate", renewalURL).Return(renewedCertificates, nil)

		mTLSClientProvider := &connectorMocks.EstablishedConnectionClientProvider{}
		mTLSClientProvider.On("CreateClient", certificateCredentials).Return(mutualTLSClient, nil)

		connectionController := newCentralConnectionController(client, minimalSyncTime, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.Error(t, err)
		assertExpectations(t, &client.Mock, &certPreserver.Mock, &mutualTLSClient.Mock, &mTLSClientProvider.Mock)
	})

	t.Run("should set error status when failed to decode pem", func(t *testing.T) {
		// given
		invalidCerts := certificates.Certificates{
			ClientCRT: []byte("invalid cert"),
		}

		client := &mocks.Client{}
		client.On("Get", context.Background(), namespacedName, mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(setupCentralConnectionToRenew).Return(nil).Twice()
		client.On("Update", context.Background(), mock.AnythingOfType("*v1alpha1.CentralConnection")).
			Run(assertErrorStatus(t)).Return(nil)

		certPreserver := &certMocks.Preserver{}
		certPreserver.On("PreserveCertificates", invalidCerts).Return(nil)

		certProvider := &certMocks.Provider{}
		certProvider.On("GetClientCredentials").Return(clientKey, clientCert, nil)
		certProvider.On("GetCertificateChain").Return(caCert, nil)

		mutualTLSClient := &connectorMocks.EstablishedConnectionClient{}
		mutualTLSClient.On("GetManagementInfo", managementInfoURL).Return(managementInfo, nil)
		mutualTLSClient.On("RenewCertificate", renewalURL).Return(invalidCerts, nil)

		mTLSClientProvider := &connectorMocks.EstablishedConnectionClientProvider{}
		mTLSClientProvider.On("CreateClient", certificateCredentials).Return(mutualTLSClient, nil)

		connectionController := newCentralConnectionController(client, minimalSyncTime, certPreserver, certProvider, mTLSClientProvider)

		// when
		_, err := connectionController.Reconcile(request)

		// then
		require.Error(t, err)
		assertExpectations(t, &client.Mock, &certPreserver.Mock, &mutualTLSClient.Mock, &mTLSClientProvider.Mock)
	})
}

func TestController_shouldRenew(t *testing.T) {

	testCases := []struct {
		renewNow        bool
		certStatus      *v1alpha1.CertificateStatus
		minimalSyncTime time.Duration
		shouldRenew     bool
	}{
		{
			certStatus: &v1alpha1.CertificateStatus{
				NotBefore: metav1.Now(),
				NotAfter:  metav1.NewTime(time.Now().Add(2000 * time.Hour)),
			},
			minimalSyncTime: 10 * time.Minute,
			shouldRenew:     false,
		},
		{
			certStatus: &v1alpha1.CertificateStatus{
				NotBefore: metav1.Now(),
				NotAfter:  metav1.Now(),
			},
			minimalSyncTime: 10 * time.Minute,
			shouldRenew:     true,
		},
		{
			certStatus: &v1alpha1.CertificateStatus{
				NotBefore: metav1.NewTime(time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)),
				NotAfter:  metav1.NewTime(time.Now().Add(3 * time.Hour)),
			},
			minimalSyncTime: 10 * time.Minute,
			shouldRenew:     true,
		},
		{
			certStatus: &v1alpha1.CertificateStatus{
				NotBefore: metav1.NewTime(time.Now()),
				NotAfter:  metav1.NewTime(time.Now().Add(30 * time.Minute)),
			},
			minimalSyncTime: 20 * time.Minute,
			shouldRenew:     true,
		},
		{
			renewNow: true,
			certStatus: &v1alpha1.CertificateStatus{
				NotBefore: metav1.Now(),
				NotAfter:  metav1.NewTime(time.Now().Add(2000 * time.Hour)),
			},
			minimalSyncTime: 10 * time.Minute,
			shouldRenew:     true,
		},
	}

	for _, testCase := range testCases {
		connection := &v1alpha1.CentralConnection{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec:       v1alpha1.CentralConnectionSpec{ManagementInfoURL: "url", RenewNow: testCase.renewNow},
			Status: v1alpha1.CentralConnectionStatus{
				CertificateStatus: testCase.certStatus,
			},
		}

		willRenew := shouldRenew(connection, testCase.minimalSyncTime)

		assert.Equal(t, testCase.shouldRenew, willRenew)
	}

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

func setupCentralConnectionToRenew(args mock.Arguments) {
	centralConnection := getCentralConnectionFromArgs(args)
	setupCentralConnectionInstance(args)

	centralConnection.Status.CertificateStatus = &v1alpha1.CertificateStatus{
		NotBefore: metav1.NewTime(time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)),
		NotAfter:  metav1.NewTime(time.Now().Add(5 * time.Minute)),
	}
}

func setupCentralConnectionToSkip(args mock.Arguments) {
	centralConnection := getCentralConnectionFromArgs(args)
	setupCentralConnectionInstance(args)

	centralConnection.Status.SynchronizationStatus = v1alpha1.SynchronizationStatus{
		LastSync:    metav1.Now(),
		LastSuccess: metav1.Now(),
	}
}

func assertExpectations(t *testing.T, mocks ...*mock.Mock) {
	for _, m := range mocks {
		m.AssertExpectations(t)
	}
}

func loadPrivateKey(t *testing.T, key []byte) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(key))
	require.NotNil(t, block)

	if privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return privateKey
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	require.NoError(t, err)

	return privateKey.(*rsa.PrivateKey)
}

func loadCertificate(t *testing.T, certificate []byte) *x509.Certificate {
	pemBlock, _ := pem.Decode(certificate)
	require.NotNil(t, pemBlock)

	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	require.NoError(t, err)

	return cert
}
