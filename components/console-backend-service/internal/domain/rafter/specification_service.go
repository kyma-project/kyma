package rafter

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"

	"crypto/tls"
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/spec"
)

//go:generate mockery -name=specificationSvc -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=specificationSvc -case=underscore -output disabled -outpkg disabled
type specificationSvc interface {
	AsyncAPI(baseURL, name string) (*spec.AsyncAPISpec, error)
}

type specificationService struct {
	cfg      Config
	endpoint string
	client   *http.Client
}

func newSpecificationService(cfg Config) (*specificationService, error) {
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

	return &specificationService{
		cfg:      cfg,
		endpoint: endpoint,
		client:   client,
	}, nil
}

func (s *specificationService) AsyncAPI(baseURL, name string) (*spec.AsyncAPISpec, error) {
	data, err := s.readData(baseURL, name)
	if err != nil || len(data) == 0 {
		return nil, err
	}

	asyncApiSpec := new(spec.AsyncAPISpec)
	err = asyncApiSpec.Decode(data)
	if err != nil {
		return nil, err
	}

	return asyncApiSpec, nil
}

func (s *specificationService) readData(baseURL, name string) ([]byte, error) {
	path := s.preparePath(baseURL, name)
	if path == "" {
		return nil, nil
	}

	data, err := s.fetch(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	return data, nil
}

func (s *specificationService) preparePath(baseURL, name string) string {
	if baseURL == "" || name == "" {
		return ""
	}

	splitBaseURL := strings.Split(baseURL, "/")
	if len(splitBaseURL) < 3 {
		return ""
	}

	bucketName := splitBaseURL[len(splitBaseURL)-2]
	assetName := splitBaseURL[len(splitBaseURL)-1]

	return fmt.Sprintf("%s/%s/%s/%s", s.endpoint, bucketName, assetName, name)
}

func (s *specificationService) fetch(url string) ([]byte, error) {
	if url == "" {
		return nil, nil
	}

	resp, err := s.client.Get(url)
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
