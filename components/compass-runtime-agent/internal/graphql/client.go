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

type Client interface {
	Do(req *graphql.Request, res interface{}) error
	DisableLogging()
}

type client struct {
	gqlClient *graphql.Client
	logs      []string
	logging   bool
}

func New(certificate tls.Certificate, graphqlEndpoint string, enableLogging bool) (Client, error) {
	httpClient := &http.Client{
		// TODO: enable when the certificates will be ready
		//Transport: &http.Transport{
		//	TLSClientConfig: &tls.Config{
		//		Certificates: []tls.Certificate{certificate},
		//	},
		//},
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

func (c *client) DisableLogging() {
	c.logging = false
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
