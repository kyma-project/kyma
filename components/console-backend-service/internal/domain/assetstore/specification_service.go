package assetstore

import (
	"io/ioutil"
	"net/http"

	"crypto/tls"
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/specification"
)

//go:generate mockery -name=specificationSvc -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=specificationSvc -case=underscore -output disabled -outpkg disabled
type specificationSvc interface {
	AsyncApi(baseURL, name string) (*specification.AsyncApiSpec, error)
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

func (s *specificationService) AsyncApi(baseURL, name string) (*specification.AsyncApiSpec, error) {
	data, err := s.readData(baseURL, name)
	if err != nil || len(data) == 0 {
		return nil, err
	}

	asyncApiSpec := new(specification.AsyncApiSpec)
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

	splitedBaseURL := strings.Split(baseURL, "/")
	if len(splitedBaseURL) < 3 {
		return ""
	}

	bucketName := splitedBaseURL[len(splitedBaseURL)-2]
	assetName := splitedBaseURL[len(splitedBaseURL)-1]

	return fmt.Sprintf("%s/%s/%s/%s", s.endpoint, bucketName, assetName, name)
}

func (s *specificationService) fetch(path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = resp.Body.Close()
	}()
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
