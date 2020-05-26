package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"time"

	graphqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/kyma-project/kyma/components/console-backend-service3/internal/authn"
	"github.com/kyma-project/kyma/components/console-backend-service3/internal/authz"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/generated"
	authenticatorpkg "k8s.io/apiserver/pkg/authentication/authenticator"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type config struct {
	Host                 string        `envconfig:"default=127.0.0.1"`
	Port                 int           `envconfig:"default=3000"`
	AllowedOrigins       []string      `envconfig:"optional"`
	Verbose              bool          `envconfig:"default=false"`
	KubeconfigPath       string        `envconfig:"optional"`
	SystemNamespaces     []string      `envconfig:"default=istio-system;knative-eventing;knative-serving;kube-public;kube-system;kyma-installer;kyma-integration;kyma-system;natss;compass-system"`
	InformerResyncPeriod time.Duration `envconfig:"default=10m"`
	ServerTimeout        time.Duration `envconfig:"default=10s"`
	Burst                int           `envconfig:"default=2"`
	OIDC                 authn.OIDCConfig
	SARCacheConfig       authz.SARCacheConfig
}

func main() {

	cfg, _, err := loadConfig("APP")
	if err != nil {
		panic(err)
	}
	parseFlags(cfg)

	restClientConfig, err := newRestClientConfig(cfg.KubeconfigPath, 2)
	if err != nil {
		panic(err)
	}

	resolver, err := graph.NewResolver(restClientConfig, 10*time.Minute)
	if err != nil {
		panic(err)
	}
	stopCh := make(chan struct{})
	resolver.WaitForCacheSync(stopCh)


	kubeClient, err := kubernetes.NewForConfig(restClientConfig)
	exitOnError(err, "Failed to instantiate Kubernetes client")

	srvConfig := generated.Config{Resolvers: resolver}


	var authenticator authenticatorpkg.Request

	authenticator, err = authn.NewOIDCAuthenticator(&cfg.OIDC)
	exitOnError(err, "Error while creating OIDC authenticator")
	sarClient := kubeClient.AuthorizationV1().SubjectAccessReviews()
	authorizer, err := authz.NewAuthorizer(sarClient, cfg.SARCacheConfig)
	exitOnError(err, "Failed to create authorizer")

	srvConfig.Directives.HasAccess = authz.NewRBACDirective(authorizer, kubeClient.Discovery())

	runServer(stopCh, cfg, srvConfig, authenticator)
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

func loadConfig(prefix string) (config, bool, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	if err != nil {
		return cfg, false, err
	}

	developmentMode := cfg.KubeconfigPath != ""

	return cfg, developmentMode, nil
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

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		glog.Fatal(wrappedError)
	}
}

func runServer(stop <-chan struct{}, cfg config, srvConfig generated.Config, authenticator authenticatorpkg.Request) {
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

	graphqlserver := graphqlhandler.NewDefaultServer(generated.NewExecutableSchema(srvConfig))

	router.HandleFunc("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", graphqlserver)

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

	glog.Infof("connect to http://%s/ for GraphQL playground", addr)

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
