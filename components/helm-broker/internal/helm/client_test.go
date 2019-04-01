package helm_test

import (
	"context"
	"encoding/pem"
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/SpectoLabs/hoverfly/core/certs"
	"github.com/ghodss/yaml"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/helm"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/helm/pkg/proto/hapi/chart"
	hapi_release5 "k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/proto/hapi/services"
)

var (
	keyFile  string
	certFile string
)

func TestMain(m *testing.M) {
	setupCerts()
	code := m.Run()
	cleanupCerts()
	os.Exit(code)
}

func setupCerts() {
	certOut, err := ioutil.TempFile("/tmp/", "certFile")
	if err != nil {
		log.Fatal(err)
	}
	keyOut, err := ioutil.TempFile("/tmp/", "keyFile")
	if err != nil {
		log.Fatal(err)
	}

	err = createDummyKeyPair(keyOut, certOut)
	keyOut.Close()
	certOut.Close()
	if err != nil {
		log.Fatal(err)
	}
	keyFile = keyOut.Name()
	certFile = certOut.Name()

}
func cleanupCerts() {
	os.Remove(keyFile)
	os.Remove(certFile)
}

func TestClientInstallSuccess(t *testing.T) {
	// given
	fakeTiller := &fakeTillerSvc{
		certFile: certFile,
		keyFile:  keyFile,
	}
	fakeTiller.SetUp(t)

	cVals := internal.ChartValues{
		"test-param": "value-test",
	}

	hClient := helm.NewClient(helm.Config{
		TillerHost:              fakeTiller.Host,
		TillerConnectionTimeout: time.Second,
		TillerTLSCrt:            certFile,
		TillerTLSKey:            keyFile,
		TillerTLSInsecure:       true,
	}, spy.NewLogDummy())

	// when
	_, err := hClient.Install(fixChart(), cVals, "r-name", "ns-name")

	// then
	assert.NoError(t, err)

	require.NotNil(t, fakeTiller.GotInstReleaseReq)
	assert.True(t, fakeTiller.GotInstReleaseReq.Wait)
	assert.Equal(t, fakeTiller.GotInstReleaseReq.Timeout, int64(time.Hour.Seconds()))
	assert.False(t, fakeTiller.GotInstReleaseReq.DryRun)
	assert.False(t, fakeTiller.GotInstReleaseReq.ReuseName)
	assert.False(t, fakeTiller.GotInstReleaseReq.DisableHooks)
	assert.Equal(t, fakeTiller.GotInstReleaseReq.Name, "r-name")
	assert.Equal(t, fakeTiller.GotInstReleaseReq.Namespace, "ns-name")
	assert.Equal(t, fakeTiller.GotInstReleaseReq.Chart, fixChart())

	b, err := yaml.Marshal(cVals)
	require.NoError(t, err)
	assert.Equal(t, fakeTiller.GotInstReleaseReq.Values, &chart.Config{Raw: string(b)})

	// clean-up
	fakeTiller.TearDown(t)
}

func TestClientDeleteSuccess(t *testing.T) {
	fakeTiller := &fakeTillerSvc{
		certFile: certFile,
		keyFile:  keyFile,
	}
	fakeTiller.SetUp(t)
	hClient := helm.NewClient(helm.Config{
		TillerHost:              fakeTiller.Host,
		TillerConnectionTimeout: time.Second,
		TillerTLSCrt:            certFile,
		TillerTLSKey:            keyFile,
		TillerTLSInsecure:       true,
	}, spy.NewLogDummy())

	// when
	err := hClient.Delete("r-name")

	// then
	assert.NoError(t, err)

	assert.NotNil(t, fakeTiller.GotDelReleaseReq)
	assert.Equal(t, fakeTiller.GotDelReleaseReq.Name, "r-name")

	// clean-up
	fakeTiller.TearDown(t)
}

type fakeTillerSvc struct {
	services.ReleaseServiceServer
	GotInstReleaseReq *services.InstallReleaseRequest
	GotDelReleaseReq  *services.UninstallReleaseRequest

	grpcSvc  *grpc.Server
	Host     string
	keyFile  string
	certFile string

	serverErr    error
	serverClosed chan struct{}
}

func (s *fakeTillerSvc) SetUp(t *testing.T) {
	s.serverClosed = make(chan struct{}, 1)
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	s.Host = lis.Addr().String()
	credentials, err := credentials.NewServerTLSFromFile(s.certFile, s.keyFile)
	require.NoError(t, err)

	s.grpcSvc = grpc.NewServer(grpc.Creds(credentials))
	services.RegisterReleaseServiceServer(s.grpcSvc, s)

	go func() {
		s.serverErr = s.grpcSvc.Serve(lis)
		close(s.serverClosed)
	}()
}

func (s *fakeTillerSvc) TearDown(t *testing.T) {
	s.grpcSvc.GracefulStop()

	select {
	case <-s.serverClosed:
	case <-time.After(time.Second):
		t.Errorf("Timeout [%v] occured when wainting to server shudown. ", time.Second)
	}
}

func (s *fakeTillerSvc) InstallRelease(ctx context.Context, instReleaseReq *services.InstallReleaseRequest) (*services.InstallReleaseResponse, error) {
	s.GotInstReleaseReq = instReleaseReq
	return &services.InstallReleaseResponse{
		Release: &hapi_release5.Release{
			Name: "Fake-Test-Release",
		},
	}, nil
}

func (s *fakeTillerSvc) UninstallRelease(ctx context.Context, delReleaseReq *services.UninstallReleaseRequest) (*services.UninstallReleaseResponse, error) {
	s.GotDelReleaseReq = delReleaseReq
	return &services.UninstallReleaseResponse{
		Release: &hapi_release5.Release{
			Name: "Fake-Test-Release",
		},
	}, nil
}

func fixChart() *chart.Chart {
	return &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    string("Fix-chart"),
			Version: "1.0.0",
		},
	}
}

func createDummyKeyPair(keyOut, certOut *os.File) error {
	cert, priv, err := certs.NewCertificatePair("Kyma", "SAP", 12*time.Hour)
	if err != nil {
		return err
	}

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	if err != nil {
		return err
	}

	err = pem.Encode(keyOut, certs.PemBlockForKey(priv))
	if err != nil {
		return err
	}

	return nil
}
