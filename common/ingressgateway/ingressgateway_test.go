package ingressgateway

import (
	"errors"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

const (
	gwFQDN          = "my-ingressgateway.my-ns.example.com"
	serviceAddress  = "12.13.14.15"
	minikubeAddress = "15.14.13.12"
)

func TestClientFromEnv(t *testing.T) {
	cases := []struct {
		name                   string
		fqdnProvided           bool
		serviceAddressResolved bool
		expectSuccess          bool
		minikubeAvailable      bool
	}{
		{
			name:                   "Ingress gateway FQDN set and found",
			fqdnProvided:           true,
			serviceAddressResolved: true,
			minikubeAvailable:      false,
			expectSuccess:          true,
		},
		{
			name:                   "Ingress gateway FQDN set but not found",
			fqdnProvided:           true,
			serviceAddressResolved: false,
			minikubeAvailable:      false,
			expectSuccess:          false,
		},
		{
			name:                   "Ingress gateway FQDN not set and minikube is used",
			fqdnProvided:           false,
			serviceAddressResolved: false,
			minikubeAvailable:      true,
			expectSuccess:          true,
		},
		{
			name:                   "Ingress gateway FQDN not set and minikube is not available",
			fqdnProvided:           false,
			serviceAddressResolved: false,
			minikubeAvailable:      false,
			expectSuccess:          false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mock := &dialerMock{}
			creator := &clientCreator{

				ingressFQDN: func() string {
					if c.fqdnProvided {
						return gwFQDN
					}
					return ""
				},

				lookupHost: func(s string) (strings []string, e error) {
					assert.Equal(t, gwFQDN, s)
					if c.serviceAddressResolved {
						return []string{serviceAddress}, nil
					}
					return nil, errors.New("address not resolved")
				},

				getMinikubeIP: func() (s string, e error) {
					if c.minikubeAvailable {
						return minikubeAddress, nil
					}
					return "", errors.New("minikube not available")
				},

				dialer: mock,
			}

			client, err := creator.Client()

			if !c.expectSuccess {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				expectedServiceAddress := serviceAddress
				if c.minikubeAvailable {
					expectedServiceAddress = minikubeAddress
				}
				_, _ = client.Get("http://example.com")
				assert.Equal(t, expectedServiceAddress+":443", mock.lastDialed)
				httpTransport, ok := client.Transport.(*http.Transport)
				assert.True(t, ok, "client.Transport is not *http.Transport")
				assert.True(t, httpTransport.TLSClientConfig.InsecureSkipVerify)
			}
		})
	}
}

func TestDefaultIngressFQDN(t *testing.T) {
	err := os.Setenv(ServiceNameEnv, gwFQDN)
	assert.Nil(t, err)

	actual := defaultIngressFQDN()
	assert.Equal(t, gwFQDN, actual)
}

func TestLookupHostWithIP(t *testing.T) {
	lookupLocalhost, err := net.LookupHost("127.0.0.1")
	assert.Nil(t, err)
	assert.Equal(t, "127.0.0.1", lookupLocalhost[0])
}

type dialerMock struct {
	lastDialed string
}

type connMock struct{}

func (connMock) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (connMock) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (connMock) Close() error {
	return nil
}

func (connMock) LocalAddr() net.Addr {
	return nil
}

func (connMock) RemoteAddr() net.Addr {
	return nil
}

func (connMock) SetDeadline(t time.Time) error {
	return nil
}

func (connMock) SetReadDeadline(t time.Time) error {
	return nil
}

func (connMock) SetWriteDeadline(t time.Time) error {
	return nil
}

func (m *dialerMock) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	m.lastDialed = address
	return connMock{}, nil
}
