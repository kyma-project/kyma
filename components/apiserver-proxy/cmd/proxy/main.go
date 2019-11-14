package main

import (
	"context"
	"crypto/tls"
	stdflag "flag"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/kyma-project/kyma/components/apiserver-proxy/cmd/proxy/reload"
	"github.com/kyma-project/kyma/components/apiserver-proxy/internal/spdy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/golang/glog"
	"github.com/hkwi/h2c"
	"github.com/kyma-project/kyma/components/apiserver-proxy/internal/authn"
	"github.com/kyma-project/kyma/components/apiserver-proxy/internal/authz"
	"github.com/kyma-project/kyma/components/apiserver-proxy/internal/proxy"
	flag "github.com/spf13/pflag"
	"golang.org/x/net/http2"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	certutil "k8s.io/client-go/util/cert"
	cliflag "k8s.io/component-base/cli/flag"
)

const (
	corsAllowOriginHeader      = "Access-Control-Allow-Origin"
	corsAllowMethodsHeader     = "Access-Control-Allow-Methods"
	corsAllowHeadersHeader     = "Access-Control-Allow-Headers"
	corsAllowCredentialsHeader = "Access-Control-Allow-Credentials"
)

var corsHeaders = []string{corsAllowOriginHeader, corsAllowMethodsHeader, corsAllowHeadersHeader, corsAllowCredentialsHeader}

type config struct {
	insecureListenAddress string
	secureListenAddress   string
	upstream              string
	upstreamForceH2C      bool
	auth                  proxy.Config
	tls                   tlsConfig
	kubeconfigLocation    string
	cors                  corsConfig
	metricsListenAddress  string
}

type tlsConfig struct {
	certFile     string
	keyFile      string
	minVersion   string
	cipherSuites []string
}

type corsConfig struct {
	allowHeaders []string
	allowOrigin  []string
	allowMethods []string
}

var versions = map[string]uint16{
	"VersionTLS10": tls.VersionTLS10,
	"VersionTLS11": tls.VersionTLS11,
	"VersionTLS12": tls.VersionTLS12,
}

func tlsVersion(versionName string) (uint16, error) {
	if version, ok := versions[versionName]; ok {
		return version, nil
	}
	return 0, fmt.Errorf("unknown tls version %q", versionName)
}

func main() {

	cfg := config{
		auth: proxy.Config{
			Authentication: &authn.AuthnConfig{
				X509:   &authn.X509Config{},
				Header: &authn.AuthnHeaderConfig{},
				OIDC:   &authn.OIDCConfig{},
			},
			Authorization: &authz.Config{},
		},
		cors: corsConfig{},
	}
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Add glog flags
	flagset.AddGoFlagSet(stdflag.CommandLine)

	// kube-rbac-proxy flags
	flagset.StringVar(&cfg.insecureListenAddress, "insecure-listen-address", "", "The address the kube-rbac-proxy HTTP server should listen on.")
	flagset.StringVar(&cfg.secureListenAddress, "secure-listen-address", "", "The address the kube-rbac-proxy HTTPs server should listen on.")
	flagset.StringVar(&cfg.upstream, "upstream", "", "The upstream URL to proxy to once requests have successfully been authenticated and authorized.")
	flagset.BoolVar(&cfg.upstreamForceH2C, "upstream-force-h2c", false, "Force h2c to communiate with the upstream. This is required when the upstream speaks h2c(http/2 cleartext - insecure variant of http/2) only. For example, go-grpc server in the insecure mode, such as helm's tiller w/o TLS, speaks h2c only")
	flagset.StringVar(&cfg.auth.Authorization.ResourceAttributesFile, "resource-attributes-file", "", "File spec of attributes-record to use for SubjectAccessReview. If unspecified, requests will attempted to be verified through non-resource-url attributes in the SubjectAccessReview.")

	// TLS flags
	flagset.StringVar(&cfg.tls.certFile, "tls-cert-file", "", "File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert)")
	flagset.StringVar(&cfg.tls.keyFile, "tls-private-key-file", "", "File containing the default x509 private key matching --tls-cert-file.")
	flagset.StringVar(&cfg.tls.minVersion, "tls-min-version", "VersionTLS12", "Minimum TLS version supported. Value must match version names from https://golang.org/pkg/crypto/tls/#pkg-constants.")
	flagset.StringSliceVar(&cfg.tls.cipherSuites, "tls-cipher-suites", nil, "Comma-separated list of cipher suites for the server. Values are from tls package constants (https://golang.org/pkg/crypto/tls/#pkg-constants). If omitted, the default Go cipher suites will be used")

	// Auth flags
	flagset.StringVar(&cfg.auth.Authentication.X509.ClientCAFile, "client-ca-file", "", "If set, any request presenting a client certificate signed by one of the authorities in the client-ca-file is authenticated with an identity corresponding to the CommonName of the client certificate.")
	flagset.BoolVar(&cfg.auth.Authentication.Header.Enabled, "auth-header-fields-enabled", false, "When set to true, kube-rbac-proxy adds auth-related fields to the headers of http requests sent to the upstream")
	flagset.StringVar(&cfg.auth.Authentication.Header.UserFieldName, "auth-header-user-field-name", "x-remote-user", "The name of the field inside a http(2) request header to tell the upstream server about the user's name")
	flagset.StringVar(&cfg.auth.Authentication.Header.GroupsFieldName, "auth-header-groups-field-name", "x-remote-groups", "The name of the field inside a http(2) request header to tell the upstream server about the user's groups")
	flagset.StringVar(&cfg.auth.Authentication.Header.GroupSeparator, "auth-header-groups-field-separator", "|", "The separator string used for concatenating multiple group names in a groups header field's value")

	//Authn OIDC flags
	flagset.StringVar(&cfg.auth.Authentication.OIDC.IssuerURL, "oidc-issuer", "", "The URL of the OpenID issuer, only HTTPS scheme will be accepted. If set, it will be used to verify the OIDC JSON Web Token (JWT).")
	flagset.StringVar(&cfg.auth.Authentication.OIDC.ClientID, "oidc-clientID", "", "The client ID for the OpenID Connect client, must be set if oidc-issuer-url is set.")
	flagset.StringVar(&cfg.auth.Authentication.OIDC.GroupsClaim, "oidc-groups-claim", "groups", "Identifier of groups in JWT claim, by default set to 'groups'")
	flagset.StringVar(&cfg.auth.Authentication.OIDC.UsernameClaim, "oidc-username-claim", "email", "Identifier of the user in JWT claim, by default set to 'email'")
	flagset.StringVar(&cfg.auth.Authentication.OIDC.GroupsPrefix, "oidc-groups-prefix", "", "If provided, all groups will be prefixed with this value to prevent conflicts with other authentication strategies.")
	flagset.StringArrayVar(&cfg.auth.Authentication.OIDC.SupportedSigningAlgs, "oidc-sign-alg", []string{"RS256"}, "Supported signing algorithms, default RS256")
	flagset.StringVar(&cfg.auth.Authentication.OIDC.CAFile, "oidc-ca-file", "", "If set, the OpenID server's certificate will be verified by one of the authorities in the oidc-ca-file, otherwise the host's root CA set will be used.")

	//Kubeconfig flag
	flagset.StringVar(&cfg.kubeconfigLocation, "kubeconfig", "", "Path to a kubeconfig file, specifying how to connect to the API server. If unset, in-cluster configuration will be used")

	// CORS flags
	flagset.StringSliceVar(&cfg.cors.allowOrigin, "cors-allow-origin", []string{"*"}, "List of CORS allowed origins")
	flagset.StringSliceVar(&cfg.cors.allowMethods, "cors-allow-methods", []string{"GET", "POST", "PUT", "DELETE"}, "List of CORS allowed methods")
	flagset.StringSliceVar(&cfg.cors.allowHeaders, "cors-allow-headers", []string{"Authorization", "Content-Type"}, "List of CORS allowed headers")

	// Prometheus
	flagset.StringVar(&cfg.metricsListenAddress, "metrics-listen-address", "", "The address the metric endpoint binds to.")

	flagset.Parse(os.Args[1:])
	kcfg := initKubeConfig(cfg.kubeconfigLocation)

	upstreamURL, err := url.Parse(cfg.upstream)
	if err != nil {
		glog.Fatalf("Failed to build parse upstream URL: %v", err)
	}

	spdyProxy := spdy.New(kcfg, upstreamURL)

	kubeClient, err := kubernetes.NewForConfig(kcfg)
	if err != nil {
		glog.Fatalf("Failed to instantiate Kubernetes client: %v", err)
	}

	var oidcAuthenticator authenticator.Request

	fileWatcherCtx, fileWatcherCtxCancel := context.WithCancel(context.Background())

	// If OIDC configuration provided, use oidc authenticator
	if cfg.auth.Authentication.OIDC.IssuerURL != "" {

		oidcAuthenticator, err = setupOIDCAuthReloader(fileWatcherCtx, cfg.auth.Authentication.OIDC)
		if err != nil {
			glog.Fatalf("Failed to instantiate OIDC authenticator: %v", err)
		}
	} else {
		//Use Delegating authenticator
		tokenClient := kubeClient.AuthenticationV1beta1().TokenReviews()
		oidcAuthenticator, err = authn.NewDelegatingAuthenticator(tokenClient, cfg.auth.Authentication)
		if err != nil {
			glog.Fatalf("Failed to instantiate delegating authenticator: %v", err)
		}
	}

	sarClient := kubeClient.AuthorizationV1beta1().SubjectAccessReviews()
	authorizer, err := authz.NewAuthorizer(sarClient)

	if err != nil {
		glog.Fatalf("Failed to create authorizer: %v", err)
	}

	authProxy := proxy.New(cfg.auth, authorizer, oidcAuthenticator)

	if err != nil {
		glog.Fatalf("Failed to create rbac-proxy: %v", err)
	}

	proxyForApiserver := strings.Contains(cfg.upstream, proxy.KUBERNETES_SERVICE)

	rp, err := newReverseProxy(upstreamURL, kcfg, proxyForApiserver)
	if err != nil {
		glog.Fatalf("Unable to create reverse proxy, %s", err)
	}

	//Prometheus
	prometheusRegistry := prometheus.NewRegistry()
	err = prometheusRegistry.Register(prometheus.NewGoCollector())
	if err != nil {
		glog.Fatalf("failed to register Go runtime metrics: %v", err)
	}

	err = prometheusRegistry.Register(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	if err != nil {
		glog.Fatalf("failed to register process metrics: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ok := authProxy.Handle(w, req)
		if !ok {
			return
		}

		if spdyProxy.IsSpdyRequest(req) {
			spdyProxy.ServeHTTP(w, req)
		} else {
			rp.ServeHTTP(w, req)
		}
	}))

	if cfg.secureListenAddress != "" {
		srv := &http.Server{Handler: getCORSHandler(mux, cfg.cors)}

		if cfg.tls.certFile == "" && cfg.tls.keyFile == "" {
			glog.Info("Generating self signed cert as no cert is provided")
			certBytes, keyBytes, err := certutil.GenerateSelfSignedCertKey("", nil, nil)
			if err != nil {
				glog.Fatalf("Failed to generate self signed cert and key: %v", err)
			}
			cert, err := tls.X509KeyPair(certBytes, keyBytes)
			if err != nil {
				glog.Fatalf("Failed to load generated self signed cert and key: %v", err)
			}

			version, err := tlsVersion(cfg.tls.minVersion)
			if err != nil {
				glog.Fatalf("TLS version invalid: %v", err)
			}

			cipherSuiteIDs, err := cliflag.TLSCipherSuites(cfg.tls.cipherSuites)
			if err != nil {
				glog.Fatalf("Failed to convert TLS cipher suite name to ID: %v", err)
			}
			srv.TLSConfig = &tls.Config{
				CipherSuites: cipherSuiteIDs,
				Certificates: []tls.Certificate{cert},
				MinVersion:   version,
				// To enable http/2
				// See net/http.Server.shouldConfigureHTTP2ForServe for more context
				NextProtos: []string{"h2"},
			}
		} else {
			certReloader, err := setupTLSCertReloader(fileWatcherCtx, cfg.tls.certFile, cfg.tls.keyFile)
			if err != nil {
				glog.Fatalf("Failed to create ReloadableTLSCertProvider: %v", err)
			}

			//Configure srv with GetCertificate function
			srv.TLSConfig = &tls.Config{
				GetCertificate: certReloader.GetCertificateFunc,
			}
		}

		l, err := net.Listen("tcp", cfg.secureListenAddress)
		if err != nil {
			glog.Fatalf("Failed to listen on secure address: %v", err)
		}
		glog.Infof("Listening securely on %v", cfg.secureListenAddress)

		go srv.ServeTLS(l, "", "")
	}

	if cfg.metricsListenAddress != "" {
		srv := &http.Server{Handler: promhttp.Handler()}

		l, err := net.Listen("tcp", cfg.metricsListenAddress)
		if err != nil {
			glog.Fatalf("Failed to listen on insecure address: %v", err)
		}
		glog.Infof("Listening for metrics on %v", cfg.metricsListenAddress)
		go srv.Serve(l)
	}

	if cfg.insecureListenAddress != "" {
		if cfg.upstreamForceH2C && !proxyForApiserver {
			// Force http/2 for connections to the upstream i.e. do not start with HTTP1.1 UPGRADE req to
			// initialize http/2 session.
			// See https://github.com/golang/go/issues/14141#issuecomment-219212895 for more context
			rp.Transport = &http2.Transport{
				// Allow http schema. This doesn't automatically disable TLS
				AllowHTTP: true,
				// Do disable TLS.
				// In combination with the schema check above. We could enforce h2c against the upstream server
				DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial(netw, addr)
				},
			}
		}

		// Background:
		//
		// golang's http2 server doesn't support h2c
		// https://github.com/golang/go/issues/16696
		//
		//
		// Action:
		//
		// Use hkwi/h2c so that you can properly handle HTTP Upgrade requests over plain TCP,
		// which is one of consequences for a h2c support.
		//
		// See https://github.com/golang/go/issues/14141 for more context.
		//
		// Possible alternative:
		//
		// We could potentially use grpc-go server's HTTP handler support
		// which would handle HTTP UPGRADE from http1.1 to http/2, especially in case
		// what you wanted kube-rbac-proxy to authn/authz was gRPC over h2c calls.
		//
		// Note that golang's http server requires a client(including gRPC) to send HTTP Upgrade req to
		// property start http/2.
		//
		// but it isn't straight-forward to understand.
		// Also note that at time of writing this, grpc-go's server implementation still lacks
		// a h2c support for communication against the upstream.
		//
		// See belows for more information:
		// - https://github.com/grpc/grpc-go/pull/1406/files
		// - https://github.com/grpc/grpc-go/issues/549#issuecomment-191458335
		// - https://github.com/golang/go/issues/14141#issuecomment-176465220
		h2cHandler := &h2c.Server{Handler: mux}

		srv := &http.Server{Handler: h2cHandler}

		l, err := net.Listen("tcp", cfg.insecureListenAddress)
		if err != nil {
			glog.Fatalf("Failed to listen on insecure address: %v", err)
		}
		glog.Infof("Listening insecurely on %v", cfg.insecureListenAddress)
		go srv.Serve(l)
	}

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	select {
	case <-term:
		glog.Info("Received SIGTERM, exiting gracefully...")
		fileWatcherCtxCancel()
	}

	//Allow for file watchers to close gracefully
	time.Sleep(1 * time.Second)
}

// Returns intiliazed config, allows local usage (outside cluster) based on provided kubeconfig or in-cluter
func initKubeConfig(kcLocation string) *rest.Config {

	if kcLocation != "" {
		kubeConfig, err := clientcmd.BuildConfigFromFlags("", kcLocation)
		if err != nil {
			glog.Fatalf("unable to build rest config based on provided path to kubeconfig file %s", err.Error())
		}
		return kubeConfig
	}

	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatal("cannot find Service Account in pod to build in-cluster rest config")
	}

	return kubeConfig
}

func newReverseProxy(target *url.URL, kcfg *rest.Config, proxyForApiserver bool) (*httputil.ReverseProxy, error) {
	rp := httputil.NewSingleHostReverseProxy(target)
	rp.ModifyResponse = deleteUpstreamCORSHeaders

	if proxyForApiserver {
		t, err := rest.TransportFor(kcfg)
		if err != nil {
			return nil, fmt.Errorf("unable to set HTTP Transport for the upstream. Details : %s", err.Error())
		}
		rp.Transport = t
	}

	return rp, nil
}

func getCORSHandler(handler http.Handler, corsCfg corsConfig) http.Handler {
	return handlers.CORS(
		handlers.AllowedOrigins(corsCfg.allowOrigin),
		handlers.AllowedMethods(corsCfg.allowMethods),
		handlers.AllowedHeaders(corsCfg.allowHeaders),
	)(handler)
}

func deleteUpstreamCORSHeaders(r *http.Response) error {
	for _, h := range corsHeaders {
		r.Header.Del(h)
	}
	return nil
}

func setupOIDCAuthReloader(fileWatcherCtx context.Context, cfg *authn.OIDCConfig) (authenticator.Request, error) {
	const eventBatchDelaySeconds = 10
	filesToWatch := []string{cfg.CAFile}

	cancelableAuthReqestConstructor := func() (authn.CancelableAuthRequest, error) {
		glog.Infof("creating new cancelable instance of authenticator.Request...")
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

func setupTLSCertReloader(fileWatcherCtx context.Context, certFile, keyFile string) (*reload.TLSCertReloader, error) {
	const eventBatchDelaySeconds = 10

	tlsConstructor := func() (*tls.Certificate, error) {
		glog.Infof("Creating new TLS Certificate from data files: %s, %s", certFile, keyFile)
		res, err := tls.LoadX509KeyPair(certFile, keyFile)
		return &res, err
	}
	//Create reloader
	result, err := reload.NewTLSCertReloader(tlsConstructor)
	if err != nil {
		return nil, err
	}

	//Start file watcher for certificate files
	tlsCertFileWatcher := reload.NewWatcher("main-tls-crt/key", []string{certFile, keyFile}, eventBatchDelaySeconds, result.Reload)
	go tlsCertFileWatcher.Run(fileWatcherCtx)

	return result, nil
}
