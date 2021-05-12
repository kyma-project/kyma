package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler/extension"

	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/authn"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/authz"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/experimental"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/origin"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/signal"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/vrischmann/envconfig"
	authenticatorpkg "k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type config struct {
	Host                 string        `envconfig:"default=127.0.0.1"`
	Port                 int           `envconfig:"default=3000"`
	AllowedOrigins       []string      `envconfig:"optional"`
	Verbose              bool          `envconfig:"default=false"`
	KubeconfigPath       string        `envconfig:"optional"`
	SystemNamespaces     []string      `envconfig:"default=istio-system;knative-eventing;kube-public;kube-system;kyma-installer;kyma-integration;kyma-system;natss;compass-system"`
	InformerResyncPeriod time.Duration `envconfig:"default=10m"`
	ServerTimeout        time.Duration `envconfig:"default=10s"`
	Burst                int           `envconfig:"default=2"`
	Application          application.Config
	Rafter               rafter.Config
	Serverless           serverless.Config
	OIDC                 authn.OIDCConfig
	SARCacheConfig       authz.SARCacheConfig
	FeatureToggles       experimental.FeatureToggles
	DebugDomain          string `envconfig:"optional"`
	EventSubscription    bool   `envconfig:"optional"`
}

func main() {
	cfg, _, err := loadConfig("APP")
	exitOnError(err, "Error while loading app config")
	parseFlags(cfg)

	k8sConfig, err := newRestClientConfig(cfg.KubeconfigPath, cfg.Burst)
	exitOnError(err, "Error while initializing REST client config")

	kubeClient, err := kubernetes.NewForConfig(k8sConfig)
	exitOnError(err, "Failed to instantiate Kubernetes client")

	resolvers, err := domain.New(kubeClient, k8sConfig, cfg.Application, cfg.Rafter, cfg.Serverless, cfg.InformerResyncPeriod, cfg.FeatureToggles, cfg.SystemNamespaces, cfg.EventSubscription)
	exitOnError(err, "Error while creating resolvers")

	gqlCfg := gqlschema.Config{Resolvers: resolvers}

	var authenticator authenticatorpkg.Request

	authenticator, err = authn.NewOIDCAuthenticator(&cfg.OIDC)
	exitOnError(err, "Error while creating OIDC authenticator")
	sarClient := kubeClient.AuthorizationV1().SubjectAccessReviews()
	authorizer, err := authz.NewAuthorizer(sarClient, cfg.SARCacheConfig)
	exitOnError(err, "Failed to create authorizer")

	gqlCfg.Directives.HasAccess = authz.NewRBACDirective(authorizer, kubeClient.Discovery())

	stopCh := signal.SetupChannel()
	resolvers.WaitForCacheSync(stopCh)

	executableSchema := gqlschema.NewExecutableSchema(gqlCfg)

	runServer(stopCh, cfg, executableSchema, authenticator)
}

func loadConfig(prefix string) (config, bool, error) {
	cfg := config{}

	err := envconfig.InitWithPrefix(&cfg, prefix)
	if err != nil {
		return cfg, false, err
	}

	if cfg.DebugDomain != "" {
		//debug mode
		cfg.Rafter.Address = strings.Join([]string{"https://storage", cfg.DebugDomain}, ".")
		cfg.Verbose = true
		cfg.Rafter.VerifySSL = true
		cfg.Application.Gateway.IntegrationNamespace = "kyma-integration"
		cfg.OIDC.IssuerURL = strings.Join([]string{"https://dex", cfg.DebugDomain}, ".")
		cfg.OIDC.ClientID = "kyma-client"
	}

	developmentMode := cfg.KubeconfigPath != ""
	return cfg, developmentMode, nil
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

func newRestClientConfig(kubeconfigPath string, burst int) (*restclient.Config, error) {
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

	config.Burst = burst
	config.UserAgent = "console-backend-service"
	return config, nil
}

func runServer(stop <-chan struct{}, cfg config, schema graphql.ExecutableSchema, authenticator authenticatorpkg.Request) {
	var allowedOrigins []string
	if len(cfg.AllowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	} else {
		allowedOrigins = cfg.AllowedOrigins
	}

	router := mux.NewRouter()

	if authenticator != nil {
		router.Use(authn.AuthMiddleware(authenticator))
	}

	router.Handle("/", playground.Handler("Kyma GQL", "/graphql"))
	graphQLHandler := newGqlServer(schema, allowedOrigins)
	router.Handle("/graphql", graphQLHandler)
	serverHandler := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			"POST", "GET", "OPTIONS",
		},
		AllowCredentials:   true,
		AllowedHeaders:     []string{"*"},
		OptionsPassthrough: false,
	}).Handler(router)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
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

func newGqlServer(es graphql.ExecutableSchema, allowedOrigins []string) *handler.Server {
	srv := handler.New(es)

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: origin.CheckFn(allowedOrigins),
		},
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New(1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})

	return srv
}
