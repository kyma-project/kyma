package graphql

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/machinebox/graphql"
)

const (
	timeout = 30 * time.Second
)

type ClientConstructor func(certificate *tls.Certificate, graphqlEndpoint string, enableLogging bool, insecureConfigFetch bool) (Client, error)

//go:generate mockery -name=Client
type Client interface {
	Do(req *graphql.Request, res interface{}) error
}

type client struct {
	gqlClient *graphql.Client
	logs      []string
	logging   bool
}

func New(certificate *tls.Certificate, graphqlEndpoint string, enableLogging, insecureConfigFetch bool) (Client, error) {
	var certificates []tls.Certificate
	if certificate != nil {
		certificates = []tls.Certificate{*certificate}
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureConfigFetch,
				Certificates:       certificates,
			},
		},
	}

	gqlClient := graphql.NewClient(graphqlEndpoint, graphql.WithHTTPClient(httpClient))

	client := &client{
		gqlClient: gqlClient,
		logging:   enableLogging,
		logs:      []string{},
	}

	client.gqlClient.Log = client.addLog

	return client, nil
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
