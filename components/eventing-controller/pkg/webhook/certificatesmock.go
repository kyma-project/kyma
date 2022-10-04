package webhook

import (
	"context"
	"crypto/x509"
	"net"

	"golang.org/x/xerrors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
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
	ErrFakeGet         = xerrors.New("fake get error")
	ErrFakeCreate      = xerrors.New("fake create error")
	ErrFakeUpdate      = xerrors.New("fake update error")
	ErrFakeCertInvalid = xerrors.New("fake cert invalid error")
	ErrFakeNotfound    = apiErrors.NewNotFound(schema.GroupResource{Group: "test", Resource: "test"}, "fake resource not found")
)

func (r *MockGetFailedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return ErrFakeGet
}

func (r *MockCreateFailedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return ErrFakeNotfound
}

func (r *MockCreateFailedClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return ErrFakeCreate
}

func (r *MockUpdateFailedClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return ErrFakeUpdate
}

func (r *MockUpdateFailedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return nil
}

type MockCertificateHandler struct {
	ICertificateHandler
}

func (r *MockCertificateHandler) isValidCertificate(cert, key []byte) (bool, error) {
	return false, ErrFakeCertInvalid
}
