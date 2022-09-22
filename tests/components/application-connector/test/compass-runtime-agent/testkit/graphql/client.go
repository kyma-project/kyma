package graphql

import (
	"context"
	"crypto/tls"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/third_party/machinebox/graphql"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	timeout = 30 * time.Second
)

type ClientConstructor func(certificate *tls.Certificate, graphqlEndpoint string, enableLogging bool, insecureConfigFetch bool) (Client, error)

//go:generate mockery --name=Client
type Client interface {
	Do(req *graphql.Request, res interface{}) error
}

type client struct {
	gqlClient *graphql.Client
	logs      []string
	logging   bool
}

func NewGraphQLClient(graphqlEndpoint string, enableLogging bool, insecureSkipVerify bool) Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
		},
	}

	gqlClient := graphql.NewClient(graphqlEndpoint, graphql.WithHTTPClient(httpClient))

	client := &client{
		gqlClient: gqlClient,
		logging:   enableLogging,
		logs:      []string{},
	}

	client.gqlClient.Log = client.addLog

	return client
}

func (c *client) Do(req *graphql.Request, res interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c.clearLogs()
	err := c.gqlClient.Run(ctx, req, res)
	if err != nil {
		for _, l := range c.logs {
			if l != "" {
				logrus.Info(l)
			}
		}
	}
	return err
}

func (c *client) addLog(log string) {
	if !c.logging {
		return
	}

	c.logs = append(c.logs, log)
}

func (c *client) clearLogs() {
	c.logs = []string{}
}
