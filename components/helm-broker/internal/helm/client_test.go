package helm_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/helm"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"k8s.io/helm/pkg/proto/hapi/chart"
	hapi_release5 "k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/proto/hapi/services"
)

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
	fakeTiller := &fakeTillerSvc{}
	fakeTiller.SetUp(t)

	hClient := helm.NewClient(helm.Config{
		TillerHost:              fakeTiller.Host,
		TillerConnectionTimeout: time.Second,
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

	s.grpcSvc = grpc.NewServer()
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
