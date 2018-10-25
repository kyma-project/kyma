package graphql

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	timeout = 3 * time.Second
)

type Client struct {
	gqlClient      *graphql.Client
	token          string
	endpoint       string
	clusterContext func(ctx context.Context, network, addr string) (net.Conn, error)
}

func New() (*Client, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, errors.Wrap(err, "while loading config")
	}

	clusterContext := dialContext(config.LocalClusterHosts)
	token, err := authenticate(config.IdProviderConfig, clusterContext)
	if err != nil {
		return nil, err
	}

	httpClient := newAuthorizedClient(token, clusterContext)
	gqlClient := graphql.NewClient(config.GraphQlEndpoint, graphql.WithHTTPClient(httpClient))
	//gqlClient.Log = func(s string) { log.Println(s) }

	return &Client{
		gqlClient:      gqlClient,
		token:          token,
		clusterContext: clusterContext,
		endpoint:       config.GraphQlEndpoint,
	}, nil
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

func (c *Client) Subscribe(req *Request) *Subscription {
	if req == nil {
		return errorSubscription(errors.New("invalid request"))
	}

	url := strings.Replace(c.endpoint, "http://", "ws://", -1)
	url = strings.Replace(url, "https://", "wss://", -1)

	connection, err := newWebsocket(url, c.token, req.req.Header, c.clusterContext)
	if err != nil {
		return errorSubscription(err)
	}

	err = connection.Handshake()
	if err != nil {
		return errorSubscription(err)
	}

	js, err := req.Json()
	if err != nil {
		return errorSubscription(errors.Wrapf(err, "while converting request to json"))
	}

	err = connection.Start(js)
	if err != nil {
		return errorSubscription(err)
	}

	return newSubscription(connection)
}

func authenticate(config idProviderConfig, dialContext func(ctx context.Context, network, addr string) (net.Conn, error)) (string, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext:     dialContext,
		},
	}

	idTokenProvider := newDexIdTokenProvider(httpClient, config)
	token, err := idTokenProvider.fetchIdToken()
	return token, err
}

func dialContext(localClusterHosts []string) func(ctx context.Context, network, addr string) (net.Conn, error) {
	ingressGatewayControllerAddr, err := net.LookupHost(ingressGatewayControllerServiceURL)
	if err != nil {

		glog.Warningf("Unable to resolve host '%s' (if you are running this test from outside of Kyma ignore this log). Root cause: %v", ingressGatewayControllerServiceURL, err)

		if minikubeIp := tryToGetMinikubeIp(); minikubeIp != "" {
			ingressGatewayControllerAddr = []string{minikubeIp}
		}
	}

	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if len(ingressGatewayControllerAddr) > 0 {
			glog.Infof("Resolving: %s", addr)
			for _, host := range localClusterHosts {
				if strings.HasPrefix(addr, host) {
					addr = strings.Replace(addr, host, ingressGatewayControllerAddr[0], 1)
					break
				}
			}
			glog.Infof("Resolved: %s", addr)
		}

		dialer := net.Dialer{}
		return dialer.DialContext(ctx, network, addr)
	}
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
