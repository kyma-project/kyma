package http_client

import (
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type HttpClientConfig struct {
	Address   string `envconfig:"default=storage.kyma.local"`
	Secure    bool   `envconfig:"default=true"`
	VerifySSL bool   `envconfig:"default=true"`
}

type HttpClient struct {
	cfg      HttpClientConfig
	endpoint string
	client   *http.Client
}

func NewHttpClient(cfg HttpClientConfig) (*HttpClient, error) {
	client := &http.Client{}
	if !cfg.VerifySSL {
		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore invalid SSL certificates
		}
		client.Transport = transCfg
	}

	protocol := "http"
	if cfg.Secure {
		protocol = protocol + "s"
	}
	endpoint := fmt.Sprintf("%s://%s", protocol, cfg.Address)

	return &HttpClient{
		cfg:      cfg,
		endpoint: endpoint,
		client:   client,
	}, nil
}

func (h *HttpClient) getFile(baseURL, name string) ([]byte, error) {
	path := h.preparePath(baseURL, name)
	if path == "" {
		return nil, nil
	}

	data, err := h.fetch(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	return data, nil
}

func (h *HttpClient) preparePath(baseURL, name string) string {
	if baseURL == "" || name == "" {
		return ""
	}

	splitBaseURL := strings.Split(baseURL, "/")
	if len(splitBaseURL) < 3 {
		return ""
	}

	bucketName := splitBaseURL[len(splitBaseURL)-2]
	assetName := splitBaseURL[len(splitBaseURL)-1]

	return fmt.Sprintf("%s/%s/%s/%s", h.endpoint, bucketName, assetName, name)
}

func (h *HttpClient) fetch(url string) ([]byte, error) {
	if url == "" {
		return nil, nil
	}

	resp, err := h.client.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "while requesting file from URL %s", url)
	}
	defer func() {
		err = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Invalid status code while downloading file from URL %s: %d. Expected: %d", url, resp.StatusCode, http.StatusOK)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading response body while downloading file from URL %s", url)
	}

	return body, nil
}
