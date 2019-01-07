package main

import (
	"context"
	"flag"
	"fmt"
	"k8s.io/apiserver/pkg/authentication/user"
	authorizerpkg "k8s.io/apiserver/pkg/authorization/authorizer"
	"net/http"
	"time"

	"k8s.io/apiserver/pkg/authentication/authenticator"

	"github.com/gorilla/mux"
	"k8s.io/client-go/kubernetes"

	"github.com/99designs/gqlgen/graphql"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/signal"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/vrischmann/envconfig"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/websocket"
	//"github.com/kyma-project/kyma/components/ui-api-layer/internal/authn"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/authz"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/origin"
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
}

func main() {
	cfg, err := loadConfig("APP")
	exitOnError(err, "Error while loading app config")
	parseFlags(cfg)

	k8sConfig, err := newRestClientConfig(cfg.KubeconfigPath)
	exitOnError(err, "Error while initializing REST client config")

	resolvers, err := domain.New(k8sConfig, cfg.Content, cfg.Application, cfg.InformerResyncPeriod)
	exitOnError(err, "Error while creating resolvers")

	kubeClient, err := kubernetes.NewForConfig(k8sConfig)
	exitOnError(err, "Failed to instantiate Kubernetes client")

	config := authn.OIDCConfig{} // TODO: prepare config
	authenticator, err := authn.NewOIDCAuthenticator(&config)
	exitOnError(err, "Error while creating OIDC authenticator")

	sarClient := kubeClient.AuthorizationV1beta1().SubjectAccessReviews()
	authorizer, err := authz.NewAuthorizer(sarClient)
	exitOnError(err, "Failed to create authorizer")

	stopCh := signal.SetupChannel()
	resolvers.WaitForCacheSync(stopCh)

	c := gqlschema.Config{Resolvers: resolvers}
	c.Directives.CheckRBAC = func(ctx context.Context, obj interface{}, next graphql.Resolver, attributes gqlschema.RBACAttributes) (res interface{}, err error) {

		// fetch user from context
		u := UserInfoForContext(ctx)

		// prepare attributes for authz
		attrs := authz.PrepareAttributes(ctx, u, attributes)
		glog.Infof("%+v", attrs)

		// check if user is allowed to get requested resource
		authorized, reason, err := authorizer.Authorize(attrs)
		// TODO: handle errors

		glog.Infof("authorized: %v, reason: %s, err: %v", authorized, reason, err)

		if authorized != authorizerpkg.DecisionAllow || err != nil {
			return nil, fmt.Errorf("access denied")
		}

		// success path TODO: delete this comment and logging attributes below
		glog.Infof("atrributes: %+v", attributes)
		glog.Infof("obj: %+v", obj)
		glog.Infof("ctx: %+v", ctx)

		return next(ctx)
	}
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
		_ = flag.Set("stderrthreshold", "INFO")
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

func runServer(stop <-chan struct{}, addr string, allowedOrigins []string, schema graphql.ExecutableSchema, authenticator authenticator.Request) {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	mux := mux.NewRouter()

	mux.Use(authn.AuthMiddleware(authenticator))

	mux.HandleFunc("/", handler.Playground("Dataloader", "/graphql"))
	mux.HandleFunc("/graphql", handler.GraphQL(schema,
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
	}).Handler(mux)

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
