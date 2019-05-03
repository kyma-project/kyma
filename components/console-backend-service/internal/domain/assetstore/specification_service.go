package assetstore

import (
	"io/ioutil"
	"net/http"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/specification"
)

//go:generate mockery -name=specificationSvc -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=specificationSvc -case=underscore -output disabled -outpkg disabled
type specificationSvc interface {
	AsyncApi(path string) (*specification.AsyncApiSpec, error)
}

type specificationService struct{}

func newSpecificationService() (*specificationService, error) {
	return &specificationService{}, nil
}

func (s *specificationService) AsyncApi(path string) (*specification.AsyncApiSpec, error) {
	if path == "" {
		return nil, nil
	}

	data, err := fetch(path)
	if err != nil {
		return nil, err
	}

	asyncApiSpec := new(specification.AsyncApiSpec)
	err = asyncApiSpec.Decode(data)
	if err != nil {
		return nil, err
	}

	return asyncApiSpec, nil
}

func fetch(path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}

	resp, err := http.Get(path)
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
