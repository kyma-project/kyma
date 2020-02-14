package proxy

import (
	"github.com/kyma-project/kyma/components/application-gateway/internal/csrf"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
)

//go:generate mockery -name=TargetConfigProvider

// TargetConfigProvider allows to fetch ProxyDestinationConfig for specific service
type TargetConfigProvider interface {
	GetDestinationConfig(id string) (ProxyDestinationConfig, apperrors.AppError)
}

// AuthType determines the secret structure
type AuthType string

const (
	Undefined   AuthType = ""
	NoAuth      AuthType = "noauth"
	Oauth       AuthType = "oauth"
	Basic       AuthType = "basicauth"
	Certificate AuthType = "certificate"
)

// ProxyDestinationConfig is Proxy configuration for specific target
type ProxyDestinationConfig struct {
	Destination Destination `json:"destination"`
	Credentials Credentials `json:"credentials"`
}

type Destination struct {
	URL               string                           `json:"url"`
	RequestParameters *authorization.RequestParameters `json:"requestParameters"`
	CSRFConfig        *csrf.CSRFConfig                 `json:"csrfConfig"`
}

type Credentials interface {
	ToCredentials() *authorization.Credentials
}

type NoAuthConfig struct{}

func (oc NoAuthConfig) ToCredentials() *authorization.Credentials {
	return nil
}

type OauthConfig struct {
	ClientId          string                          `json:"clientId"`
	ClientSecret      string                          `json:"clientSecret"`
	TokenURL          string                          `json:"tokenUrl"`
	RequestParameters authorization.RequestParameters `json:"requestParameters,omitempty"`
}

func (oc OauthConfig) ToCredentials() *authorization.Credentials {
	return &authorization.Credentials{
		OAuth: &authorization.OAuth{
			URL:               oc.TokenURL,
			ClientID:          oc.ClientId,
			ClientSecret:      oc.ClientSecret,
			RequestParameters: &oc.RequestParameters,
		},
	}
}

type BasicAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (bc BasicAuthConfig) ToCredentials() *authorization.Credentials {
	return &authorization.Credentials{
		BasicAuth: &authorization.BasicAuth{
			Username: bc.Username,
			Password: bc.Password,
		},
	}
}

type CertificateConfig struct {
	Certificate []byte `json:"certificate"`
	PrivateKey  []byte `json:"privateKey"`
}

func (cc CertificateConfig) ToCredentials() *authorization.Credentials {
	return &authorization.Credentials{
		CertificateGen: &authorization.CertificateGen{
			PrivateKey:  cc.PrivateKey,
			Certificate: cc.Certificate,
		},
	}
}
