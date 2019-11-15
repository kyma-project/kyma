package authentication

import (
	"fmt"
	"time"
)

type EnvConfig struct {
	Domain        string
	UserEmail     string
	UserPassword  string
	ClientTimeout time.Duration
}

type IdProviderConfig struct {
	DexConfig       dexConfig
	ClientConfig    clientConfig
	UserCredentials userCredentials
	RetryConfig     retryConfig
}

type dexConfig struct {
	BaseUrl           string
	AuthorizeEndpoint string
	TokenEndpoint     string
}

type clientConfig struct {
	ID             string
	RedirectUri    string
	TimeoutSeconds time.Duration
}

type retryConfig struct {
	MaxAttempts uint
	Delay       time.Duration
}

type userCredentials struct {
	Username string
	Password string
}

func BuildIdProviderConfig(envConfig EnvConfig) IdProviderConfig {
	return IdProviderConfig{
		DexConfig: dexConfig{
			BaseUrl:           fmt.Sprintf("https://dex.%s", envConfig.Domain),
			AuthorizeEndpoint: fmt.Sprintf("https://dex.%s/auth", envConfig.Domain),
			TokenEndpoint:     fmt.Sprintf("https://dex.%s/token", envConfig.Domain),
		},
		ClientConfig: clientConfig{
			ID:             "kyma-client",
			RedirectUri:    "http://127.0.0.1:5555/callback",
			TimeoutSeconds: envConfig.ClientTimeout,
		},
		RetryConfig: retryConfig{
			MaxAttempts: 4,
			Delay:       3 * time.Second,
		},
		UserCredentials: userCredentials{
			Username: envConfig.UserEmail,
			Password: envConfig.UserPassword,
		},
	}
}
