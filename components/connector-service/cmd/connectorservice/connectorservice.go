package main

import (
	"net/http"
	"strconv"
	"sync"

	connectoruuid "github.com/kyma-project/kyma/components/connector-service/internal/uuid"
	uuid "github.com/satori/go.uuid"

	"github.com/kyma-project/kyma/components/connector-service/internal/applications"
	"github.com/kyma-project/kyma/components/connector-service/internal/kymagroup"
	kymagroups "github.com/kyma-project/kyma/components/connector-service/pkg/client/clientset/versioned"

	"github.com/gorilla/mux"
	apps "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/connector-service/internal/api/application/externalapi"
	"github.com/kyma-project/kyma/components/connector-service/internal/api/application/internalapi"
	kymaextapi "github.com/kyma-project/kyma/components/connector-service/internal/api/kymacluster/externalapi"
	kymaintapi "github.com/kyma-project/kyma/components/connector-service/internal/api/kymacluster/internalapi"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/monitoring"
	"github.com/kyma-project/kyma/components/connector-service/internal/secrets"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
	"github.com/kyma-project/kyma/components/connector-service/internal/verification"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const (
	caSecretName = "nginx-auth-ca" // TODO - as flag?
)

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	log.Info("Starting Certificate Service.")

	options := parseArgs()
	log.Infof("Options: %s", options)
	env := parseEnv()
	log.Infof("Environment variables: %s", env)

	servers, err := setup(options, env)
	if err != nil {
		log.Errorf("Failed to setup components, %s", err.Error())
		return
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)

	for _, server := range servers {
		go func(srv *http.Server) {
			log.Info(srv.ListenAndServe())
		}(server)
	}

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Info(http.ListenAndServe(":9090", nil))
	}()

	wg.Wait()
}

func setup(options *options, env *environment) ([]*http.Server, error) {
	subjectValues := csrSubject(env)

	middlewares, appErr := monitoring.SetupMonitoringMiddleware()
	if appErr != nil {
		log.Errorf("Error while setting up monitoring: %s", appErr)
	}

	tokenCache := tokens.NewTokenCache(options.tokenExpirationMinutes)
	tokenService := tokens.NewTokenService(options.tokenLength, tokenCache)

	secretsRepository, kymaGroupRepository, appRepository, appErr := newK8sClients(options.namespace)
	if appErr != nil {
		log.Errorf("Failed to initialize secrets repository. %s", appErr.Error())
		return nil, appErr
	}

	verificationService := verification.NewVerificationService(options.global)

	certUtil := certificates.NewCertificateUtility()
	certificateService := certificates.NewCertificateService(secretsRepository, certUtil, caSecretName, subjectValues)

	externalHandler := newExternalHandler(certificateService, tokenService, options, middlewares, subjectValues, kymaGroupRepository, appRepository)
	internalHandler := newInternalHandler(tokenService, options.connectorServiceHost, middlewares, verificationService)

	uuidGenerator := connectoruuid.GeneratorFunc(func() string {
		return uuid.NewV4().String()
	})

	externalSrv := &http.Server{
		Addr:    ":" + strconv.Itoa(options.externalAPIPort),
		Handler: externalHandler,
	}

	internalSrv := &http.Server{
		Addr:    ":" + strconv.Itoa(options.internalAPIPort),
		Handler: internalHandler,
	}

	servers := []*http.Server{
		internalSrv, externalSrv,
	}

	if options.global {
		kymaClustersAPIs := newKymaClusterAPIs(tokenService, certificateService, subjectValues, options, middlewares, kymaGroupRepository, uuidGenerator)
		servers = append(servers, kymaClustersAPIs...)
	}

	return servers, nil
}

func csrSubject(env *environment) certificates.CSRSubject {
	return certificates.CSRSubject{
		Country:            env.country,
		Organization:       env.organization,
		OrganizationalUnit: env.organizationalUnit,
		Locality:           env.locality,
		Province:           env.province,
	}
}

func newExternalHandler(certService certificates.Service, tokenService tokens.Service, opts *options, middlewares []mux.MiddlewareFunc, subjectValues certificates.CSRSubject, groupsRepository kymagroup.Repository, appsRepository applications.Repository) http.Handler {
	rh := externalapi.NewSignatureHandler(tokenService, certService, opts.connectorServiceHost, opts.domainName, groupsRepository, appsRepository)
	ih := externalapi.NewInfoHandler(tokenService, opts.connectorServiceHost, opts.domainName, subjectValues, groupsRepository)
	return externalapi.NewHandler(rh, ih, middlewares)
}

func newInternalHandler(tokenGenerator tokens.Service, host string, middlewares []mux.MiddlewareFunc, verificationSvc verification.Service) http.Handler {
	th := internalapi.NewTokenHandler(verificationSvc, tokenGenerator, host)
	return internalapi.NewHandler(th, middlewares)
}

func newK8sClients(namespace string) (secrets.Repository, kymagroup.Repository, applications.Repository, apperrors.AppError) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, nil, nil, apperrors.Internal("failed to read k8s in-cluster configuration, %s", err)
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, nil, nil, apperrors.Internal("failed to create k8s core client, %s", err)
	}
	sei := coreClientset.CoreV1().Secrets(namespace)

	applicationClientset, err := apps.NewForConfig(k8sConfig)
	if err != nil {
		return nil, nil, nil, apperrors.Internal("failed to create application client, %s", err)
	}
	appsClient := applicationClientset.ApplicationconnectorV1alpha1().Applications()

	kymaGroupCientset, err := kymagroups.NewForConfig(k8sConfig)
	if err != nil {
		return nil, nil, nil, apperrors.Internal("failed to create k8s core client, %s", err)
	}
	groupsClient := kymaGroupCientset.ApplicationconnectorV1alpha1().KymaGroups()

	return secrets.NewRepository(sei), kymagroup.NewKymaGroupRepository(groupsClient), applications.NewApplicationRepository(appsClient), nil
}

func newKymaClusterAPIs(tokenService tokens.Service, certificateService certificates.Service, subjectValues certificates.CSRSubject, options *options, middlewares []mux.MiddlewareFunc, groupRepository kymagroup.Repository, generator connectoruuid.Generator) []*http.Server {
	clusterTokenHandler := kymaintapi.NewTokenHandler(tokenService, options.connectorServiceHost, generator)

	internalSrv := &http.Server{
		Addr:    ":" + strconv.Itoa(options.internalKymaAPIPort),
		Handler: kymaintapi.NewHandler(clusterTokenHandler, middlewares),
	}

	clusterInfoHandler := kymaextapi.NewInfoHandler(tokenService, options.connectorServiceHost, subjectValues)
	clusterSignHandler := kymaextapi.NewSignatureHandler(tokenService, certificateService, options.connectorServiceHost, groupRepository)

	externalSrv := &http.Server{
		Addr:    ":" + strconv.Itoa(options.externalKymaAPIPort),
		Handler: kymaextapi.NewHandler(clusterSignHandler, clusterInfoHandler, middlewares),
	}

	return []*http.Server{
		internalSrv, externalSrv,
	}
}
