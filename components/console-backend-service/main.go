package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/authn"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/authz"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/experimental"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/origin"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/signal"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/vrischmann/envconfig"
	authenticatorpkg "k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type config struct {
	Host                 string        `envconfig:"default=127.0.0.1"`
	Port                 int           `envconfig:"default=3000"`
	AllowedOrigins       []string      `envconfig:"optional"`
	Verbose              bool          `envconfig:"default=false"`
	KubeconfigPath       string        `envconfig:"optional"`
	InformerResyncPeriod time.Duration `envconfig:"default=10m"`
	ServerTimeout        time.Duration `envconfig:"default=10s"`
	AuthEnabled          bool          `envconfig:"default=true"`
	Application          application.Config
	AssetStore           assetstore.Config
	OIDC                 authn.OIDCConfig
	SARCacheConfig       authz.SARCacheConfig
	FeatureToggles       experimental.FeatureToggles
	Tracing              tracing.Config
}

func main() {
	cfg, err := loadConfig("APP")
	exitOnError(err, "Error while loading app config")
	parseFlags(cfg)

	k8sConfig, err := newRestClientConfig(cfg.KubeconfigPath)
	exitOnError(err, "Error while initializing REST client config")

	stopCh := signal.SetupChannel()

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	listener, err := net.Listen("tcp", addr)
	exitOnError(err, "Error while binding listener")
	glog.Infof("Listening on %s", addr)

	err = run(listener, stopCh, cfg, k8sConfig)
	if err != nil {
		glog.Fatal(err)
	}
}

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
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

func run(listener net.Listener, stopCh <-chan struct{}, cfg config, k8sConfig *restclient.Config) error {
	resolvers, err := domain.New(k8sConfig, cfg.Application, cfg.AssetStore, cfg.InformerResyncPeriod, cfg.FeatureToggles)
	if err != nil {
		return errors.Wrap(err, "Error while creating resolvers")
	}

	kubeClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return errors.Wrap(err, "Failed to instantiate Kubernetes client")
	}

	gqlCfg := gqlschema.Config{Resolvers: resolvers}
	var authenticator authenticatorpkg.Request
	if cfg.AuthEnabled {
		authenticator, err = authn.NewOIDCAuthenticator(&cfg.OIDC)
		exitOnError(err, "Error while creating OIDC authenticator")
		sarClient := kubeClient.AuthorizationV1beta1().SubjectAccessReviews()
		authorizer, err := authz.NewAuthorizer(sarClient, cfg.SARCacheConfig)
		exitOnError(err, "Failed to create authorizer")

		gqlCfg.Directives.HasAccess = authz.NewRBACDirective(authorizer, kubeClient.Discovery())
	}

	resolvers.WaitForCacheSync(stopCh)
	executableSchema := gqlschema.NewExecutableSchema(gqlCfg)
	runServer(listener, stopCh, cfg, executableSchema, authenticator)
	return nil
}

func runServer(listener net.Listener, stop <-chan struct{}, cfg config, schema graphql.ExecutableSchema, authenticator authenticatorpkg.Request) {
	setupTracing(cfg.Tracing, cfg.Port)
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
	router.HandleFunc("/", handler.Playground("Dataloader", "/graphql"))
	graphQLHandler := handler.GraphQL(schema,
		handler.WebsocketUpgrader(websocket.Upgrader{
			CheckOrigin: origin.CheckFn(allowedOrigins),
		}),
		handler.Tracer(tracing.New()),
	)
	router.HandleFunc("/graphql", tracing.NewWithParentSpan(cfg.Tracing.ServiceSpanName, graphQLHandler))
	serverHandler := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			"POST", "GET", "OPTIONS",
		},
		AllowCredentials:   true,
		AllowedHeaders:     []string{"*"},
		OptionsPassthrough: false,
	}).Handler(router)

	srv := &http.Server{Handler: serverHandler}

	go func() {
		<-stop
		// Interrupt signal received - shut down the server
		if err := srv.Shutdown(context.Background()); err != nil {
			glog.Errorf("HTTP server Shutdown: %v", err)
		}
	}()

	if err := srv.Serve(listener); err != http.ErrServerClosed {
		glog.Errorf("HTTP server ListenAndServe: %v", err)
	}
}

func setupTracing(cfg tracing.Config, hostPort int) {
	collector, err := zipkin.NewHTTPCollector(cfg.CollectorUrl)
	exitOnError(err, "Error while initializing zipkin")
	recorder := zipkin.NewRecorder(collector, cfg.Debug, strconv.Itoa(hostPort), cfg.ServiceSpanName)
	tracer, err := zipkin.NewTracer(recorder, zipkin.TraceID128Bit(false))
	exitOnError(err, "Error while initializing tracer")
	opentracing.SetGlobalTracer(tracer)
}
