package webhook

import (
	"context"
	"crypto/x509"
	"net"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/cert"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MockClientGoCert struct{}

func (r *MockClientGoCert) generateSelfSignedCertKey(host string, alternateIPs []net.IP, alternateDNS []string) ([]byte, []byte, error) {
	return nil, nil, xerrors.New("fake self signed certificate generation failed")
}
func (r *MockClientGoCert) parseCertsPEM(pemCerts []byte) ([]*x509.Certificate, error) {
	return cert.ParseCertsPEM(pemCerts)
}
func (r *MockClientGoCert) newPoolFromBytes(pemBlock []byte) (*x509.CertPool, error) {
	return cert.NewPoolFromBytes(pemBlock)
}

type MockGetFailedClient struct {
	client.Client
}

type MockCreateFailedClient struct {
	client.Client
}

type MockUpdateFailedClient struct {
	client.Client
}

var (
	fakeGetErr         = xerrors.New("fake get error")
	fakeCreateErr      = xerrors.New("fake create error")
	fakeUpdateErr      = xerrors.New("fake update error")
	fakeCertInvalidErr = xerrors.New("fake cert invalid error")
	fakeNotfoundErr    = apiErrors.NewNotFound(schema.GroupResource{"bla", "bla"}, "fake resource not found")
)

func (r *MockGetFailedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return fakeGetErr
}

func (r *MockCreateFailedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return fakeNotfoundErr
}

func (r *MockCreateFailedClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return fakeCreateErr
}

func (r *MockUpdateFailedClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return fakeUpdateErr
}

func (r *MockUpdateFailedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	obj = &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      "name",
			Namespace: "namespace",
		},
	}
	return nil
}

type MockCertificateHandler struct {
	ICertificateHandler
}

func (r *MockCertificateHandler) isValidCertificate(cert, key []byte) (bool, error) {
	return false, fakeCertInvalidErr
}
