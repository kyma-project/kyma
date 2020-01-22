package fileclient

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type Config struct {
	Address   string `envconfig:"default=storage.kyma.local"`
	Secure    bool   `envconfig:"default=true"`
	VerifySSL bool   `envconfig:"default=true"`
}

type FileClient struct {
	cfg      Config
	endpoint string
	client   *http.Client
}

func New(cfg Config) (*FileClient, error) {
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

	return &FileClient{
		cfg:      cfg,
		endpoint: endpoint,
		client:   client,
	}, nil
}

func (f *FileClient) Get(filePath string) ([]byte, error) {
	path := f.preparePath(filePath)
	if path == "" {
		return nil, nil
	}

	data, err := f.fetch(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	return data, nil
}

func (f *FileClient) preparePath(path string) string {
	if path == "" {
		return ""
	}

	splitBaseURL := strings.Split(path, "/")
	if len(splitBaseURL) < 3 {
		return ""
	}

	bucketName := splitBaseURL[len(splitBaseURL)-3]
	assetName := splitBaseURL[len(splitBaseURL)-2]
	fileName := splitBaseURL[len(splitBaseURL)-1]

	return fmt.Sprintf("%s/%s/%s/%s", f.endpoint, bucketName, assetName, fileName)
}

func (f *FileClient) fetch(url string) ([]byte, error) {
	if url == "" {
		return nil, nil
	}

	resp, err := f.client.Get(url)
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
