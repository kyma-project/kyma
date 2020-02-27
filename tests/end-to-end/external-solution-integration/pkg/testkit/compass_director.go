package testkit

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	gcli "github.com/machinebox/graphql"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

const timeout = time.Second * 10

type CompassDirectorClient struct {
	gqlClient   *gcli.Client
	graphqlizer gql.Graphqlizer
	fixtures    helpers.CompassFixtures
	state       CompassDirectorClientState
}

type CompassDirectorClientState interface {
	GetScenariosLabelKey() string
	GetDefaultTenant() string
	GetRuntimeID() string
	GetDexSecret() (string, string)
}

func NewCompassDirectorClientOrDie(coreClient *kubernetes.Clientset, state CompassDirectorClientState, domain string) *CompassDirectorClient {
	idTokenConfig, err := getIDTokenProviderConfig(coreClient, state, domain)
	if err != nil {
		panic(err)
	}
	dexToken, err := idtokenprovider.Authenticate(idTokenConfig)
	if err != nil {
		panic(err)
	}
	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	return &CompassDirectorClient{
		gqlClient:   dexGraphQLClient,
		graphqlizer: gql.Graphqlizer{},
		fixtures:    helpers.NewCompassFixtures(),
		state:       state,
	}
}

func (dc *CompassDirectorClient) RegisterApplication(in graphql.ApplicationRegisterInput) (graphql.ApplicationExt, error) {
	appInputGQL, err := dc.graphqlizer.ApplicationRegisterInputToGQL(in)
	if err != nil {
		return graphql.ApplicationExt{}, err
	}

	createRequest := dc.fixtures.FixRegisterApplicationRequest(appInputGQL)
	app := graphql.ApplicationExt{}
	err = dc.runOperation(createRequest, &app)
	return app, err
}

func (dc *CompassDirectorClient) GetApplication(appID string) (graphql.ApplicationExt, error) {
	request := dc.fixtures.FixGetApplication(appID)
	app := graphql.ApplicationExt{}
	err := dc.runOperation(request, &app)
	if err != nil {
		return graphql.ApplicationExt{}, errors.Wrapf(err, "while getting application with id: %s", appID)
	}
	return app, nil
}

func (dc *CompassDirectorClient) UnregisterApplication(id string) (graphql.ApplicationExt, error) {
	req := dc.fixtures.FixUnregisterApplicationRequest(id)
	app := graphql.ApplicationExt{}
	err := dc.runOperation(req, &app)
	if err != nil {
		return graphql.ApplicationExt{}, err
	}

	return app, nil
}

func (dc *CompassDirectorClient) GetOneTimeTokenForApplication(applicationID string) (graphql.OneTimeTokenForApplicationExt, error) {
	req := dc.fixtures.FixRequestOneTimeTokenForApplication(applicationID)

	var oneTimeToken graphql.OneTimeTokenForApplicationExt
	err := dc.runOperation(req, &oneTimeToken)
	if err != nil {
		return graphql.OneTimeTokenForApplicationExt{}, err
	}

	return oneTimeToken, nil
}

func getIDTokenProviderConfig(coreClient *kubernetes.Clientset, state CompassDirectorClientState, domain string) (idtokenprovider.Config, error) {
	secretName, secretNamespace := state.GetDexSecret()
	secretInterface := coreClient.CoreV1().Secrets(secretNamespace)
	secretsRepository := helpers.NewSecretRepository(secretInterface)
	dexSecret, err := secretsRepository.Get(secretName)
	if err != nil {
		return idtokenprovider.Config{}, err
	}
	return idtokenprovider.NewConfig(dexSecret.UserEmail, dexSecret.UserPassword, domain, timeout)
}

func (dc *CompassDirectorClient) runOperation(req *gcli.Request, resp interface{}) error {
	m := dc.resultMapperFor(&resp)

	req.Header.Set("Tenant", dc.state.GetDefaultTenant())

	return dc.withRetryOnTemporaryConnectionProblems(func() error { return dc.gqlClient.Run(context.Background(), req, &m) })
}

func (dc *CompassDirectorClient) withRetryOnTemporaryConnectionProblems(risky func() error) error {
	return retry.Do(risky, retry.Attempts(7), retry.Delay(time.Second), retry.OnRetry(func(n uint, err error) {
		logrus.WithField("component", "TestContext").Warnf("OnRetry: attempts: %d, error: %v", n, err)

	}), retry.LastErrorOnly(true), retry.RetryIf(func(err error) bool {
		return strings.Contains(err.Error(), "connection refused") ||
			strings.Contains(err.Error(), "connection reset by peer")
	}))
}

// resultMapperFor returns generic object that can be passed to Run method for storing response.
// In GraphQL, set `result` alias for your query
func (dc *CompassDirectorClient) resultMapperFor(target interface{}) genericGQLResponse {
	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		panic("target has to be a pointer")
	}
	return genericGQLResponse{
		Result: target,
	}
}

type genericGQLResponse struct {
	Result interface{} `json:"result"`
}
