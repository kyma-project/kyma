package graphql

import (
	"context"
	"fmt"
	"github.com/avast/retry-go"
	"net/http"
	"strings"
	"time"

	"log"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	timeout = 10 * time.Second
)

type Client struct {
	gqlClient    *graphql.Client
	endpoint     string
	logs         []string
	logging      bool
}

func New(port int) (*Client, error) {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	endpoint := fmt.Sprintf("http://127.0.0.1:%v/graphql", port)
	gqlClient := graphql.NewClient(endpoint, graphql.WithHTTPClient(httpClient))

	client := &Client{
		gqlClient: gqlClient,
		endpoint:  endpoint,
		logging:   true,
		logs:      []string{},
	}

	client.gqlClient.Log = client.addLog

	return client, nil
}

func (c *Client) DoQuery(q string, res interface{}) error {
	req := NewRequest(q)
	return c.Do(req, res)
}

func (c *Client) Do(req *Request, res interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c.clearLogs()
	err := retry.Do(func() error {
		return c.gqlClient.Run(ctx, req.req, res)
	})
	if err != nil {
		for _, l := range c.logs {
			if l != "" {
				log.Println(l)
			}
		}
	}
	return err
}

func (c *Client) Subscribe(req *Request) *Subscription {
	if req == nil {
		return errorSubscription(errors.New("invalid request"))
	}

	url := strings.Replace(c.endpoint, "http://", "ws://", -1)
	url = strings.Replace(url, "https://", "wss://", -1)

	connection, err := newWebsocket(url, req.req.Header)
	if err != nil {
		return errorSubscription(err)
	}

	err = connection.Handshake()
	if err != nil {
		return errorSubscription(err)
	}

	js, err := req.JSON()
	if err != nil {
		return errorSubscription(errors.Wrapf(err, "while converting request to json"))
	}

	err = connection.Start(js)
	if err != nil {
		return errorSubscription(err)
	}

	return newSubscription(connection)
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
