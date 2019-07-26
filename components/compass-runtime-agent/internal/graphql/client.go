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
	timeout = 10 * time.Second
)

type Client struct {
	gqlClient *graphql.Client
	endpoint  string
	logs      []string
	logging   bool
}

func New(certificate tls.Certificate, graphqlEndpoint string) (*Client, error) {
	//config, err := loadConfig(AdminUser) // by default create client capable of performing all operations on all resources
	//if err != nil {
	//	return nil, errors.Wrap(err, "while loading config")
	//}
	//
	//token, err := authenticate(config.IdProviderConfig)
	//if err != nil {
	//	return nil, err
	//}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{certificate},
			},
		},
	}

	gqlClient := graphql.NewClient(graphqlEndpoint, graphql.WithHTTPClient(httpClient))

	client := &Client{
		gqlClient: gqlClient,
		//token:     token,
		//endpoint:  config.GraphQLEndpoint,
		logging: true,
		logs:    []string{},
		//Config:    config,
	}

	client.gqlClient.Log = client.addLog

	return client, nil
}

func (c *Client) DoQuery(q string, res interface{}) error {
	req := graphql.NewRequest(q)
	return c.Do(req, res)
}

func (c *Client) Do(req *graphql.Request, res interface{}) error {
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

func (c *Client) DisableLogging() {
	c.logging = false
}

func (c *Client) addLog(log string) {
	if !c.logging {
		return
	}

	c.logs = append(c.logs, log)
}

func (c *Client) clearLogs() {
	c.logs = []string{}
}
