package graphql

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/golang/glog"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	ingressGatewayControllerServiceURL = "istio-ingressgateway.istio-system.svc.cluster.local"
	timeout                            = 3 * time.Second
)

type Client struct {
	gqlClient *graphql.Client
}

func New() (*Client, error) {

	config := loadConfig()

	dexHttpClient := newHttpClient(config.localClusterHosts, nil)

	idTokenProvider := newDexIdTokenProvider(dexHttpClient, config.idProviderConfig)

	gqlHttpClient := newHttpClient(config.localClusterHosts, idTokenProvider)

	gqlClient := graphql.NewClient(config.graphQlEndpoint, graphql.WithHTTPClient(gqlHttpClient))

	return &Client{gqlClient}, nil
}

func (c *Client) DoQuery(q string, res interface{}) error {
	req := NewRequest(q)
	return c.Do(req, res)
}

func (c *Client) Do(req *Request, res interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return c.gqlClient.Run(ctx, req.req, res)
}

type transportWithIdTokensSupport struct {
	http.Transport
	idTokenProvider idTokenProvider
	currentIdToken  string
}

func (t *transportWithIdTokensSupport) RoundTrip(req *http.Request) (*http.Response, error) {

	if t.idTokenProvider != nil {

		if t.currentIdToken == "" {
			if idToken, err := t.idTokenProvider.fetchIdToken(); err != nil {
				return nil, errors.Wrap(err, "Request canceled.")
			} else {
				t.currentIdToken = idToken
			}
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.currentIdToken))
	}
	return t.Transport.RoundTrip(req)
}

func newHttpClient(localClusterHosts []string, idTokenProvider idTokenProvider) *http.Client {

	tr := &transportWithIdTokensSupport{
		Transport: http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		idTokenProvider: idTokenProvider,
	}

	ingressGatewayControllerAddr, err := net.LookupHost(ingressGatewayControllerServiceURL)
	if err != nil {

		glog.Warningf("Unable to resolve host '%s' (if you are running this test from outside of Kyma ignore this log). Root cause: %v", ingressGatewayControllerServiceURL, err)

		if minikubeIp := tryToGetMinikubeIp(); minikubeIp != "" {
			ingressGatewayControllerAddr = []string{minikubeIp}
		}
	}

	if len(ingressGatewayControllerAddr) > 0 {

		glog.Infof("Ingress controller address: '%s'", ingressGatewayControllerAddr[0])

		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {

			glog.Infof("Resolving: %s", addr)
			for _, host := range localClusterHosts {
				if strings.HasPrefix(addr, host) {
					addr = strings.Replace(addr, host, ingressGatewayControllerAddr[0], 1)
					break
				}
			}
			glog.Infof("Resolved: %s", addr)

			dialer := net.Dialer{}
			return dialer.DialContext(ctx, network, addr)
		}
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10,
	}

	return client
}

func tryToGetMinikubeIp() string {
	mipCmd := exec.Command("minikube", "ip")
	if mipOut, err := mipCmd.Output(); err != nil {
		glog.Warningf("Error while getting minikube IP (ignore this message if you are running this test inside Kyma). Root cause: %s", err)
		return ""
	} else {
		return strings.Trim(string(mipOut), "\n")
	}
}
