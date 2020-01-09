package testkit

import (
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/gql"
	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/idtokenprovider"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
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

func NewCompassDirectorClientOrDie(coreClient *kubernetes.Clientset, state CompassDirectorClientState, domain string) (*CompassDirectorClient, error) {
	idTokenConfig, err := getIDTokenProviderConfig(coreClient, state, domain)
	if err != nil {
		return nil, err
	}
	dexToken, err := idtokenprovider.Authenticate(idTokenConfig)
	if err != nil {
		return nil, err
	}
	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	return &CompassDirectorClient{
		gqlClient:   dexGraphQLClient,
		graphqlizer: gql.Graphqlizer{},
		fixtures:    helpers.NewCompassFixtures(),
		state:       state,
	}, nil
}

func (dc *CompassDirectorClient) RegisterApplication(in graphql.ApplicationRegisterInput) (graphql.ApplicationExt, error) {
	appInputGQL, err := dc.graphqlizer.ApplicationRegisterInputToGQL(in)
	if err != nil {
		return graphql.ApplicationExt{}, err
	}

	createRequest := dc.fixtures.FixRegisterApplicationRequest(appInputGQL)
	app := graphql.ApplicationExt{}
	err = dc.runOperation(createRequest, &app)
	if err != nil {
		return graphql.ApplicationExt{}, err
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

func (dc *CompassDirectorClient) AddScenarioToRuntime(runtimeID, scenarioName string) error {
	rtmReq := dc.fixtures.FixRuntimeRequest(runtimeID)
	rtm := graphql.RuntimeExt{}
	err := dc.runOperation(rtmReq, &rtm)
	if err != nil {
		return err
	}

	var scenarios []string
	switch v := rtm.Labels[dc.state.GetScenariosLabelKey()].(type) {
	case []interface{}:
		for _, label := range v {
			s, ok := label.(string)
			if !ok {
				return errors.New("invalid scenarios value")
			}
			if s == scenarioName {
				return nil
			}
			scenarios = append(scenarios, s)
		}
	case nil:
		scenarios = []string{}
	default:
		return errors.New("invalid scenarios value")
	}

	scenarios = append(scenarios, scenarioName)

	setLabelReq := dc.fixtures.FixSetRuntimeLabelRequest(runtimeID, dc.state.GetScenariosLabelKey(), scenarios)
	label := graphql.Label{}
	err = dc.runOperation(setLabelReq, &label)
	if err != nil {
		return err
	}

	return nil
}

func (dc *CompassDirectorClient) RemoveScenarioFromRuntime(runtimeID, scenarioName string) error {
	rtmReq := dc.fixtures.FixRuntimeRequest(runtimeID)
	rtm := graphql.RuntimeExt{}
	err := dc.runOperation(rtmReq, &rtm)
	if err != nil {
		return err
	}

	removed := false
	var scenarios []string
	switch v := rtm.Labels[dc.state.GetScenariosLabelKey()].(type) {
	case []interface{}:
		for _, label := range v {
			s, ok := label.(string)
			if !ok {
				return errors.New("invalid scenarios value")
			}
			if s == scenarioName {
				removed = true
				continue
			}
			scenarios = append(scenarios, s)
		}
	case nil:
		return nil
	default:
		return errors.New("invalid scenarios value")
	}

	if !removed {
		return nil
	}

	var req *gcli.Request
	if len(scenarios) == 0 {
		req = dc.fixtures.FixDeleteRuntimeLabelRequest(runtimeID, dc.state.GetScenariosLabelKey())
	} else {
		req = dc.fixtures.FixSetRuntimeLabelRequest(runtimeID, dc.state.GetScenariosLabelKey(), scenarios)
	}
	label := graphql.Label{}
	err = dc.runOperation(req, &label)
	if err != nil {
		return err
	}

	return nil
}

func (dc *CompassDirectorClient) GetOneTimeTokenForApplication(applicationID string) (graphql.OneTimeTokenExt, error) {
	req := dc.fixtures.FixRequestOneTimeTokenForApplication(applicationID)

	var oneTimeToken graphql.OneTimeTokenExt
	err := dc.runOperation(req, &oneTimeToken)
	if err != nil {
		return graphql.OneTimeTokenExt{}, err
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
