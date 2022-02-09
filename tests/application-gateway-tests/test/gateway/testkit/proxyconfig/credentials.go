package proxyconfig

import (
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/proxyconfig"
)

type ConfigBuilder struct {
	proxyConfig proxyconfig.ProxyDestinationConfig
}

func NewConfigBuilder(targetURL string) *ConfigBuilder {
	return &ConfigBuilder{
		proxyConfig: proxyconfig.ProxyDestinationConfig{
			TargetURL:     targetURL,
			Configuration: proxyconfig.Configuration{},
		}}
}

func (cb *ConfigBuilder) WithOAuth(clientId, clientSecret, tokenURL string, reqParams authorization.RequestParameters) *ConfigBuilder {
	cb.proxyConfig.Configuration.Credentials = proxyconfig.OauthConfig{
		ClientId:          clientId,
		ClientSecret:      clientSecret,
		TokenURL:          tokenURL,
		RequestParameters: reqParams,
	}
	return cb
}

func (cb *ConfigBuilder) WithBasicAuth(username, password string) *ConfigBuilder {
	cb.proxyConfig.Configuration.Credentials = proxyconfig.BasicAuthConfig{
		Username: username,
		Password: password,
	}
	return cb
}

func (cb *ConfigBuilder) WithCertificate(cert, key []byte) *ConfigBuilder {
	cb.proxyConfig.Configuration.Credentials = proxyconfig.CertificateConfig{
		Certificate: cert,
		PrivateKey:  key,
	}
	return cb
}

func (cb *ConfigBuilder) WithRequestParameters(params authorization.RequestParameters) *ConfigBuilder {
	cb.proxyConfig.Configuration.RequestParameters = &params
	return cb
}

func (cb *ConfigBuilder) WithCSRF(tokenURL string) *ConfigBuilder {
	cb.proxyConfig.Configuration.CSRFConfig = &proxyconfig.CSRFConfig{TokenURL: tokenURL}
	return cb
}

func (cb *ConfigBuilder) ToConfig() proxyconfig.ProxyDestinationConfig {
	return cb.proxyConfig
}
