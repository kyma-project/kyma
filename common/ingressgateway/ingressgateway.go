package ingressgateway

import (
	"crypto/tls"
	"fmt"
	"golang.org/x/net/context"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	// This environmental variable must hold FQDN of ingress gateway which client should access
	ServiceNameEnv = "INGRESSGATEWAY_FQDN"
)

// Dialer is an interface which abstracts net.Dialer.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// clientCreator abstracts interactions with outside world for easy testing.
type clientCreator struct {
	Getenv func(string) string
	LookupHost func(string) ([]string, error)
	GetMinikubeIP func() (string, error)
	Dialer Dialer
}

// Default returns clientCreator ready for production use.
func Default() *clientCreator {
	return &clientCreator{
		Getenv: os.Getenv,
		LookupHost: net.LookupHost,
		GetMinikubeIP: getMinikubeIP,
		Dialer: &net.Dialer{
			Timeout: 30 * time.Second,
		},
	}
}

// ServiceAddressFromEnv returns address of ingress gateway read from environment. It does so by reading
// service name from environmental variable and resolving it. If env variable is not set output of `minikube ip` will
// be used as address to support local testing.
func (cc *clientCreator) ServiceAddressFromEnv() (string, error) {
	serviceUrl := cc.Getenv(ServiceNameEnv)

	if serviceUrl != "" {
		addresses, err := cc.LookupHost(serviceUrl)
		if err != nil {
			return "", fmt.Errorf("unable to lookup host %s: %s", serviceUrl, err)
		}
		return addresses[0], nil
	}

	minikubeIP, err := cc.GetMinikubeIP()
	if err != nil {
		return "", fmt.Errorf("cannot get minikube IP: %s", err)
	}
	return minikubeIP, nil
}

// ClientFromEnv returns http.Client which dials ingress gateway service address instead of the one provided in request
// URL. For details how address is resolved see ServiceAddressFromEnv.
func (cc *clientCreator) ClientFromEnv() (*http.Client, error) {
	ingressGatewayAddr, err := cc.ServiceAddressFromEnv()
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			addr = ingressGatewayAddr + ":443"
			return cc.Dialer.DialContext(ctx, network, addr)
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
