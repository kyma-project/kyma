package main

import (
	"context"
	"encoding/base64"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/iam-kubeconfig-service/cmd/generator/reload"
	"github.com/kyma-project/kyma/components/iam-kubeconfig-service/internal/authn"
	"github.com/kyma-project/kyma/components/iam-kubeconfig-service/pkg/kube_config"
	log "github.com/sirupsen/logrus"
	"k8s.io/apiserver/pkg/authentication/authenticator"
)

const (
	oidcIssuerURLFlag = "oidc-issuer-url"
	oidcClientIDFlag  = "oidc-client-id"
	clusterNameFlag   = "kube-config-cluster-name"
	apiserverURLFlag  = "kube-config-url"
	clusterCAFileFlag = "kube-config-ca-file"
	oidcCAFileFlag    = "oidc-ca-file"
	logLevelFlag      = "log-level"
)

func main() {

	cfg := readAppConfig()

	log.Info("Starting IAM kubeconfig service")

	fileWatcherCtx, fileWatcherCtxCancel := context.WithCancel(context.Background())

	oidcAuthenticator, err := setupOIDCAuthReloader(fileWatcherCtx, &cfg.oidc)

	if err != nil {
		log.Fatalf("Cannot create OIDC Authenticator, %v", err)
	}

	clusterCAReloader, err := setupClusterCAValueReloader(fileWatcherCtx, cfg.clusterCAFilePath)
	if err != nil {
		log.Fatalf("Cannot create reloadable cluster CA file provider, %v", err)
	}

	kubeConfig := kube_config.NewKubeConfig(cfg.clusterName, cfg.apiserverURL, clusterCAReloader.GetString, cfg.namespace)

	kubeConfigEndpoints := kube_config.NewEndpoints(kubeConfig)

	router := mux.NewRouter()
	router.Use(authn.AuthMiddleware(oidcAuthenticator))
	router.Methods("GET").Path("/kube-config").HandlerFunc(kubeConfigEndpoints.GetKubeConfig)

	healthRouter := mux.NewRouter()
	healthRouter.Methods("GET").Path("/health/ready").HandlerFunc(kubeConfigEndpoints.GetHealthStatus)

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(cfg.port), router)
		log.Errorf("Error serving HTTP: %v", err)
		term <- os.Interrupt
	}()

	log.Infof("IAM kubeconfig service started on port: %d...", cfg.port)

	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(cfg.healthPort), healthRouter)
		log.Errorf("Error serving HTTP: %v", err)
		term <- os.Interrupt
	}()

	log.Infof("Health endpoint started on port %d...", cfg.healthPort)

	select {
	case <-term:
		log.Info("Received SIGTERM, exiting gracefully...")
		fileWatcherCtxCancel()
	}

	//Allow for file watchers to close gracefully
	time.Sleep(1 * time.Second)
}

func readAppConfig() *appConfig {

	portArg := flag.Int("port", 8000, "Application port")
	healthPortArg := flag.Int("health-port", 9000, "Application health status port")
	clusterNameArg := flag.String(clusterNameFlag, "", "Name of the Kubernetes cluster")
	apiserverUrlArg := flag.String(apiserverURLFlag, "", "URL of the Kubernetes Apiserver")
	clusterCAFileArg := flag.String(clusterCAFileFlag, "", "File with Certificate Authority of the Kubernetes cluster, also used for OIDC authentication")
	namespaceArg := flag.String("kube-config-ns", "", "Default namespace of the Kubernetes context")

	oidcIssuerURLArg := flag.String(oidcIssuerURLFlag, "", "OIDC: The URL of the OpenID issuer. Used to verify the OIDC JSON Web Token (JWT)")
	oidcClientIDArg := flag.String(oidcClientIDFlag, "", "OIDC: The client ID for the OpenID Connect client")
	oidcUsernameClaimArg := flag.String("oidc-username-claim", "email", "OIDC: Identifier of the user in JWT claim")
	oidcGroupsClaimArg := flag.String("oidc-groups-claim", "groups", "OIDC: Identifier of groups in JWT claim")
	oidcUsernamePrefixArg := flag.String("oidc-username-prefix", "", "OIDC: If provided, all users will be prefixed with this value to prevent conflicts with other authentication strategies")
	oidcGroupsPrefixArg := flag.String("oidc-groups-prefix", "", "OIDC: If provided, all groups will be prefixed with this value to prevent conflicts with other authentication strategies")
	oidcCAFileArg := flag.String(oidcCAFileFlag, "", "File with Certificate Authority of the Kubernetes cluster, also used for OIDC authentication")
	logLevelArg := flag.String(logLevelFlag, "info", "Log level (trace, debug, info, warning, error, fatal and panic)")

	var oidcSupportedSigningAlgsArg multiValFlag = []string{}
	flag.Var(&oidcSupportedSigningAlgsArg, "oidc-supported-signing-algs", "OIDC supported signing algorithms")

	flag.Parse()

	errors := false

	if *clusterNameArg == "" {
		log.Errorf("Name of the Kubernetes cluster is required (-%s)", clusterNameFlag)
		errors = true
	}

	if *apiserverUrlArg == "" {
		log.Errorf("URL of the Kubernetes Apiserver is required (-%s)", apiserverURLFlag)
		errors = true
	}

	if *clusterCAFileArg == "" {
		log.Errorf("Cluster CA file path is required (-%s)", clusterCAFileFlag)
		errors = true
	}

	if *oidcIssuerURLArg == "" {
		log.Errorf("OIDC Issuer URL is required (-%s)", oidcIssuerURLFlag)
		errors = true
	}

	if *oidcClientIDArg == "" {
		log.Errorf("OIDC Client ID is required (-%s)", oidcClientIDFlag)
		errors = true
	}

	if errors {
		flag.Usage()
		os.Exit(1)
	}

	logLevel, err := log.ParseLevel(*logLevelArg)
	if err == nil {
		log.SetLevel(logLevel)
	} else {
		log.Error(err, ". Defaulting to \"info\"")
		log.SetLevel(log.InfoLevel)
	}

	if len(oidcSupportedSigningAlgsArg) == 0 {
		oidcSupportedSigningAlgsArg = []string{"RS256"}
	}

	return &appConfig{
		port:              *portArg,
		healthPort:        *healthPortArg,
		clusterName:       *clusterNameArg,
		apiserverURL:      *apiserverUrlArg,
		clusterCAFilePath: *clusterCAFileArg,
		namespace:         *namespaceArg,
		oidc: authn.OIDCConfig{
			IssuerURL:            *oidcIssuerURLArg,
			ClientID:             *oidcClientIDArg,
			CAFilePath:           *oidcCAFileArg,
			UsernameClaim:        *oidcUsernameClaimArg,
			UsernamePrefix:       *oidcUsernamePrefixArg,
			GroupsClaim:          *oidcGroupsClaimArg,
			GroupsPrefix:         *oidcGroupsPrefixArg,
			SupportedSigningAlgs: oidcSupportedSigningAlgsArg,
		},
	}
}

func readCAFromFile(caFile string) (string, error) {

	caBytes, caErr := ioutil.ReadFile(caFile)
	if caErr != nil {
		return "", caErr
	}

	return base64.StdEncoding.EncodeToString(caBytes), nil
}

type appConfig struct {
	port              int
	healthPort        int
	clusterName       string
	apiserverURL      string
	clusterCAFilePath string
	namespace         string
	oidc              authn.OIDCConfig
}

//Support for multi-valued flag: -flagName=val1 -flagName=val2 etc.
type multiValFlag []string

func (vals *multiValFlag) String() string {
	res := "["

	if len(*vals) > 0 {
		res = res + (*vals)[0]
	}

	for _, v := range *vals {
		res = res + ", " + v
	}
	res = res + "]"
	return res
}

func (vals *multiValFlag) Set(value string) error {
	*vals = append(*vals, value)
	return nil
}

func setupOIDCAuthReloader(fileWatcherCtx context.Context, cfg *authn.OIDCConfig) (authenticator.Request, error) {
	const eventBatchDelaySeconds = 10
	filesToWatch := []string{cfg.CAFilePath}

	cancelableAuthReqestConstructor := func() (authn.CancelableAuthRequest, error) {
		log.Infof("creating a new cancelable instance of authenticator.Request...")
		return authn.NewOIDCAuthenticator(cfg)
	}

	//Create reloader
	result, err := reload.NewCancelableAuthReqestReloader(cancelableAuthReqestConstructor)
	if err != nil {
		return nil, err
	}

	//Setup file watcher
	oidcCAFileWatcher := reload.NewWatcher("oidc-ca-dex-tls-cert", filesToWatch, eventBatchDelaySeconds, result.Reload)
	go oidcCAFileWatcher.Run(fileWatcherCtx)

	return result, nil
}

func setupClusterCAValueReloader(fileWatcherCtx context.Context, caFilePath string) (*reload.StringValueReloader, error) {
	const eventBatchDelaySeconds = 10
	filesToWatch := []string{caFilePath}

	caDataConstructorFunc := func() (string, error) {
		log.Infof("Reading Certificate Authority of the Kubernetes cluster from file: %s", caFilePath)
		caFileData, err := readCAFromFile(caFilePath)
		if err != nil {
			log.Errorf("Error while reading Certificate Authority of the Kubernetes cluster. Root cause: %v", err)
		}

		return caFileData, err
	}

	//Create reloader
	result, err := reload.NewStringValueReloader(caDataConstructorFunc)
	if err != nil {
		return nil, err
	}

	//Setup file watcher
	clusterCAFileWatcher := reload.NewWatcher("cluster-ca-crt", filesToWatch, eventBatchDelaySeconds, result.Reload)
	go clusterCAFileWatcher.Run(fileWatcherCtx)

	return result, nil
}
