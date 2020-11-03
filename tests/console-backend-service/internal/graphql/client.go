package graphql

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
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
	token        string
	cachedTokens map[User]string
	endpoint     string
	logs         []string
	logging      bool
	Config       config
}

func New() (*Client, error) {
	config, err := loadConfig(AdminUser) // by default create client capable of performing all operations on all resources
	if err != nil {
		return nil, errors.Wrap(err, "while loading config")
	}

	dexURL, parseErr := url.Parse(config.IdProviderConfig.DexConfig.BaseUrl)
	if parseErr != nil {
		return nil, errors.Wrap(parseErr, "Could not parse Dex URL from config")
	}

	_, err = net.DialTimeout("tcp", dexURL.Host+dexURL.Path+":80", 60*time.Second)
	if err != nil {
		return nil, errors.Wrap(err, "Could not resolve Dex URL within a reasonable time")
	}

	token, err := authenticate(config.IdProviderConfig)
	if err != nil {
		return nil, err
	}

	httpClient := newAuthorizedClient(token)
	gqlClient := graphql.NewClient(config.GraphQLEndpoint, graphql.WithHTTPClient(httpClient))

	client := &Client{
		gqlClient: gqlClient,
		token:     token,
		endpoint:  config.GraphQLEndpoint,
		logging:   true,
		logs:      []string{},
		Config:    config,
	}

	client.gqlClient.Log = client.addLog
	client.cachedTokens = make(map[User]string)
	client.cachedTokens[AdminUser] = token

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
	err := c.gqlClient.Run(ctx, req.req, res)
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

	connection, err := newWebsocket(url, c.token, req.req.Header)
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

func (c *Client) ChangeUser(user User) error {
	var token string

	config, err := loadConfig(user)
	if err != nil {
		return errors.Wrap(err, "while loading config")
	}

	if user != NoUser {
		if c.cachedTokens[user] != "" {
			token = c.cachedTokens[user]
		} else {
			token, err = authenticate(config.IdProviderConfig)
			if err != nil {
				return err
			}
			c.cachedTokens[user] = token
		}
	} else {
		token = ""
	}

	httpClient := newAuthorizedClient(token)
	gqlClient := graphql.NewClient(config.GraphQLEndpoint, graphql.WithHTTPClient(httpClient))

	c.token = token
	c.gqlClient = gqlClient

	c.gqlClient.Log = c.addLog

	return nil
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

func authenticate(config idProviderConfig) (string, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	idTokenProvider := newDexIdTokenProvider(httpClient, config)
	token, err := idTokenProvider.fetchIdToken()
	return token, err
}
