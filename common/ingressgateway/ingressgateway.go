package ingressgateway

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/net/context"
)

const (
	// This environmental variable must hold address of ingress gateway which client should access.
	// Address may be IP or FQDN. If this variable is not present creator will fallback to output of `minikube ip`.
	ServiceNameEnv = "INGRESSGATEWAY_ADDRESS"
)

// dialer is an interface which abstracts net.dialer.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type ClientCreator interface {
	// ServiceAddress returns address of ingress gateway read from environment. It does so by reading
	// service name from environmental variable and resolving it. If env variable is not set output of `minikube ip` will
	// be used as address to support local testing.
	ServiceAddress() (string, error)

	// Client returns http.Client which dials ingress gateway service address instead of the one provided in request
	// URL. For details how address is resolved see ServiceAddressFromEnv.
	Client() (*http.Client, error)
}

// clientCreator abstracts interactions with outside world for easy testing.
type clientCreator struct {
	ingressFQDN   func() string
	lookupHost    func(string) ([]string, error)
	getMinikubeIP func() (string, error)
	dialer        Dialer
}

// FromEnv returns ClientCreator ready for production use.
func FromEnv() ClientCreator {
	return &clientCreator{
		ingressFQDN:   defaultIngressFQDN,
		lookupHost:    net.LookupHost,
		getMinikubeIP: getMinikubeIP,
		dialer: &net.Dialer{
			Timeout: 30 * time.Second,
		},
	}
}

func (cc *clientCreator) ServiceAddress() (string, error) {
	serviceUrl := cc.ingressFQDN()

	if serviceUrl != "" {
		addresses, err := cc.lookupHost(serviceUrl)
		if err != nil {
			return "", fmt.Errorf("unable to lookup host %s: %s", serviceUrl, err)
		}
		return addresses[0], nil
	}

	minikubeIP, err := cc.getMinikubeIP()
	if err != nil {
		return "", fmt.Errorf("cannot get minikube IP: %s", err)
	}
	return minikubeIP, nil
}

func (cc *clientCreator) Client() (*http.Client, error) {
	ingressGatewayAddr, err := cc.ServiceAddress()
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			addr = ingressGatewayAddr + ":443"
			return cc.dialer.DialContext(ctx, network, addr)
		},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10,
	}

	return client, nil
}

func getMinikubeIP() (string, error) {
	mipCmd := exec.Command("minikube", "ip")
	mipOut, err := mipCmd.Output()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(mipOut), "\n"), nil
}

func defaultIngressFQDN() string {
	return os.Getenv(ServiceNameEnv)
}
