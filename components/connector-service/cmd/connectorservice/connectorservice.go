package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/errorhandler"
	"github.com/kyma-project/kyma/components/connector-service/internal/externalapi"
	"github.com/kyma-project/kyma/components/connector-service/internal/internalapi"
	"github.com/kyma-project/kyma/components/connector-service/internal/secrets"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
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

	tokenCache := tokencache.NewTokenCache(options.tokenExpirationMinutes)
	certUtil := certificates.NewCertificateUtility()
	tokenGenerator := tokens.NewTokenGenerator(options.tokenLength, tokenCache)

	externalHandler := newExternalHandler(tokenCache, certUtil, tokenGenerator, options, env)
	internalHandler := newInternalHandler(tokenCache, tokenGenerator, options.connectorServiceHost)

	externalSrv := &http.Server{
		Addr:    ":" + strconv.Itoa(options.externalAPIPort),
		Handler: externalHandler,
	}

	internalSrv := &http.Server{
		Addr:    ":" + strconv.Itoa(options.internalAPIPort),
		Handler: internalHandler,
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		log.Info(externalSrv.ListenAndServe())
	}()

	go func() {
		log.Info(internalSrv.ListenAndServe())
	}()

	wg.Wait()
}

func newExternalHandler(cache tokencache.TokenCache, utility certificates.CertificateUtility, tokenGenerator tokens.TokenGenerator, opts *options, env *environment) http.Handler {
	secretsRepository, appErr := newSecretsRepository(opts.namespace)
	if appErr != nil {
		log.Infof("Failed to create secrets repository. %s", appErr.Error())
		return errorhandler.NewErrorHandler(500, fmt.Sprintf("Failed to create secrets repository: %s", appErr.Error()))
	}

	subjectValues := certificates.CSRSubject{
		Country:            env.country,
		Organization:       env.organization,
		OrganizationalUnit: env.organizationalUnit,
		Locality:           env.locality,
		Province:           env.province,
	}
	rh := externalapi.NewSignatureHandler(cache, utility, secretsRepository, opts.connectorServiceHost, opts.domainName, subjectValues)
	ih := externalapi.NewInfoHandler(cache, tokenGenerator, opts.connectorServiceHost, opts.domainName, subjectValues)
	return externalapi.NewHandler(rh, ih)
}

func newInternalHandler(cache tokencache.TokenCache, tokenGenerator tokens.TokenGenerator, host string) http.Handler {
	th := internalapi.NewTokenHandler(tokenGenerator, host)
	return internalapi.NewHandler(th)
}

func newSecretsRepository(namespace string) (secrets.Repository, apperrors.AppError) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, apperrors.Internal("failed to read k8s in-cluster configuration, %s", err)
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("failed to create k8s core client, %s", err)
	}

	sei := coreClientset.CoreV1().Secrets(namespace)

	return secrets.NewRepository(sei), nil
}
