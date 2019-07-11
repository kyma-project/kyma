package lookup

import (
	"fmt"
	"github.com/kyma-project/kyma/components/connector-service/internal/graphql"
	"io/ioutil"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/tidwall/gjson"
)

const (
	filename = "config.json"
	timeout  = 30 * time.Second
	query    = `{"query":"{ applications(where: {accountId: \"%s\", groupName: \"%s\", appName: \"%s\"}) {name account { id } groups { id name clusters { id name endpoints { gateway } } } } }"}`
)

type LookupService interface {
	Fetch(context clientcontext.ApplicationContext) (string, error)
}

type lookupService struct {
	graphQLService graphql.GraphQLService
	configFilePath string
}

func NewGraphQLLookupService(graphQLService graphql.GraphQLService, configFilePath string) LookupService {
	return &lookupService{graphQLService: graphQLService, configFilePath: configFilePath}
}

func (ls lookupService) Fetch(context clientcontext.ApplicationContext) (string, error) {
	file, e := os.Open(ls.configFilePath + filename)

	if e != nil {
		return "", apperrors.Internal("Error while reading config file: %s", e)
	}

	lookUpConfig, e := ls.graphQLService.ReadConfig(file)

	if e != nil {
		return "", e
	}

	query := createQuery(context)

	response, err := ls.graphQLService.SendRequest(query, lookUpConfig, timeout)

	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	body, e := ioutil.ReadAll(response.Body)

	if e != nil {
		return "", apperrors.Internal("Error reading response body: %s", e)
	}

	return getGatewayUrl(body), nil
}

func createQuery(context clientcontext.ApplicationContext) string {
	return fmt.Sprintf(query, context.Tenant, context.Group, context.Application)
}

func getGatewayUrl(body []byte) string {
	stringBody := string(body)
	logrus.Info(stringBody)
	gatewayUrl := gjson.Get(stringBody, "data.applications.0.groups.0.clusters.0.endpoints.gateway")
	return gatewayUrl.Str
}
