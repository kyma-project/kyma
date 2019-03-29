package helm_test

import (
	"context"
	"encoding/pem"
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

const (
	keyFile  string = "/tmp/dummy.key"
	certFile string = "/tmp/dummy.crt"
)

func TestMain(m *testing.M) {
	err := createDummyKeyPair(keyFile, certFile)
	if err != nil {
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestClientInstallSuccess(t *testing.T) {
	// given
	fakeTiller := &fakeTillerSvc{}
	fakeTiller.SetUp(t)

	cVals := internal.ChartValues{
		"test-param": "value-test",
	}

	hClient := helm.NewClient(helm.Config{
		TillerHost:              fakeTiller.Host,
		TillerConnectionTimeout: time.Second,
		TillerTLSCrt:            certFile,
		TillerTLSKey:            keyFile,
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

	// Clean-up
	fakeTiller.TearDown(t)
}

func TestClientDeleteSuccess(t *testing.T) {
	// given
	//err := createDummyKeyPair(keyFile, certFile)
	//assert.NoError(t, err)

	fakeTiller := &fakeTillerSvc{}
	fakeTiller.SetUp(t)

	hClient := helm.NewClient(helm.Config{
		TillerHost:              fakeTiller.Host,
		TillerConnectionTimeout: time.Second,
		TillerTLSCrt:            certFile,
		TillerTLSKey:            keyFile,
	}, spy.NewLogDummy())

	// when
	err := hClient.Delete("r-name")

	// then
	assert.NoError(t, err)

	assert.NotNil(t, fakeTiller.GotDelReleaseReq)
	assert.Equal(t, fakeTiller.GotDelReleaseReq.Name, "r-name")

	// Clean-up
	fakeTiller.TearDown(t)
}

type fakeTillerSvc struct {
	services.ReleaseServiceServer
	GotInstReleaseReq *services.InstallReleaseRequest
	GotDelReleaseReq  *services.UninstallReleaseRequest

	grpcSvc      *grpc.Server
	Host         string
	serverErr    error
	serverClosed chan struct{}
}

func (s *fakeTillerSvc) SetUp(t *testing.T) {
	s.serverClosed = make(chan struct{}, 1)
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	s.Host = lis.Addr().String()
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	require.NoError(t, err)

	s.grpcSvc = grpc.NewServer(grpc.Creds(creds))
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

func createDummyKeyPair(keyFile, certFile string) error {
	cert, priv, err := certs.NewCertificatePair("Kyma", "SAP", 12*time.Hour)
	if err != nil {
		return err
	}
	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	certOut.Close()

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	pem.Encode(keyOut, certs.PemBlockForKey(priv))
	keyOut.Close()

	return nil
}
