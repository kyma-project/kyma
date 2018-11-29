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
	serviceNameEnv = "INGRESSGATEWAY_FQDN"
)

func ServiceAddress() (string, error) {
	serviceUrl := os.Getenv(serviceNameEnv)

	if serviceUrl != "" {
		addresses, err := net.LookupHost(serviceUrl)
		if err != nil {
			return "", fmt.Errorf("unable to lookup host %s: %s", serviceUrl, err)
		}
		return addresses[0], nil
	}

	minikubeIP, err := getMinikubeIP()
	if err != nil {
		return "", fmt.Errorf("cannot get minikube IP: %s", err)
	}
	return minikubeIP, nil
}

func Client() (*http.Client, error) {
	ingressGatewayAddr, err := ServiceAddress()
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			addr = ingressGatewayAddr + ":443"
			return dialer.DialContext(ctx, network, addr)
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
