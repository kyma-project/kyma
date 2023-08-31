package webhook

import (
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	SystemNamespace string `envconfig:"default=kyma-system"`
	ServiceName     string `envconfig:"default=serverless-webhook"`
	SecretName      string `envconfig:"default=serverless-webhook"`
	Port            int    `envconfig:"default=8443"`
	LogConfigPath   string `envconfig:"default=/appdata/log_config.yaml"`
	ConfigPath      string `envconfig:"default=/appdata/config.yaml"`
}

func ReadValidationConfigV1Alpha2OrDie() *serverlessv1alpha2.ValidationConfig {
	validationCfg := &serverlessv1alpha2.ValidationConfig{}
	if err := envconfig.InitWithPrefix(validationCfg, "WEBHOOK_VALIDATION"); err != nil {
		panic(errors.Wrap(err, "while reading env defaulting variables"))
	}
	return validationCfg
}
