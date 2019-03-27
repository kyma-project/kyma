package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/authn"

	authenticatorpkg "k8s.io/apiserver/pkg/authentication/authenticator"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/experimental"
	"k8s.io/client-go/kubernetes"

	"github.com/99designs/gqlgen/graphql"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/signal"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/vrischmann/envconfig"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/websocket"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/authz"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/content"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/origin"
)

type config struct {
	Host                 string   `envconfig:"default=127.0.0.1"`
	Port                 int      `envconfig:"default=3000"`
	AllowedOrigins       []string `envconfig:"optional"`
	Verbose              bool     `envconfig:"default=false"`
	KubeconfigPath       string   `envconfig:"optional"`
	Content              content.Config
	InformerResyncPeriod time.Duration `envconfig:"default=10m"`
	ServerTimeout        time.Duration `envconfig:"default=10s"`
	Application          application.Config
	OIDC                 authn.OIDCConfig
	SARCacheConfig       authz.SARCacheConfig
	FeatureToggles       experimental.FeatureToggles
}

func main() {
	cfg, err := loadConfig("APP")
	exitOnError(err, "Error while loading app config")
	parseFlags(cfg)

	k8sConfig, err := newRestClientConfig(cfg.KubeconfigPath)
	exitOnError(err, "Error while initializing REST client config")

	resolvers, err := domain.New(k8sConfig, cfg.Content, cfg.Application, cfg.InformerResyncPeriod, cfg.FeatureToggles)
	exitOnError(err, "Error while creating resolvers")

	kubeClient, err := kubernetes.NewForConfig(k8sConfig)
	exitOnError(err, "Failed to instantiate Kubernetes client")

	authenticator, err := authn.NewOIDCAuthenticator(&cfg.OIDC)
	exitOnError(err, "Error while creating OIDC authenticator")
	sarClient := kubeClient.AuthorizationV1beta1().SubjectAccessReviews()
	authorizer, err := authz.NewAuthorizer(sarClient, cfg.SARCacheConfig)
	exitOnError(err, "Failed to create authorizer")

	stopCh := signal.SetupChannel()
	resolvers.WaitForCacheSync(stopCh)

	c := gqlschema.Config{Resolvers: resolvers}
	c.Directives.HasAccess = authz.NewRBACDirective(authorizer, kubeClient.Discovery())
	executableSchema := gqlschema.NewExecutableSchema(c)
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	runServer(stopCh, addr, cfg.AllowedOrigins, executableSchema, authenticator)
}

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		glog.Fatal(wrappedError)
	}
}

func parseFlags(cfg config) {
	if cfg.Verbose {
		err := flag.Set("stderrthreshold", "INFO")
		if err != nil {
			glog.Error(errors.Wrap(err, "while parsing flags"))
		}
	}
	flag.Parse()
}

func newRestClientConfig(kubeconfigPath string) (*restclient.Config, error) {
	var config *restclient.Config
	var err error
	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		config, err = restclient.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}
	return config, nil
}

func runServer(stop <-chan struct{}, addr string, allowedOrigins []string, schema graphql.ExecutableSchema, authenticator authenticatorpkg.Request) {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	router := mux.NewRouter()

	router.Use(authn.AuthMiddleware(authenticator))

	router.HandleFunc("/", handler.Playground("Dataloader", "/graphql"))
	router.HandleFunc("/graphql", handler.GraphQL(schema,
		handler.WebsocketUpgrader(websocket.Upgrader{
			CheckOrigin: origin.CheckFn(allowedOrigins),
		}),
	))

	serverHandler := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			"POST", "GET", "OPTIONS",
		},
		AllowCredentials:   true,
		AllowedHeaders:     []string{"*"},
		OptionsPassthrough: false,
	}).Handler(router)

	srv := &http.Server{Addr: addr, Handler: serverHandler}

	glog.Infof("Listening on %s", addr)

	go func() {
		<-stop
		// Interrupt signal received - shut down the server
		if err := srv.Shutdown(context.Background()); err != nil {
			glog.Errorf("HTTP server Shutdown: %v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		glog.Errorf("HTTP server ListenAndServe: %v", err)
	}
}
