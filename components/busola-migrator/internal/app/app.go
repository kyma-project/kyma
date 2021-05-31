package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/config"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/jwt"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/kubernetes"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/model"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/renderer"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/uaa"

	jwxjwt "github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

//go:generate mockery --name=UAAClient --output=automock --outpkg=automock --case=underscore
type UAAClient interface {
	GetOpenIDConfiguration() (uaa.OpenIDConfiguration, error)
	GetAuthorizationEndpointWithParams(authzEndpoint, oauthState string) (string, error)
	GetToken(tokenEndpoint string, authCode string) (map[string]interface{}, error)
}

//go:generate mockery --name=HTMLRenderer --output=automock --outpkg=automock --case=underscore
type HTMLRenderer interface {
	RenderTemplate(w io.Writer, templateName renderer.TemplateName, data interface{}) error
}

//go:generate mockery --name=JWTService --output=automock --outpkg=automock --case=underscore
type JWTService interface {
	ParseAndVerify(jwtSrc, jwksURI string) (jwxjwt.Token, error)
	GetUser(token jwxjwt.Token) (model.User, error)
}

//go:generate mockery --name=K8sClient --output=automock --outpkg=automock --case=underscore
type K8sClient interface {
	EnsureUserPermissions(user model.User) error
}

type App struct {
	UAAEnabled bool

	busolaURL    string
	fsRoot       http.FileSystem
	fsAssets     http.FileSystem
	uaaClient    UAAClient
	uaaOIDConfig uaa.OpenIDConfiguration
	k8sClient    K8sClient
	jwtService   JWTService

	htmlRenderer HTMLRenderer
}

func New(cfg config.Config, busolaURL string, kubeConfig *rest.Config) (App, error) {
	wd, _ := os.Getwd()
	dir := path.Join(wd, "static")
	if cfg.StaticFilesDIR != "" {
		dir = cfg.StaticFilesDIR
	}

	// UAA Migrator functionality is disabled, no need to initialize all clients
	if !cfg.UAA.Enabled {
		return App{
			UAAEnabled:   cfg.UAA.Enabled,
			busolaURL:    busolaURL,
			fsRoot:       http.Dir(dir),
			fsAssets:     http.Dir(path.Join(dir, "assets")),
			htmlRenderer: renderer.New(path.Join(dir, "templates")),
		}, nil
	}

	uaaCfg := cfg.UAA
	uaaCfg.RedirectURI = fmt.Sprintf("https://dex.%s/callback", cfg.Domain)
	uaaClient := uaa.NewClient(uaaCfg)
	uaaOIDConfig, err := uaaClient.GetOpenIDConfiguration()
	if err != nil {
		return App{}, errors.Wrap(err, "while getting OpenID configuration")
	}

	k8sClient, err := kubernetes.New(kubeConfig)
	if err != nil {
		return App{}, errors.Wrap(err, "while creating Kubernetes client")
	}

	return App{
		UAAEnabled:   cfg.UAA.Enabled,
		busolaURL:    busolaURL,
		fsRoot:       http.Dir(dir),
		fsAssets:     http.Dir(path.Join(dir, "assets")),
		uaaClient:    uaaClient,
		uaaOIDConfig: uaaOIDConfig,
		k8sClient:    k8sClient,
		jwtService:   jwt.NewService(),
		htmlRenderer: renderer.New(path.Join(dir, "templates")),
	}, nil
}
