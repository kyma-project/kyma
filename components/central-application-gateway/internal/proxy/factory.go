package proxy

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/model"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"
)

type pathExtractorFunc func(*url.URL) (model.APIIdentifier, *url.URL, *url.URL, apperrors.AppError)
type gatewayURLExtractorFunc func(*url.URL) (*url.URL, apperrors.AppError)

// New creates proxy for handling user's services calls
func New(
	serviceDefService metadata.ServiceDefinitionService,
	authorizationStrategyFactory authorization.StrategyFactory,
	csrfTokenStrategyFactory csrf.TokenStrategyFactory,
	config Config) http.Handler {

	pathExtractor := func(u *url.URL) (model.APIIdentifier, *url.URL, *url.URL, apperrors.AppError) {
		path := u.EscapedPath()

		trimmed := strings.Trim(path, "/")
		split := strings.Split(trimmed, "/")

		if len(split) < 2 || split[0] == path {
			return model.APIIdentifier{}, nil, nil, apperrors.WrongInput("path must contain Application and Service name")
		}

		apiIdentifier := model.APIIdentifier{
			Application: split[0],
			Service:     split[1],
		}

		targetAPIPath := strings.Join(split[2:], "/")

		gwURL := *u
		gwURL.Path = "/" + strings.Join(split[:2], "/")

		// TODO: make sure
		targetURL, _ := url.Parse(targetAPIPath) // it shouldn't be possible for this to fail?

		return apiIdentifier, targetURL, &gwURL, nil
	}

	apiExtractor := apiExtractor{
		serviceDefService: serviceDefService,
	}

	return &proxy{
		cache:                        NewCache(config.ProxyCacheTTL),
		proxyTimeout:                 config.ProxyTimeout,
		authorizationStrategyFactory: authorizationStrategyFactory,
		csrfTokenStrategyFactory:     csrfTokenStrategyFactory,
		extractPathFunc:              pathExtractor,
		apiExtractor:                 apiExtractor,
	}
}

func NewForCompass(
	serviceDefService metadata.ServiceDefinitionService,
	authorizationStrategyFactory authorization.StrategyFactory,
	csrfTokenStrategyFactory csrf.TokenStrategyFactory,
	config Config) http.Handler {

	extractFunc := func(u *url.URL) (model.APIIdentifier, *url.URL, *url.URL, apperrors.AppError) {
		path := u.EscapedPath()
		trimmed := strings.Trim(path, "/")
		split := strings.Split(trimmed, "/")

		if len(split) < 3 || split[0] == path {
			return model.APIIdentifier{}, nil, nil, apperrors.WrongInput("path must contain Application, Service and Entry name")
		}

		apiIdentifier := model.APIIdentifier{
			Application: split[0],
			Service:     split[1],
			Entry:       split[2],
		}

		targetAPIPath := strings.Join(split[3:], "/")

		gwURL := *u
		gwURL.Path = "/" + strings.Join(split[:3], "/")

		targetURL, _ := url.Parse(targetAPIPath)

		return apiIdentifier, targetURL, &gwURL, nil
	}

	apiExtractor := compassAPIExtractor{
		serviceDefService: serviceDefService,
	}

	return &proxy{
		cache:                        NewCache(config.ProxyCacheTTL),
		proxyTimeout:                 config.ProxyTimeout,
		authorizationStrategyFactory: authorizationStrategyFactory,
		csrfTokenStrategyFactory:     csrfTokenStrategyFactory,
		extractPathFunc:              extractFunc,
		apiExtractor:                 apiExtractor,
	}
}

type apiExtractor struct {
	serviceDefService metadata.ServiceDefinitionService
}

func (ae apiExtractor) Get(identifier model.APIIdentifier) (*model.API, apperrors.AppError) {
	return ae.serviceDefService.GetAPIByServiceName(identifier.Application, identifier.Service)
}

type compassAPIExtractor struct {
	serviceDefService metadata.ServiceDefinitionService
}

func (ae compassAPIExtractor) Get(identifier model.APIIdentifier) (*model.API, apperrors.AppError) {
	return ae.serviceDefService.GetAPIByEntryName(identifier.Application, identifier.Service, identifier.Entry)
}
