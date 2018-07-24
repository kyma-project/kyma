package graphql

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/machinebox/graphql"
)

const (
	endpointEnvName = "ENDPOINT"
	defaultEndpoint = "http://127.0.0.1:3000/graphql"
	timeout         = 3 * time.Second
)

type Client struct {
	gqlClient *graphql.Client
}

func New() (*Client, error) {
	endpoint := os.Getenv(endpointEnvName)
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	gqlClient := graphql.NewClient(endpoint, graphql.WithHTTPClient(httpClient))
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
