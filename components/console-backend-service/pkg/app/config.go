package app

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/authn"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/authz"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/experimental"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/tracing"
	"time"
)

type Config struct {
	Host                 string        `envconfig:"default=127.0.0.1"`
	Port                 int           `envconfig:"default=3000"`
	AllowedOrigins       []string      `envconfig:"optional"`
	Verbose              bool          `envconfig:"default=false"`
	KubeconfigPath       string        `envconfig:"optional"`
	InformerResyncPeriod time.Duration `envconfig:"default=10m"`
	ServerTimeout        time.Duration `envconfig:"default=10s"`
	Application          application.Config
	AssetStore           assetstore.Config
	OIDC                 authn.OIDCConfig
	SARCacheConfig       authz.SARCacheConfig
	FeatureToggles       experimental.FeatureToggles
	Tracing              tracing.Config
}
