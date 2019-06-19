package main

import (
	"encoding/base64"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/iam-kubeconfig-service/internal/authn"
	"github.com/kyma-project/kyma/components/iam-kubeconfig-service/pkg/kube_config"
	log "github.com/sirupsen/logrus"
)

const (
	oidcIssuerURLFlag = "oidc-issuer-url"
	oidcClientIDFlag  = "oidc-client-id"
	clusterNameFlag   = "kube-config-cluster-name"
	apiserverURLFlag  = "kube-config-url"
	clusterCAFileFlag = "kube-config-ca-file"
	oidcCAFileFlag = "oidc-ca-file"
)

func main() {

	cfg := readAppConfig()

	log.Infof("Starting IAM kubeconfig service on port: %d...", cfg.port)

	oidcAuthenticator, err := authn.NewOIDCAuthenticator(&cfg.oidc)
	if err != nil {
		log.Fatalf("Cannot create OIDC Authenticator, %v", err)
	}

	kubeConfig := kube_config.NewKubeConfig(cfg.clusterName, cfg.apiserverURL, cfg.clusterCA, cfg.namespace)

	kubeConfigEndpoints := kube_config.NewEndpoints(kubeConfig)

	router := mux.NewRouter()
	router.Use(authn.AuthMiddleware(oidcAuthenticator))
	router.Methods("GET").Path("/kube-config").HandlerFunc(kubeConfigEndpoints.GetKubeConfig)

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(cfg.port), router))
}

func readAppConfig() *appConfig {

	portArg := flag.Int("port", 8000, "Application port")
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

	if len(oidcSupportedSigningAlgsArg) == 0 {
		oidcSupportedSigningAlgsArg = []string{"RS256"}
	}

	clusterCAValue := readCAFromFile(*clusterCAFileArg)

	return &appConfig{
		port:         *portArg,
		clusterName:  *clusterNameArg,
		apiserverURL: *apiserverUrlArg,
		clusterCA:    clusterCAValue,
		namespace:    *namespaceArg,
		oidc: authn.OIDCConfig{
			IssuerURL:            *oidcIssuerURLArg,
			ClientID:             *oidcClientIDArg,
			CAFile:               *oidcCAFileArg,
			UsernameClaim:        *oidcUsernameClaimArg,
			UsernamePrefix:       *oidcUsernamePrefixArg,
			GroupsClaim:          *oidcGroupsClaimArg,
			GroupsPrefix:         *oidcGroupsPrefixArg,
			SupportedSigningAlgs: oidcSupportedSigningAlgsArg,
		},
	}
}

func readCAFromFile(caFile string) string {

	caBytes, caErr := ioutil.ReadFile(caFile)
	if caErr != nil {
		log.Fatalf("Error while reading Certificate Authority of the Kubernetes cluster. Root cause: %v", caErr)
	}

	return base64.StdEncoding.EncodeToString(caBytes)
}

type appConfig struct {
	port         int
	clusterName  string
	apiserverURL string
	clusterCA    string
	namespace    string
	oidc         authn.OIDCConfig
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
