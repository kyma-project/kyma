package jwt

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type envConfig struct {
	Domain           string        `envconfig:"TEST_DOMAIN,default=kyma.local"`
	UserEmail        string        `envconfig:"TEST_USER_EMAIL"`
	UserPassword     string        `envconfig:"TEST_USER_PASSWORD"`
	ClientTimeout    time.Duration `envconfig:"TEST_CLIENT_TIMEOUT,default=10s"` //Don't forget the unit!
	RetryMaxAttempts uint          `envconfig:"TEST_RETRY_MAX_ATTEMPTS,default=5"`
	RetryDelay       uint          `envconfig:"TEST_RETRY_DELAY,default=5"`
}

//Config JWT configuration structure
type Config struct {
	IdProviderConfig idProviderConfig
	EnvConfig        envConfig
}

type idProviderConfig struct {
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

//LoadConfig Generate test config from envs
func LoadConfig() (Config, error) {
	env := envConfig{}
	err := envconfig.Init(&env)
	if err != nil {
		return Config{}, errors.Wrap(err, "while loading environment variables")
	}

	config := Config{EnvConfig: env}

	config.IdProviderConfig = idProviderConfig{
		DexConfig: dexConfig{
			BaseUrl:           fmt.Sprintf("https://dex.%s", env.Domain),
			AuthorizeEndpoint: fmt.Sprintf("https://dex.%s/auth", env.Domain),
			TokenEndpoint:     fmt.Sprintf("https://dex.%s/token", env.Domain),
		},
		ClientConfig: clientConfig{
			ID:             "kyma-client",
			RedirectUri:    "http://127.0.0.1:5555/callback",
			TimeoutSeconds: env.ClientTimeout,
		},
		RetryConfig: retryConfig{
			MaxAttempts: env.RetryMaxAttempts,
			Delay:       time.Duration(env.RetryDelay) * time.Second,
		},
	}

	config.IdProviderConfig.UserCredentials = userCredentials{
		Username: env.UserEmail,
		Password: env.UserPassword,
	}

	return config, nil
}
