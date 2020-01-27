package fileclient

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type Config struct {
	VerifySSL bool `envconfig:"default=true"`
}

type FileClient struct {
	cfg    Config
	client *http.Client
}

func New(cfg Config) (*FileClient, error) {
	client := &http.Client{}
	if !cfg.VerifySSL {
		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore invalid SSL certificates
		}
		client.Transport = transCfg
	}

	return &FileClient{
		cfg:    cfg,
		client: client,
	}, nil
}

func (f *FileClient) Get(path string, headers http.Header) ([]byte, error) {
	data, err := f.fetch(path, headers)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	return data, nil
}

func (f *FileClient) fetch(url string, headers http.Header) ([]byte, error) {
	if url == "" {
		return nil, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating request: %v", req)
	}
	if headers != nil {
		req.Header = headers
	}

	resp, err := f.client.Do(req)
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
