package middlewares

import "github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"

type LookupService interface {
	Fetch(context clientcontext.ApplicationContext, configFilePath string) (string, error)
}

type GraphQLLookupService struct{}

func NewGraphQLLookupService() *GraphQLLookupService {
	return &GraphQLLookupService{}
}

func (ls *GraphQLLookupService) Fetch(context clientcontext.ApplicationContext, configFilePath string) (string, error) {
	return "", nil
}
